/**
 * This file is part of Dscuss.
 * Copyright (C) 2014-2016  Vitaly Minko
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

#include <string.h>
#include <glib.h>
#include "config.h"
#include "packet.h"
#include "message.h"
#include "util.h"
#include "connection.h"
#include "subscriptions.h"
#include "payload_hello.h"
#include "handshake.h"


#define DSCUSS_HANDSHAKE_MAX_TIMESTAMP_DISCREPANCY  300
#define DSCUSS_HANDSHAKE_TIMEOUT                    15


/**
 * Handle for handshaking.
 */
struct _DscussHandshakeHandle
{
  DscussConnection* connection;
  const DscussUser* self;
  DscussPrivateKey* self_privkey;
  GSList* self_subscriptions;
  DscussDb* dbh;
  DscussHandshakeCallback callback;
  gpointer user_data;
  guint fail_id;
  guint timeout_id;
  DscussUser* peers_user;
};


static DscussHandshakeHandle*
handshake_handle_new (DscussConnection* connection,
                      const DscussUser* self,
                      DscussPrivateKey* self_privkey,
                      GSList* self_subscriptions,
                      DscussDb* dbh,
                      DscussHandshakeCallback callback,
                      gpointer user_data)
{
  DscussHandshakeHandle* handle = g_new0 (DscussHandshakeHandle, 1);
  handle->connection = connection;
  handle->self = self;
  handle->self_privkey = self_privkey;
  handle->self_subscriptions = self_subscriptions;
  handle->dbh = dbh;
  handle->callback = callback;
  handle->user_data = user_data;
  handle->fail_id = 0;
  handle->timeout_id = 0;
  handle->peers_user = NULL;
  return handle;
}


void
handshake_handle_free (DscussHandshakeHandle* handle)
{
  if (handle->fail_id != 0)
    {
      g_source_remove (handle->fail_id);
    }
  if (handle->timeout_id != 0)
    {
      g_source_remove (handle->timeout_id);
    }
  dscuss_free_non_null (handle->peers_user, dscuss_user_free);
  g_free (handle);
}


static gboolean
handshake_fail (gpointer user_data)
{
  DscussHandshakeHandle* handle = user_data;
  g_assert (handle != NULL);

  g_debug ("Handshake error: failed to handshake with the node '%s'",
           dscuss_connection_get_description (handle->connection));

  handle->fail_id = 0;
  dscuss_connection_cancel_io (handle->connection);
  handle->callback (FALSE, NULL, NULL, handle->user_data);
  handshake_handle_free (handle);

  return FALSE;
}


static void
handshake_schedule_fail (DscussHandshakeHandle* handle)
{
  if (handle->fail_id != 0)
    {
      g_source_remove (handle->fail_id);
    }
  handle->fail_id = g_idle_add_full (G_PRIORITY_HIGH,
                                     handshake_fail,
                                     handle,
                                     NULL);
}


static gboolean
handshake_on_hello_received (DscussConnection* connection,
                             DscussPacket* packet,
                             gboolean result,
                             gpointer user_data)
{
  DscussHandshakeHandle* handle = user_data;
  gchar* payload = NULL;
  gsize payload_size = 0;
  DscussPayloadHello* pld_hello = NULL;
  GDateTime* datetime_now = NULL;
  gint64 ts_discrepancy = 0;
  gboolean handshake_result = FALSE;
  GSList* subscriptions = NULL;

  if (!result)
    {
      g_debug ("Handshake error: failed to read Hello from connection '%s'",
               dscuss_connection_get_description (connection));
      goto out;
    }
  g_debug ("Handshaking: received Hello from the node '%s'",
           dscuss_connection_get_description (connection));


  DscussPacketType type = dscuss_packet_get_type (packet);
  if (type != DSCUSS_PACKET_TYPE_HELLO)
    {
      g_warning ("Handshake error: protocol violation detected:"
                 " node '%s' sent unexpected packet of type '%d'."
                 " Expected: %d (peer's user for handshaking)",
                 dscuss_connection_get_description (connection), type,
                 DSCUSS_PACKET_TYPE_HELLO);
      goto out;
    }

  if (!dscuss_packet_verify (packet,
                             dscuss_user_get_public_key (handle->peers_user)))
    {
      g_warning ("Handshake error: signature of the Hello packet is invalid");
      goto out;
    }

  dscuss_packet_get_payload (packet,
                             (const gchar**) &payload,
                             &payload_size);

  pld_hello = dscuss_payload_hello_deserialize (payload,
                                                payload_size);
  if (pld_hello == NULL)
    {
      g_warning ("Handshake error: failed to parse the Hello payload");
      goto out;
    }

   /* Validate Hello fields */
  if (memcmp (dscuss_user_get_id (handle->peers_user),
              dscuss_payload_hello_get_receiver_id (pld_hello),
              sizeof (DscussHash)))
    {
      g_warning ("Handshake error: wrong receiver ID: '%s'",
                 dscuss_crypto_hash_to_string (dscuss_payload_hello_get_receiver_id (pld_hello)));
      g_debug ("Expected receiver ID: '%s'",
                 dscuss_crypto_hash_to_string (dscuss_user_get_id (handle->peers_user)));
      goto out;
    }

  datetime_now = g_date_time_new_now_utc ();
  ts_discrepancy = ABS (g_date_time_to_unix (datetime_now) -
                        g_date_time_to_unix (dscuss_payload_hello_get_datetime (pld_hello)));
  if (ts_discrepancy > DSCUSS_HANDSHAKE_MAX_TIMESTAMP_DISCREPANCY)
    {
      g_warning ("Handshake error: timestamp discrepancy exceeds the limit:"
                 " %" G_GINT64_FORMAT, ts_discrepancy);
      goto out;
    }

  /* Finally handshake succeeded */
  subscriptions = dscuss_subscriptions_copy (dscuss_payload_hello_get_subscriptions (pld_hello));
  dscuss_connection_cancel_io (handle->connection);
  handle->callback (TRUE,
                    handle->peers_user,
                    subscriptions,
                    handle->user_data);
  handle->peers_user = NULL;
  handshake_handle_free (handle);
  handshake_result = TRUE;

out:
  dscuss_free_non_null (packet, dscuss_packet_free);
  dscuss_free_non_null (datetime_now, g_date_time_unref);
  dscuss_free_non_null (pld_hello, dscuss_payload_hello_free);
  if (!handshake_result)
    handshake_schedule_fail (handle);
  return FALSE;
}


static void
handshake_on_hello_sent (DscussConnection* connection,
                         const DscussPacket* packet,
                         gboolean result,
                         gpointer user_data)
{
  DscussHandshakeHandle* handle = user_data;

  dscuss_packet_free ((DscussPacket*) packet);

  if (!result)
    {
      g_warning ("Handshake error: failed to send hello to the node '%s'",
                 dscuss_connection_get_description (connection));
      handshake_schedule_fail (handle);
    }
  else
    {
      g_debug ("Handshaking: Hello successfully sent to the node '%s'",
               dscuss_connection_get_description (connection));
    }
}


static gboolean
handshake_send_hello (DscussHandshakeHandle* handle)
{
  DscussPayloadHello* pld_hello = NULL;
  gchar* serialized_hello = NULL;
  gsize serialized_hello_len = 0;
  DscussPacket* hello_packet = NULL;
  gboolean result = FALSE;

  g_assert (handle != NULL);

  g_debug ("Handshaking: trying to send Hello to the node '%s'",
           dscuss_connection_get_description (handle->connection));

  pld_hello = dscuss_payload_hello_new (dscuss_user_get_id (handle->self),
                                        handle->self_subscriptions);
  if (!dscuss_payload_hello_serialize (pld_hello,
                                       &serialized_hello,
                                       &serialized_hello_len))
    {
      g_warning ("Handshake error: failed to serialize the Hello payload");
      goto out;
    }

  hello_packet = dscuss_packet_new (DSCUSS_PACKET_TYPE_HELLO,
                                    serialized_hello,
                                    serialized_hello_len);
  g_free (serialized_hello);
  dscuss_packet_sign (hello_packet, handle->self_privkey);
  dscuss_connection_send (handle->connection,
                          hello_packet,
                          handshake_on_hello_sent,
                          handle);
  result = TRUE;

out:
  dscuss_free_non_null (pld_hello, dscuss_payload_hello_free);
  return result;
}


static gboolean
handshake_on_user_received (DscussConnection* connection,
                            DscussPacket* packet,
                            gboolean result,
                            gpointer user_data)
{
  DscussHandshakeHandle* handle = user_data;
  gchar* payload = NULL;
  gsize payload_size = 0;
  DscussUser* user = NULL;
  DscussUser* stored_user = NULL;
  gboolean handle_user_result = FALSE;

  if (!result)
    {
      g_debug ("Handshake error: failed to read User from connection '%s'",
               dscuss_connection_get_description (connection));
      goto out;
    }
  g_debug ("Handshaking: received User from the connection '%s'",
           dscuss_connection_get_description (handle->connection));

  DscussPacketType type = dscuss_packet_get_type (packet);
  if (type != DSCUSS_PACKET_TYPE_USER)
    {
      g_warning ("Handshake error: protocol violation detected:"
                 " node '%s' sent unexpected packet of type '%d'."
                 " Expected: %d (peer's user for handshaking)",
                 dscuss_connection_get_description (handle->connection), type,
                 DSCUSS_PACKET_TYPE_USER);
      goto out;
    }

  dscuss_packet_get_payload (packet,
                             (const gchar**) &payload,
                             &payload_size);
  user = dscuss_user_deserialize (payload,
                                  payload_size);
  if (user == NULL)
    {
      g_debug ("Handshake error: failed to parse the User");
      goto out;
    }

  /* TBD: optimize: dscuss_db_has_user() */
  stored_user = dscuss_db_get_user (handle->dbh,
                                    dscuss_user_get_id (user));
  if (stored_user == NULL)
    {
      if (!dscuss_db_put_user (handle->dbh, user))
        {
          g_warning ("Handshake error:"
                     " failed to store the user '%s' of the node '%s'",
                     dscuss_user_get_description (user),
                     dscuss_connection_get_description (handle->connection));
          goto out;
        }
    }
  handle->peers_user = user;

  /* Send our user in response. */
  if (!handshake_send_hello (handle))
    {
      g_warning ("Handshake error: failed to send Hello to the node");
      goto out;
    }
  dscuss_connection_set_receive_callback (handle->connection,
                                          handshake_on_hello_received,
                                          handle);
  handle_user_result = TRUE;

out:
  dscuss_free_non_null (packet, dscuss_packet_free);
  dscuss_free_non_null (stored_user, dscuss_user_free);
  if (!handle_user_result)
    {
      dscuss_free_non_null (user, dscuss_user_free);
      handle->peers_user = NULL;
      handshake_schedule_fail (handle);
    }
  return handle_user_result;
}


static void
handshake_on_user_sent (DscussConnection* connection,
                        const DscussPacket* packet,
                        gboolean result,
                        gpointer user_data)
{
  DscussHandshakeHandle* handle = user_data;

  dscuss_packet_free ((DscussPacket*) packet);

  if (!result)
    {
      g_warning ("Handshake error: failed to send our user to the node '%s'",
                 dscuss_connection_get_description (connection));
      handshake_schedule_fail (handle);
    }
  else
    {
      g_debug ("Handshaking: our User successfully sent to the node '%s'",
               dscuss_connection_get_description (connection));
    }
}


static gboolean
handshake_send_user (DscussHandshakeHandle* handle)
{
  gchar* serialized_user = NULL;
  gsize serialized_user_len = 0;
  DscussPacket* user_packet = NULL;

  g_assert (handle != NULL);

  g_debug ("Handshaking: trying to send out User to the node '%s'",
           dscuss_connection_get_description (handle->connection));

  if (!dscuss_user_serialize (handle->self,
                              &serialized_user,
                              &serialized_user_len))
    {
      g_warning ("Handshake error: failed to serialize the user '%s'",
                 dscuss_user_get_description (handle->self));
      return FALSE;
    }
  user_packet = dscuss_packet_new (DSCUSS_PACKET_TYPE_USER,
                                   serialized_user,
                                   serialized_user_len);
  g_free (serialized_user);
  /* This packet goes without signature. */
  dscuss_connection_send (handle->connection,
                          user_packet,
                          handshake_on_user_sent,
                          handle);
  return TRUE;
}


DscussHandshakeHandle*
dscuss_handshake_start (DscussConnection* connection,
                        const DscussUser* self,
                        DscussPrivateKey* self_privkey,
                        GSList* self_subscriptions,
                        DscussDb* dbh,
                        DscussHandshakeCallback callback,
                        gpointer user_data)
{
  DscussHandshakeHandle* handle = NULL;

  g_assert (connection != NULL);
  g_assert (self != NULL);
  g_assert (self_privkey != NULL);
  g_assert (self_subscriptions != NULL);
  g_assert (dbh != NULL);
  g_assert (callback != NULL);

  g_debug ("Handshaking: starting handshake process with '%s'",
           dscuss_connection_get_description (connection));

  handle = handshake_handle_new (connection,
                                 self,
                                 self_privkey,
                                 self_subscriptions,
                                 dbh,
                                 callback,
                                 user_data);

  if (!handshake_send_user (handle))
    {
      g_warning ("Handshake error: failed to send our user to the peer");
      handshake_schedule_fail (handle);
      return handle;
    }

  /* Timeout for failing handshake */
  handle->timeout_id = g_timeout_add_seconds (DSCUSS_HANDSHAKE_TIMEOUT,
                                              handshake_fail,
                                              handle);
  dscuss_connection_set_receive_callback (handle->connection,
                                          handshake_on_user_received,
                                          handle);
  return handle;
}


void
dscuss_handshake_cancel (DscussHandshakeHandle* handle)
{
  g_assert (handle != NULL);
  dscuss_connection_cancel_io (handle->connection);
  handshake_handle_free (handle);
}
