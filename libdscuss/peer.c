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
#include "peer.h"
#include "subscriptions.h"
#include "handshake.h"


#define DSCUSS_PEER_DESCRIPTION_MAX_LEN  120

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
   * and @c type_context is a const @a gpointer to arbitrary data for handling
   * packet if this type (maybe @c NULL if no context required).
   */
  GHashTable* expected_types;

  /**
   * Handle for handshaking.
   */
  DscussHandshakeHandle* handshake_handle;

  /**
   * TRUE if we've handshaked with this peer.
   */
  gboolean is_handshaked;

  /**
   * Peer's user.
   */
  DscussUser* user;

  /**
   * Peer's subscriptions.
   */
  GSList* subscriptions;
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
peer_send_context_free (PeerSendContext* ctx)
{
  dscuss_entity_unref (ctx->entity);
  g_free (ctx);
}

/**** End of PeerSendContext *************************************************/


/**** PeerHandshakeContext ***************************************************/

typedef struct
{
  DscussPeer* peer;
  DscussPeerHandshakeCallback callback;
  gpointer user_data;
} PeerHandshakeContext;


static PeerHandshakeContext*
peer_handshake_context_new (DscussPeer* peer,
                            DscussPeerHandshakeCallback callback,
                            gpointer user_data)
{
  PeerHandshakeContext* ctx = g_new0 (PeerHandshakeContext, 1);
  ctx->peer = peer;
  ctx->callback = callback;
  ctx->user_data = user_data;
  return ctx;
}


static void
peer_handshake_context_free (PeerHandshakeContext* ctx)
{
  g_free (ctx);
}

/**** End of PeerHandshakeContext ********************************************/


static gboolean
on_new_packet (DscussConnection* connection,
               DscussPacket* packet,
               gboolean result,
               gpointer user_data)
{
  DscussPeer* peer = user_data;
  gpointer type_context = NULL;

  if (!result)
    {
      g_debug ("Failed to read from connection '%s'",
               dscuss_connection_get_description (connection));
      peer->receive_callback (peer,
                              NULL,
                              FALSE,
                              peer->receive_data);
      return FALSE;
    }

  g_debug ("New packet received from peer '%s': %s",
           dscuss_peer_get_description (peer),
           dscuss_packet_get_description (packet));

  DscussPacketType type = dscuss_packet_get_type (packet);
  gboolean is_found = g_hash_table_lookup_extended (peer->expected_types,
                                                    GINT_TO_POINTER (type),
                                                    NULL,
                                                    &type_context);
  if (!is_found)
    {
      g_warning ("Protocol violation detected:"
                 " peer '%s' sent unexpected packet of type '%d'.",
                 dscuss_peer_get_description (peer), type);
      dscuss_packet_free (packet);
      peer->receive_callback (peer,
                              NULL,
                              FALSE,
                              peer->receive_data);
      return FALSE;
    }

  switch (type)
    {
    case DSCUSS_PACKET_TYPE_USER:
      /* TBD */
      g_assert_not_reached ();
      break;

    case DSCUSS_PACKET_TYPE_MSG:
      g_debug ("This is a Message packet");

      gchar* payload = NULL;
      gsize payload_size = 0;
      dscuss_packet_get_payload (packet,
                                 (const gchar**) &payload,
                                 &payload_size);
      DscussMessage* msg = dscuss_message_deserialize (payload,
                                                       payload_size);
      if (msg == NULL)
        {
          g_warning ("Malformed Message packet: failed to parse.");
          dscuss_packet_free (packet);
          peer->receive_callback (peer,
                                  NULL,
                                  FALSE,
                                  peer->receive_data);
          return FALSE;
        }
      peer->receive_callback (peer,
                              (DscussEntity*) msg,
                              TRUE,
                              peer->receive_data);
      break;

    case DSCUSS_PACKET_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }
  dscuss_packet_free (packet);
  return TRUE;
}


DscussPeer*
dscuss_peer_new (GSocketConnection* socket_connection,
                 gboolean is_incoming,
                 DscussPeerDisconnectCallback disconn_callback,
                 gpointer disconn_data)
{
  DscussPeer* peer = g_new0 (DscussPeer, 1);
  peer->connection = dscuss_connection_new (socket_connection,
                                            is_incoming);
  peer->disconn_callback = disconn_callback;
  peer->disconn_data = disconn_data;
  peer->handshake_handle = NULL;
  peer->is_handshaked = FALSE;
  peer->receive_callback = NULL;
  peer->receive_data = NULL;
  peer->user = NULL;
  peer->subscriptions = NULL;
  peer->expected_types = g_hash_table_new_full (g_direct_hash,
                                                g_direct_equal,
                                                NULL,
                                                NULL);
  return peer;
}


void
dscuss_peer_free_full (DscussPeer* peer,
                       DscussPeerDisconnectReason reason,
                       gpointer reason_data)
{
  if (peer == NULL)
    return;

  peer->disconn_callback (peer,
                          reason,
                          reason_data,
                          peer->disconn_data);

  if (peer->handshake_handle != NULL)
    dscuss_handshake_cancel (peer->handshake_handle);

  dscuss_free_non_null (peer->user, dscuss_user_free);
  dscuss_free_non_null (peer->subscriptions, dscuss_subscriptions_free);
  dscuss_free_non_null (peer->connection, dscuss_connection_free);
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


const DscussUser*
dscuss_peer_get_user (const DscussPeer* peer)
{
  g_assert (peer != NULL);
  return (peer->is_handshaked) ? peer->user : NULL;
}


const GSList*
dscuss_peer_get_subscriptions (const DscussPeer* peer)
{
  g_assert (peer != NULL);
  return (peer->is_handshaked) ? peer->subscriptions : NULL;
}


const gchar*
dscuss_peer_get_description (DscussPeer* peer)
{
  gchar* id_str = NULL;

  g_assert (peer != NULL);

  if (peer->is_handshaked)
    {
      id_str = dscuss_data_to_hex ((const gpointer) dscuss_user_get_id(peer->user),
                                   sizeof (DscussHash),
                                   NULL);
      id_str[5] = '\0';
      g_snprintf (description_buf,
                  DSCUSS_PEER_DESCRIPTION_MAX_LEN,
                  "%s-%s",
                  dscuss_user_get_nickname (peer->user),
                  id_str);
      g_free (id_str);
    }
  else
    {
      g_snprintf (description_buf,
                  DSCUSS_PEER_DESCRIPTION_MAX_LEN,
                  "(not handshaked), %s",
                  dscuss_connection_get_description (peer->connection));
    }

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
  ctx->callback (ctx->peer,
                 ctx->entity,
                 result,
                 ctx->user_data);
  peer_send_context_free (ctx);
}


gboolean
dscuss_peer_send (DscussPeer* peer,
                  DscussEntity* entity,
                  DscussPrivateKey* privkey,
                  DscussPeerSendCallback callback,
                  gpointer user_data)
{
  DscussPacket* packet = NULL;
  DscussMessage* msg = NULL;
  gchar* serialized_payload = NULL;
  gsize serialized_payload_len = 0;

  g_assert (peer != NULL);
  g_assert (peer->connection != NULL);
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
      if (!dscuss_message_serialize (msg,
                                     &serialized_payload,
                                     &serialized_payload_len))
        {
          g_warning ("Failed to serialize the message '%s'",
                     dscuss_message_get_description (msg));
          return FALSE;
        }
      packet = dscuss_packet_new (DSCUSS_PACKET_TYPE_MSG,
                                  serialized_payload,
                                  serialized_payload_len);
      dscuss_packet_sign (packet, privkey);
      g_free (serialized_payload);
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
  return TRUE;
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


gboolean
dscuss_peer_is_handshaked (DscussPeer* peer)
{
  g_assert (peer != NULL);
  return peer->is_handshaked;
}


static void
on_handshaked (gboolean result,
               DscussUser* peers_user,
               GSList* peers_subscriptions,
               gpointer user_data)
{
  PeerHandshakeContext* ctx = user_data;
  g_assert (user_data != NULL);

  DscussPeer* peer = ctx->peer;
  g_assert (peer != NULL);

  peer->handshake_handle = NULL;

  if (result)
    {
      g_debug ("Successfully handshaked with the peer '%s'",
               dscuss_peer_get_description (peer));
      peer->user = peers_user;
      peer->subscriptions = peers_subscriptions;
      peer->is_handshaked = TRUE;
      /* TBD: find better solution.
       * dscuss_connection_set_receive_callback will be set via
       * dscuss_peer_set_receive_callback
       **/
      /* Now we expect ordinary entities: messages, TBD: users and operations */
      g_hash_table_insert (peer->expected_types,
                           GINT_TO_POINTER (DSCUSS_PACKET_TYPE_MSG), NULL);
    }
  else
    {
      g_debug ("Failed to handshake with the peer '%s'",
               dscuss_peer_get_description (peer));
      dscuss_free_non_null (peer->connection, dscuss_connection_free);
    }
  ctx->callback (ctx->peer,
                 result,
                 ctx->user_data);
  peer_handshake_context_free (ctx);
}


void
dscuss_peer_handshake (DscussPeer* peer,
                       const DscussUser* self,
                       DscussPrivateKey* self_privkey,
                       GSList* self_subscriptions,
                       DscussDb* dbh,
                       DscussPeerHandshakeCallback callback,
                       gpointer user_data)
{
  g_assert (self != NULL);
  g_assert (self_privkey != NULL);
  g_assert (self_subscriptions != NULL);
  g_assert (callback != NULL);

  PeerHandshakeContext* ctx = peer_handshake_context_new (peer,
                                                          callback,
                                                          user_data);
  peer->handshake_handle = dscuss_handshake_start (peer->connection,
                                                   self,
                                                   self_privkey,
                                                   self_subscriptions,
                                                   dbh,
                                                   on_handshaked,
                                                   ctx);
}

