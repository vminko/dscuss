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

#include <string.h>
#include <glib.h>
#include "config.h"
#include "header.h"
#include "packet.h"
#include "message.h"
#include "util.h"
#include "connection.h"
#include "peer.h"
#include "subscriptions.h"
#include "payload_hello.h"


#define DSCUSS_PEER_DESCRIPTION_MAX_LEN        120
#define DSCUSS_PEER_MAX_TIMESTAMP_DISCREPANCY  300
#define DSCUSS_PEER_HANDSHAKE_TIMEOUT          15

static gchar description_buf[DSCUSS_PEER_DESCRIPTION_MAX_LEN];


struct PeerHandshakeContext;
static void
peer_handshake_context_free (struct PeerHandshakeContext* ctx);


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
   * Context for handshaking.
   */
  struct PeerHandshakeContext* handshake_ctx;

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
  peer->handshake_ctx = NULL;
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

  if (peer->handshake_ctx != NULL)
    peer_handshake_context_free (peer->handshake_ctx);

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
  peer_send_context_free (ctx, result);
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


/**** HANDSHAKING ************************************************************/

/**** PeerHandshakeContext ****/

struct PeerHandshakeContext
{
  const DscussUser* user;
  DscussPrivateKey* privkey;
  GSList* subscriptions;
  DscussDb* dbh;
  DscussPeerHandshakeCallback callback;
  gpointer user_data;
  guint handshake_fail_id;
  guint handshake_timeout_id;
};


static struct PeerHandshakeContext*
peer_handshake_context_new (const DscussUser* user,
                            DscussPrivateKey* privkey,
                            GSList* subscriptions,
                            DscussDb* dbh,
                            DscussPeerHandshakeCallback callback,
                            gpointer user_data)
{
  struct PeerHandshakeContext* ctx = g_new0 (struct PeerHandshakeContext, 1);
  ctx->user = user;
  ctx->privkey = privkey;
  ctx->subscriptions = subscriptions;
  ctx->dbh = dbh;
  ctx->callback = callback;
  ctx->user_data = user_data;
  ctx->handshake_fail_id = 0;
  ctx->handshake_timeout_id = 0;
  return ctx;
}


static void
peer_handshake_context_free (struct PeerHandshakeContext* ctx)
{
  if (ctx->handshake_fail_id != 0)
    {
      g_source_remove (ctx->handshake_fail_id);
    }
  if (ctx->handshake_timeout_id != 0)
    {
      g_source_remove (ctx->handshake_timeout_id);
    }
  g_free (ctx);
}

/**** End of PeerHandshakeContext ****/


gboolean
dscuss_peer_is_handshaked (DscussPeer* peer)
{
  g_assert (peer != NULL);
  return peer->is_handshaked;
}


static gboolean
peer_handshake_failed (gpointer user_data)
{
  DscussPeer* peer = user_data;

  g_assert (peer != NULL);

  g_debug ("Failed to handshake with the peer '%s'",
           dscuss_peer_get_description (peer));

  peer->handshake_ctx->handshake_fail_id = 0;

  dscuss_free_non_null (peer->user, dscuss_user_free);
  peer->user = NULL;
  peer->handshake_ctx->callback (peer, FALSE, peer->handshake_ctx->user_data);
  dscuss_free_non_null (peer->connection, dscuss_connection_free);
  peer_handshake_context_free (peer->handshake_ctx);
  peer->handshake_ctx = NULL;

  return FALSE;
}


static void
peer_handshake_schedule_fail (DscussPeer* peer)
{
  if (peer->handshake_ctx->handshake_fail_id != 0)
    {
      g_source_remove (peer->handshake_ctx->handshake_fail_id);
    }
  peer->handshake_ctx->handshake_fail_id = g_idle_add_full (G_PRIORITY_HIGH,
                                                            peer_handshake_failed,
                                                            peer,
                                                            NULL);
}


static gboolean
peer_handshake_on_hello_received (DscussConnection* connection,
                                  DscussPacket* packet,
                                  gboolean result,
                                  gpointer user_data)
{
  DscussPeer* peer = user_data;
  gchar* payload = NULL;
  gsize payload_size = 0;
  DscussPayloadHello* pld_hello = NULL;
  GDateTime* datetime_now = NULL;
  gint64 ts_discrepancy = 0;
  gboolean handshake_result = FALSE;

  g_debug ("Received Hello from the peer '%s'",
           dscuss_peer_get_description (peer));

  if (!result)
    {
      g_debug ("Handshake error: failed to read Hello from connection '%s'",
               dscuss_connection_get_description (connection));
      goto out;
    }

  DscussPacketType type = dscuss_packet_get_type (packet);
  if (type != DSCUSS_PACKET_TYPE_HELLO)
    {
      g_warning ("Handshake error - protocol violation detected:"
                 " peer '%s' sent unexpected packet of type '%d'."
                 " Expected: %d (peer's user for handshaking)",
                 dscuss_peer_get_description (peer), type,
                 DSCUSS_PACKET_TYPE_HELLO);
      goto out;
    }

  if (!dscuss_packet_verify (packet,
                             dscuss_user_get_public_key (peer->user)))
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
  if (memcmp (dscuss_user_get_id (peer->user),
              dscuss_payload_hello_get_receiver_id (pld_hello),
              sizeof (DscussHash)))
    {
      g_warning ("Handshake error: wrong receiver ID: '%s'",
                 dscuss_crypto_hash_to_string (dscuss_payload_hello_get_receiver_id (pld_hello)));
      g_debug ("Expected receiver ID: '%s'",
                 dscuss_crypto_hash_to_string (dscuss_user_get_id (peer->user)));
      goto out;
    }

  datetime_now = g_date_time_new_now_utc ();
  ts_discrepancy = ABS (g_date_time_to_unix (datetime_now) -
                        g_date_time_to_unix (dscuss_payload_hello_get_datetime (pld_hello)));
  if (ts_discrepancy > DSCUSS_PEER_MAX_TIMESTAMP_DISCREPANCY)
    {
      g_warning ("Handshake error: timestamp discrepancy exceeds the limit:"
                 " %" G_GINT64_FORMAT, ts_discrepancy);
      goto out;
    }

  /* Finally handshake succeeded */
  peer->subscriptions = dscuss_subscriptions_copy (dscuss_payload_hello_get_subscriptions (pld_hello));
  peer->is_handshaked = TRUE;
  peer->handshake_ctx->callback (peer, TRUE, peer->handshake_ctx->user_data);
  peer_handshake_context_free (peer->handshake_ctx);
  peer->handshake_ctx = NULL;
  handshake_result = TRUE;

  /* Now we expect ordinary entities: messages, TBD: users and operations */
  g_hash_table_insert (peer->expected_types,
                       GINT_TO_POINTER (DSCUSS_PACKET_TYPE_MSG), NULL);

out:
  dscuss_free_non_null (packet, dscuss_packet_free);
  dscuss_free_non_null (datetime_now, g_date_time_unref);
  dscuss_free_non_null (pld_hello, dscuss_payload_hello_free);
  if (!handshake_result)
    peer_handshake_schedule_fail (peer);
  return FALSE;
}


static void
peer_handshake_on_hello_sent (DscussConnection* connection,
                              const DscussPacket* packet,
                              gboolean result,
                              gpointer user_data)
{
  DscussPeer* peer = user_data;

  dscuss_packet_free ((DscussPacket*) packet);

  if (!result)
    {
      g_warning ("Handshake error: failed to send hello to the peer '%s'",
                 dscuss_peer_get_description (peer));
      peer_handshake_schedule_fail (peer);
    }
  else
    {
      g_debug ("Hello successfully sent to the peer '%s'",
               dscuss_peer_get_description (peer));

    }
}


static gboolean
peer_handshake_send_hello (DscussPeer* peer)
{
  DscussPayloadHello* pld_hello = NULL;
  gchar* serialized_hello = NULL;
  gsize serialized_hello_len = 0;
  DscussPacket* hello_packet = NULL;
  gboolean result = FALSE;

  g_assert (peer != NULL);

  g_debug ("Trying to send Hello to the peer '%s'",
           dscuss_peer_get_description (peer));

  pld_hello = dscuss_payload_hello_new (dscuss_user_get_id (peer->handshake_ctx->user),
                                        peer->handshake_ctx->subscriptions);
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
  dscuss_packet_sign (hello_packet, peer->handshake_ctx->privkey);
  dscuss_connection_send (peer->connection,
                          hello_packet,
                          peer_handshake_on_hello_sent,
                          peer);
  result = TRUE;

out:
  dscuss_free_non_null (pld_hello, dscuss_payload_hello_free);
  return result;
}


static gboolean
peer_handshake_on_user_received (DscussConnection* connection,
                                 DscussPacket* packet,
                                 gboolean result,
                                 gpointer user_data)
{
  DscussPeer* peer = user_data;
  gchar* payload = NULL;
  gsize payload_size = 0;
  DscussUser* user = NULL;
  DscussUser* stored_user = NULL;
  gboolean handle_user_result = FALSE;

  g_debug ("Received user of the peer '%s'",
           dscuss_peer_get_description (peer));

  if (!result)
    {
      g_debug ("Handshake error: failed to read User from connection '%s'",
               dscuss_connection_get_description (connection));
      goto out;
    }

  DscussPacketType type = dscuss_packet_get_type (packet);
  if (type != DSCUSS_PACKET_TYPE_USER)
    {
      g_warning ("Handshake error - protocol violation detected:"
                 " peer '%s' sent unexpected packet of type '%d'."
                 " Expected: %d (peer's user for handshaking)",
                 dscuss_peer_get_description (peer), type,
                 DSCUSS_PACKET_TYPE_USER);
      goto out;
    }

  g_debug ("Handshaking: got user of the peer '%s'",
           dscuss_peer_get_description (peer));

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
  stored_user = dscuss_db_get_user (peer->handshake_ctx->dbh,
                                    dscuss_user_get_id (user));
  if (stored_user == NULL)
    {
      if (!dscuss_db_put_user (peer->handshake_ctx->dbh, user))
        {
          g_warning ("Handshake error:"
                     " failed to store the user '%s' of the peer '%s'",
                     dscuss_user_get_description (user),
                     dscuss_peer_get_description (peer));
          goto out;
        }
    }
  peer->user = user;

  /* Send our user in response. */
  if (!peer_handshake_send_hello (peer))
    {
      g_warning ("Handshake error: failed to send Hello to the peer");
      goto out;
    }
  dscuss_connection_set_receive_callback (peer->connection,
                                          peer_handshake_on_hello_received,
                                          peer);
  handle_user_result = TRUE;

out:
  dscuss_free_non_null (packet, dscuss_packet_free);
  dscuss_free_non_null (stored_user, dscuss_user_free);
  if (!handle_user_result)
    {
      dscuss_free_non_null (user, dscuss_user_free);
      peer->user = NULL;
      peer_handshake_schedule_fail (peer);
    }
  return handle_user_result;
}


static void
peer_handshake_on_user_sent (DscussConnection* connection,
                             const DscussPacket* packet,
                             gboolean result,
                             gpointer user_data)
{
  DscussPeer* peer = user_data;

  dscuss_packet_free ((DscussPacket*) packet);

  if (!result)
    {
      g_warning ("Handshake error: failed to send our user to the peer '%s'",
                 dscuss_peer_get_description (peer));
      peer_handshake_schedule_fail (peer);
    }
  else
    {
      g_debug ("Our user successfully sent to the peer '%s'",
               dscuss_peer_get_description (peer));
    }
}


static gboolean
peer_handshake_send_user (DscussPeer* peer)
{
  gchar* serialized_user = NULL;
  gsize serialized_user_len = 0;
  DscussPacket* user_packet = NULL;

  g_assert (peer != NULL);

  g_debug ("Trying to send out User to the peer '%s'",
           dscuss_peer_get_description (peer));

  if (!dscuss_user_serialize (peer->handshake_ctx->user,
                              &serialized_user,
                              &serialized_user_len))
    {
      g_warning ("Handshake error: failed to serialize the user '%s'",
                 dscuss_user_get_description (peer->handshake_ctx->user));
      return FALSE;
    }
  user_packet = dscuss_packet_new (DSCUSS_PACKET_TYPE_USER,
                                   serialized_user,
                                   serialized_user_len);
  g_free (serialized_user);
  /* This packet goes without signature. */
  dscuss_connection_send (peer->connection,
                          user_packet,
                          peer_handshake_on_user_sent,
                          peer);
  return TRUE;
}


void
dscuss_peer_handshake (DscussPeer* peer,
                       const DscussUser* user,
                       DscussPrivateKey* privkey,
                       GSList* subscriptions,
                       DscussDb* dbh,
                       DscussPeerHandshakeCallback callback,
                       gpointer user_data)
{

  g_assert (peer != NULL);
  g_assert (user != NULL);
  g_assert (privkey != NULL);
  g_assert (subscriptions != NULL);
  g_assert (callback != NULL);

  g_debug ("Starting handshake process with peer '%s'",
           dscuss_peer_get_description (peer));

  peer->handshake_ctx = peer_handshake_context_new (user,
                                                    privkey,
                                                    subscriptions,
                                                    dbh,
                                                    callback,
                                                    user_data);

  if (!peer_handshake_send_user (peer))
    {
      g_warning ("Handshake error: failed to send our user to the peer");
      peer_handshake_schedule_fail (peer);
      return;
    }

  /* Timeout for failing handshake */
  peer->handshake_ctx->handshake_timeout_id =
      g_timeout_add_seconds (DSCUSS_PEER_HANDSHAKE_TIMEOUT,
                             peer_handshake_failed,
                             peer);

  dscuss_connection_set_receive_callback (peer->connection,
                                          peer_handshake_on_user_received,
                                          peer);
}

/**** END OF HANDSHAKING *****************************************************/
