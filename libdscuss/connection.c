/**
 * This file is part of Dscuss.
 * Copyright (C) 2014-2015  Vitaly Minko
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * Additional permission under GNU GPL version 3 section 7
 *
 * If you modify this program, or any covered work, by linking or
 * combining it with the OpenSSL project's OpenSSL library (or a
 * modified version of that library), containing parts covered by the
 * terms of the OpenSSL or SSLeay licenses, the copyright holders
 * grants you additional permission to convey the resulting work.
 * Corresponding Source for a non-source form of such a combination
 * shall include the source code for the parts of OpenSSL used as well
 * as that of the covered work.
 */

#include <glib.h>
#include <gio/gio.h>
#include "config.h"
#include "packet.h"
#include "util.h"
#include "connection.h"


#define DSCUSS_CONNECTION_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_CONNECTION_DESCRIPTION_MAX_LEN];


/**
 * Handle for a network connection.
 */
struct _DscussConnection
{
  /**
   * Socket connection.
   */
  GSocketConnection* socket_connection;

  /**
   * @c TRUE if connection is incoming, @c FALSE otherwise.
   */
  gboolean is_incoming;

  /**
   * Cancellable for input operations.
   */
  GCancellable* icancellable;

  /**
   * Cancellable for output operations.
   */
  GCancellable* ocancellable;

  /**
   * Called when @c socket_connection receives new entities.
   */
  DscussConnectionReceiveCallback receive_callback;

  /**
   * User data for the @c receive_callback.
   */
  gpointer receive_data;

  /**
   * Queue of outgoing packets.
   */
  GQueue* oqueue;

  /**
   * Header of the packet to Buffer for reading packet data.
   */
  DscussHeader* header;

  /**
   * Buffer for reading packet data.
   */
  gchar read_buf[DSCUSS_PACKET_MAX_SIZE];

  /**
   * How many bytes have been written to @c read_buf.
   */
  gssize read_offset;
};


/**** ConnectionSendContext **************************************************/

typedef struct
{
  DscussConnection* connection;
  DscussConnectionSendCallback callback;
  gpointer user_data;
  const DscussPacket* packet;
  gchar* buffer;
  gsize length;
  gsize offset;
} ConnectionSendContext;


static ConnectionSendContext*
connecion_send_context_new (DscussConnection* connection,
                            const DscussPacket* packet,
                            DscussConnectionSendCallback callback,
                            gpointer user_data)
{
  ConnectionSendContext* ctx = g_new0 (ConnectionSendContext, 1);
  ctx->connection = connection;
  ctx->callback = callback;
  ctx->user_data = user_data;
  ctx->packet = packet;
  dscuss_packet_serialize (packet,
                           &ctx->buffer,
                           &ctx->length);
  ctx->offset = 0;
  return ctx;
}


static void
connection_send_context_free_full (ConnectionSendContext* ctx, gboolean result)
{
  if (ctx->callback != NULL)
    {
      ctx->callback (ctx->connection,
                     ctx->packet,
                     result,
                     ctx->user_data);
    }
  g_free (ctx->buffer);
  g_free (ctx);
}


static void
connecion_send_context_free (ConnectionSendContext* ctx)
{
  connection_send_context_free_full (ctx, FALSE);
}

/**** End of ConnectionSendContext *******************************************/


static void
send_head_packet (DscussConnection* connection);


static void
read_packet (DscussConnection* connection);


DscussConnection*
dscuss_connection_new (GSocketConnection* socket_connection,
                       gboolean is_incoming)
{
  DscussConnection* connection = g_new0 (DscussConnection, 1);
  connection->socket_connection = socket_connection;
  connection->is_incoming = is_incoming;
  connection->icancellable = g_cancellable_new ();
  connection->ocancellable = g_cancellable_new ();
  connection->receive_callback = NULL;
  connection->receive_data = NULL;
  connection->oqueue = g_queue_new ();
  connection->header = NULL;
  connection->read_offset = 0;
  return connection;
}


void
dscuss_connection_free (DscussConnection* connection)
{
  if (connection == NULL)
    return;

  if (connection->icancellable)
    {
      g_cancellable_cancel (connection->icancellable);
      g_object_unref (connection->icancellable);
    }
  if (connection->ocancellable)
    {
      g_cancellable_cancel (connection->ocancellable);
      g_object_unref (connection->ocancellable);
    }
  g_io_stream_close (G_IO_STREAM (connection->socket_connection), NULL, NULL);
  g_object_unref (connection->socket_connection);
  g_queue_free_full (connection->oqueue,
                     (GDestroyNotify) connecion_send_context_free);
  if (connection->header != NULL)
    dscuss_header_free (connection->header);
  g_free (connection);
  g_debug ("Connection successfully freed");
}


const gchar*
dscuss_connection_get_description (DscussConnection* connection)
{
  g_assert (connection != NULL);

  GSocketAddress* sockaddr =
    g_socket_connection_get_remote_address (connection->socket_connection,
                                            NULL);

  GInetAddress* addr =
    g_inet_socket_address_get_address (G_INET_SOCKET_ADDRESS(sockaddr));
  gchar* addr_str = g_inet_address_to_string (addr);
  guint16 port =
    g_inet_socket_address_get_port (G_INET_SOCKET_ADDRESS(sockaddr));
  g_snprintf (description_buf, 
              DSCUSS_CONNECTION_DESCRIPTION_MAX_LEN,
              "%s:%d", addr_str, port);
  g_free (addr_str);
  g_object_unref (sockaddr);
  return description_buf;
}


static void
ostream_write_cb (GObject* source, GAsyncResult* res, gpointer user_data)
{
  GOutputStream* out = G_OUTPUT_STREAM (source);
  ConnectionSendContext* ctx = user_data;
  GError* error = NULL;
  gssize nwrote;

  nwrote = g_output_stream_write_finish (out, res, &error);
  if (error && g_error_matches (error, G_IO_ERROR, G_IO_ERROR_CANCELLED))
    {
      g_debug ("Could not write to the connection:"
               " operation was canceled");
      g_error_free (error);
      return;
    }

  DscussConnection* connection = ctx->connection;
  if (error)
    {
      g_warning ("Could not write to the connection '%s': %s",
                 dscuss_connection_get_description (connection),
                 error->message);
      g_error_free (error);

      g_assert (g_queue_remove (connection->oqueue, ctx));
      connection_send_context_free_full (ctx, FALSE);
      return;
    }

  g_assert_cmpint (nwrote, <=, ctx->length - ctx->offset);

  ctx->offset += nwrote;
  if (ctx->offset == ctx->length)
    {
      g_debug ("Packet successfully written");
      g_assert (g_queue_remove (connection->oqueue, ctx));
      connection_send_context_free_full (ctx, TRUE);
      send_head_packet (connection);
    }
  else
    {
      g_debug ("Writing remaining %" G_GSIZE_FORMAT " bytes",
               ctx->length - ctx->offset);
      g_output_stream_write_async (out, ctx->buffer + ctx->offset,
                                   ctx->length - ctx->offset,
                                   G_PRIORITY_DEFAULT, connection->ocancellable,
                                   ostream_write_cb, ctx);
    }
}


static void
send_head_packet (DscussConnection* connection)
{
  GOutputStream* out = NULL;
  ConnectionSendContext* ctx = NULL;

  g_assert (connection != NULL);
  g_assert (connection->oqueue);

  if (! g_queue_is_empty (connection->oqueue))
    {
      ctx = g_queue_peek_head (connection->oqueue);
      g_debug ("Writing packet %s to the connection '%s'",
               dscuss_packet_get_description (ctx->packet),
               dscuss_connection_get_description (connection));
      out = g_io_stream_get_output_stream (G_IO_STREAM (connection->socket_connection));
      g_output_stream_write_async (out, ctx->buffer, ctx->length,
                                   G_PRIORITY_DEFAULT, connection->ocancellable,
                                   ostream_write_cb, ctx);
    }
}


void
dscuss_connection_send (DscussConnection* connection,
                        const DscussPacket* packet,
                        DscussConnectionSendCallback callback,
                        gpointer user_data)
{
  g_assert (connection != NULL);
  g_debug ("Sending packet %s",
           dscuss_packet_get_description (packet));

  ConnectionSendContext* ctx = connecion_send_context_new (connection,
                                                           packet,
                                                           callback,
                                                           user_data);
  g_queue_push_tail (connection->oqueue, ctx);

  /* Start processing the queue if it was empty. */
  if (g_queue_get_length (connection->oqueue) == 1)
    {
      send_head_packet (connection);
    }
}


static void
connection_clear_receive_callbacks (DscussConnection* connection)
{
  connection->receive_callback = NULL;
  connection->receive_data = NULL;
}


static void
istream_read_cb (GObject* source, GAsyncResult* res, gpointer user_data)
{
  GInputStream* in = G_INPUT_STREAM (source);
  DscussConnection* connection = user_data;
  GError* error = NULL;
  gssize nread;

  nread = g_input_stream_read_finish (in, res, &error);
  if (nread == -1)
    {
      if (g_error_matches (error, G_IO_ERROR, G_IO_ERROR_CANCELLED))
        {
          g_debug ("Could not read from the connection: operation was cancelled");
          g_error_free (error);
          return;
        }
      g_warning ("Could not read from the connection '%s': %s",
                 dscuss_connection_get_description (connection),
                 error->message);
      g_error_free (error);
      connection->receive_callback (connection,
                                    NULL,
                                    FALSE,
                                    connection->receive_data);
      connection_clear_receive_callbacks (connection);
      return;
    }
  if (nread == 0)
    {
      g_debug ("Could not read from the connection '%s':"
               " connection was closed",
               dscuss_connection_get_description (connection));
      connection->receive_callback (connection,
                                    NULL,
                                    FALSE,
                                    connection->receive_data);
      connection_clear_receive_callbacks (connection);
      return;
    }

  gssize length = (connection->read_offset < dscuss_header_get_size ()) ?
                   dscuss_header_get_size () :
                   dscuss_header_get_packet_size (connection->header);

  g_assert_cmpint (nread, <=, length - connection->read_offset);

  connection->read_offset += nread;
  /* Have we received all the requested data? */
  if (connection->read_offset == length)
    {
      /* Yes! Did we request just header? */
      if (length == dscuss_header_get_size ())
        {
          if (connection->header != NULL)
            dscuss_header_free (connection->header);
          connection->header = dscuss_header_deserialize (connection->read_buf);
          g_debug ("Packet header successfully read: %s",
                   dscuss_header_get_description (connection->header));
          gsize packet_size = dscuss_header_get_packet_size (connection->header);
          if (packet_size > DSCUSS_PACKET_MAX_SIZE)
            {
              g_warning ("Protocol violation detected:"
                         " packet size '%" G_GSIZE_FORMAT "' exceeds"
                         " maximum limit '%d'.",
                         packet_size, DSCUSS_PACKET_MAX_SIZE);
              connection->receive_callback (connection,
                                            NULL,
                                            FALSE,
                                            connection->receive_data);
              connection_clear_receive_callbacks (connection);
              return;
            }
          if (packet_size > dscuss_header_get_size ())
            {
              /* Receive the packet body */
              g_input_stream_read_async (in,
                                         connection->read_buf + dscuss_header_get_size (),
                                         packet_size - dscuss_header_get_size (),
                                         G_PRIORITY_DEFAULT, connection->icancellable,
                                         istream_read_cb, connection);
              return;
            }
        }
      g_debug ("Whole packet successfully read");
      DscussPacket* packet =
        dscuss_packet_deserialize (connection->header,
                                   connection->read_buf + dscuss_header_get_size ());
      if (packet == NULL)
        {
          g_warning ("Protocol violation detected: invalid packet");
          connection->receive_callback (connection,
                                        NULL,
                                        FALSE,
                                        connection->receive_data);
          connection_clear_receive_callbacks (connection);
          return;
        }
      if (connection->receive_callback (connection,
                                        packet,
                                        TRUE,
                                        connection->receive_data))
        read_packet (connection);
      else
        connection_clear_receive_callbacks (connection);
    }
  else
    {
      /* No. Read remaining data. */
      g_debug ("Reading remaining %" G_GSSIZE_FORMAT " bytes",
               length - connection->read_offset);
      g_input_stream_read_async (in, connection->read_buf + connection->read_offset,
                                 length - connection->read_offset,
                                 G_PRIORITY_DEFAULT, connection->icancellable,
                                 istream_read_cb, connection);
    }
}


static void
read_packet (DscussConnection* connection)
{
  GInputStream* in = NULL;

  g_assert (connection != NULL);
  g_debug ("Trying to read from the connection '%s' %" G_GSIZE_FORMAT " bytes",
            dscuss_connection_get_description (connection),
            dscuss_header_get_size ());
  in = g_io_stream_get_input_stream (G_IO_STREAM (connection->socket_connection));
  connection->read_offset = 0;
  g_input_stream_read_async (in, connection->read_buf,
                             dscuss_header_get_size (),
			     G_PRIORITY_DEFAULT, connection->icancellable,
			     istream_read_cb, connection);
}


void
dscuss_connection_set_receive_callback (DscussConnection* connection,
                                        DscussConnectionReceiveCallback callback,
                                        gpointer user_data)
{
  g_assert (connection != NULL);
  g_assert (callback != NULL);

  gboolean start_reading = (connection->receive_callback == NULL);

  connection->receive_callback = callback;
  connection->receive_data = user_data;

  if (start_reading)
    read_packet (connection);
}


gboolean
dscuss_connection_is_incoming (DscussConnection* connection)
{
  g_assert (connection != NULL);
  return connection->is_incoming;
}


void
dscuss_connection_cancel_all_io (DscussConnection* connection)
{
  g_assert (connection != NULL);

  dscuss_connection_cancel_io (connection, DSCUSS_CONNECTION_IO_TYPE_RX);
  dscuss_connection_cancel_io (connection, DSCUSS_CONNECTION_IO_TYPE_TX);
}


void
dscuss_connection_cancel_io (DscussConnection* connection,
                             DscussConnectionIoType io_type)
{
  g_assert (connection != NULL);

  g_debug ("Cancelling I/O of the type %d of the connection '%s'",
           io_type,
           dscuss_connection_get_description (connection));

  switch (io_type)
    {
    case DSCUSS_CONNECTION_IO_TYPE_TX:
      if (connection->ocancellable != NULL)
        {
          g_queue_free_full (connection->oqueue,
                             (GDestroyNotify) connecion_send_context_free);
          connection->oqueue = g_queue_new ();
          g_cancellable_cancel (connection->ocancellable);
          g_object_unref (connection->ocancellable);
          connection->ocancellable = g_cancellable_new ();
        }
      break;

    case DSCUSS_CONNECTION_IO_TYPE_RX:
      if (connection->icancellable != NULL)
        {
          connection->receive_callback = NULL;
          g_cancellable_cancel (connection->icancellable);
          g_object_unref (connection->icancellable);
          connection->icancellable = g_cancellable_new ();
        }
      break;

    default:
      g_assert_not_reached ();
    }
}
