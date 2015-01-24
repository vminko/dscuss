/**
 * This file is part of Dscuss.
 * Copyright (C) 2014  Vitaly Minko
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
#include "header.h"
#include "packet.h"
#include "message.h"
#include "util.h"
#include "connection.h"
#include "peer.h"


#define DSCUSS_PEER_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_PEER_DESCRIPTION_MAX_LEN];


/**
 * Handle for a peer.
 */
struct _DscussPeer
{
  /**
   * Connection with the peer.
   */
  DscussConnection* connection;

  /**
   * Called when @c connection gets terminated.
   */
  DscussPeerDisconnectCallback disconn_callback;

  /**
   * User data for the @c disconn_callback.
   */
  gpointer disconn_data;

  /**
   * Called when we've successfully handshaked with this peer.
   */
  DscussPeerHandshakeCallback handshake_callback;

  /**
   * User data for the @c handshake_callback.
   */
  gpointer handshake_data;

  /**
   * Called when we receive new entities from this peer.
   */
  DscussPeerReceiveCallback receive_callback;

  /**
   * User data for the @c receive_callback.
   */
  gpointer receive_data;

  /**
   * Hash of expected packet types in the following format
   * [type_id -> type_context]
   * where @c type_id is a @a DscussPacketType,
   * and @c type_context is a @a gpointer to arbitrary data for handling
   * packet if this type (maybe @c NULL if no context required).
   */
  GHashTable* expected_types;

  /**
   * Peer ID (hash of its public key)..
   */
  // TBD: id;
};


/**** PeerSendContext ********************************************************/

typedef struct
{
  DscussPeer* peer;
  DscussEntity* entity;
  DscussPeerSendCallback callback;
  gpointer user_data;
} PeerSendContext;


static PeerSendContext*
peer_send_context_new (DscussPeer* peer,
                       DscussEntity* entity,
                       DscussPeerSendCallback callback,
                       gpointer user_data)
{
  dscuss_entity_ref (entity);
  PeerSendContext* ctx = g_new0 (PeerSendContext, 1);
  ctx->peer = peer;
  ctx->entity = entity;
  ctx->callback = callback;
  ctx->user_data = user_data;
  return ctx;
}


static void
peer_send_context_free (PeerSendContext* ctx, gboolean result)
{
  ctx->callback (ctx->peer,
                 ctx->entity,
                 result,
                 ctx->user_data);
  dscuss_entity_unref (ctx->entity);
  g_free (ctx);
}

/**** End of PeerSendContext *************************************************/


void
on_new_packet (DscussConnection* connection,
               DscussPacket* packet,
               gboolean result,
               gpointer user_data)
{
  DscussPeer* peer = user_data;

  if (!result)
    {
      g_debug ("Failed to read from connection '%s'",
               dscuss_connection_get_description (connection));
      peer->receive_callback (peer,
                              NULL,
                              FALSE,
                              peer->receive_data);
      return;
    }

  g_debug ("New packet received from peer '%s': %s",
           dscuss_peer_get_description (peer),
           dscuss_packet_get_description (packet));
  switch (dscuss_packet_get_type (packet))
    {
    case DSCUSS_PACKET_TYPE_USER:
      g_assert_not_reached ();
      /* TBD */
      break;

    case DSCUSS_PACKET_TYPE_MSG:
      g_debug ("This is a Message packet");

      gchar* payload = NULL;
      gsize payload_size = 0;

      dscuss_packet_get_payload (packet,
                                 (const gchar**) &payload,
                                 &payload_size);

      if (payload[payload_size - 1] != '\0')
        {
          g_warning ("Malformed Message packet payload, fixing '\\0'.");
          payload[payload_size - 1] = '\0';
        }

      /* Payload of the message packet is the actual message in
       * plaintext */
      DscussMessage* msg = dscuss_message_new (payload);
      peer->receive_callback (peer,
                              (DscussEntity*) msg,
                              TRUE,
                              peer->receive_data);
      break;;

    case DSCUSS_PACKET_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }
  dscuss_packet_free (packet);
}


DscussPeer*
dscuss_peer_new (GSocketConnection* socket_connection,
                 gboolean is_incoming,
                 DscussPeerDisconnectCallback disconn_callback,
                 gpointer disconn_data,
                 DscussPeerHandshakeCallback handshake_callback,
                 gpointer handshake_data)
{
  DscussPeer* peer = g_new0 (DscussPeer, 1);
  peer->connection = dscuss_connection_new (socket_connection,
                                            is_incoming);
  peer->disconn_callback = disconn_callback;
  peer->disconn_data = disconn_data;
  peer->handshake_callback = handshake_callback;
  peer->handshake_data = handshake_data;
  peer->receive_callback = NULL;
  peer->receive_data = NULL;
  peer->expected_types = g_hash_table_new_full (g_direct_hash,
                                                g_direct_equal,
                                                NULL,
                                                NULL);

  if (dscuss_connection_is_incoming (peer->connection))
    {
      g_hash_table_insert (peer->expected_types,
                           GINT_TO_POINTER (DSCUSS_PACKET_TYPE_HELLO),
                           NULL);
    }
  else
    {
      /*DscussPacket* packet = dscuss_packet_hello_new (PubKey,
                                                      RegDate,
                                                      NickName,
                                                      Interests,
                                                      );
      dscuss_connection_send (peer->connection,
                              packet,
                              on_hello_sent,
                              peer);*/
    }

  return peer;
}


/*
on_hello_sent (DscussConnection* connection,
               const DscussPacket* packet,
               gboolean result,
               gpointer user_data)
{
  DscussPeer* peer = user_data;
  g_hash_table_insert (peer->expected_types,
                       DSCUSS_PACKET_TYPE_HELLO,
                       NULL);
}*/


static void
dscuss_peer_free_full (DscussPeer* peer,
                       DscussPeerDisconnectReason reason,
                       gpointer reason_data)
{
  if (peer == NULL)
    return;

  peer->disconn_callback (peer,
                          reason,
                          NULL,
                          peer->disconn_data);

  dscuss_connection_free (peer->connection);
  g_hash_table_destroy (peer->expected_types);
  g_free (peer);
  g_debug ("Peer successfully freed");
}


void
dscuss_peer_free (DscussPeer* peer)
{
  dscuss_peer_free_full (peer,
                         DSCUSS_PEER_DISCONNECT_REASON_CLOSED,
                         NULL);
}


const gchar*
dscuss_peer_get_description (DscussPeer* peer)
{
  g_assert (peer != NULL);
  /* TBD */
  g_snprintf (description_buf, 
              DSCUSS_PEER_DESCRIPTION_MAX_LEN,
              "%s",
              "TBD: show peer's id");
  return description_buf;
}


const gchar*
dscuss_peer_get_connecton_description (DscussPeer* peer)
{
  g_assert (peer != NULL);
  return dscuss_connection_get_description (peer->connection);
}


void
on_packet_sent (DscussConnection* connection,
                const DscussPacket* packet,
                gboolean result,
                gpointer user_data)
{
  PeerSendContext* ctx = user_data;
  if (!result)
    g_debug ("Failed to send packet %s to the peer '%s'",
             dscuss_packet_get_description (packet),
             dscuss_peer_get_description (ctx->peer));
  dscuss_packet_free ((DscussPacket*) packet);
  peer_send_context_free (ctx, result);
}


void
dscuss_peer_send (DscussPeer* peer,
                  DscussEntity* entity,
                  DscussPeerSendCallback callback,
                  gpointer user_data)
{
  DscussPacket* packet = NULL;
  DscussMessage* msg = NULL;

  g_assert (peer != NULL);
  g_assert (entity != NULL);
  g_debug ("Sending entity '%s'",
           dscuss_entity_get_description (entity));

  /* FIXME: use actual signatures */
  struct DscussSignature signature;
  memset (&signature, 0, sizeof (struct DscussSignature));

  switch (dscuss_entity_get_type (entity))
    {
    case DSCUSS_ENTITY_TYPE_USER:
      g_assert_not_reached ();
      /* TBD */
      break;

    case DSCUSS_ENTITY_TYPE_MSG:
      msg = (DscussMessage*) entity;
      const gchar* msg_text = dscuss_message_get_content (msg);
      packet = dscuss_packet_new_full (DSCUSS_PACKET_TYPE_MSG,
                                       msg_text,
                                       strlen (msg_text) + 1,
                                       &signature);
      break;

    case DSCUSS_ENTITY_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }

  PeerSendContext* ctx = peer_send_context_new (peer,
                                                entity,
                                                callback,
                                                user_data);
  dscuss_connection_send (peer->connection,
                          packet,
                          on_packet_sent,
                          ctx);
}


void
dscuss_peer_set_receive_callback (DscussPeer* peer,
                                  DscussPeerReceiveCallback callback,
                                  gpointer user_data)
{
  g_assert (peer != NULL);
  if (peer->receive_callback != NULL ||
      peer->receive_data != NULL)
    {
      g_warning ("Attempt to override DscussPeerReceiveCallback");
      return;
    }
  peer->receive_callback = callback;
  peer->receive_data = user_data;
  dscuss_connection_set_receive_callback (peer->connection,
                                          on_new_packet,
                                          peer);
}
