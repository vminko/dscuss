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

/**
 * @file peer.h  Internal API for a connected peer.
 * @brief Peer provides a high-level API for communication with other nodes:
 * sending/receiving entities, syncing, etc.  All peers are handled by core.
 * Once a new peer connection is established, the peer gets passed to the core.
 * Once a peer gets disconnected, the network subsystem is notified via @c
 * disconn_callback callback.
 */


#ifndef DSCUSS_PEER_H
#define DSCUSS_PEER_H

#include <glib.h>
#include "entity.h"
#include "connection.h"
#include "user.h"
#include "db.h"
#include "include/peer.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Explains why a peed was disconnected
 */
typedef enum
{
  /*
   * Connection was broken due to some foreign factor.
   */
  DSCUSS_PEER_DISCONNECT_REASON_BROKEN = 0,
  /*
   * We have intentionally closed the connection.
   */
  DSCUSS_PEER_DISCONNECT_REASON_CLOSED,
  /*
   * We have another connection with the same peer.
   */
  DSCUSS_PEER_DISCONNECT_REASON_DUPLICATE,
  /*
   * We have no common interests with this peer.
   */
  DSCUSS_PEER_DISCONNECT_REASON_NO_COMMON_INTERESTS,
  /*
   * This peer is banned.
   */
  DSCUSS_PEER_DISCONNECT_REASON_BANNED,
  /*
   * This peer has violated the protocol.
   */
  DSCUSS_PEER_DISCONNECT_REASON_VIOLATION,

} DscussPeerDisconnectReason;

/**
 * Callback used for notifying the network subsystem about disconnected peers.
 *
 * @param peer         Peer which was disconnected.
 * @param reason       Explains why the connection with the peer was terminated.
 * @param readon_data  Additional data which depends on reason type.
 * @param user_data    The user data.
 */
typedef void (*DscussPeerDisconnectCallback)(DscussPeer* peer,
                                             DscussPeerDisconnectReason reason,
                                             gpointer reason_data,
                                             gpointer user_data);

/**
 * Callback used for notifying that handshaking is over.
 * If handshake failed, connection with the peer is closed.
 * No more sending/receiving operations can be performed.
 *
 * @param peer         Peer which we've handshaked with.
 * @param result       @c TRUE if the handshaking was successful,
 *                     @c FALSE otherwise.
 * @param user_data    The user data.
 */
typedef void (*DscussPeerHandshakeCallback)(DscussPeer* peer,
                                            gboolean result,
                                            gpointer user_data);

/**
 * Callback returns result of a send operation.
 *
 * @param peer       Peer which we've been trying to send
 *                   the packet to.
 * @param entity     Entity which we've been trying to send.
 * @param result     @c TRUE if the entity has been successfully sent,
 *                   @c FALSE otherwise.
 * @param user_data  The user data.
 */
typedef void (*DscussPeerSendCallback)(DscussPeer* peer,
                                       const DscussEntity* entity,
                                       gboolean result,
                                       gpointer user_data);

/**
 * Callback used for notifying about incoming entities.
 *
 * @param peer       Peer to which we've been trying to send
 *                   the entity.
 * @param entity     Received entity.
 * @param result     @c TRUE if the entity has been successfully sent,
 *                   @c FALSE otherwise.
 * @param user_data  The user data.
 */
typedef void (*DscussPeerReceiveCallback)(DscussPeer* peer,
                                          DscussEntity* entity,
                                          gboolean result,
                                          gpointer user_data);

/**
 * Creates a new peer.
 *
 * @param socket_connection   Socket connection, which will be used for
 *                            receiving and transmitting data.
 * @param is_incoming         @c TRUE if @a socket_connection is incoming,
 *                            @c FALSE otherwise.
 * @param disconn_callback    Function to call when peer gets disconnected from
 *                            us.
 * @param disconn_data        User data to pass to disconn_callback.
 * @param handshake_callback  Function to call when we've successfully
 *                            handshaked with this peer.
 * @param handshake_data      User data to pass to handshake_callback.
 *
 * @return New connection handle.
 */
DscussPeer*
dscuss_peer_new (GSocketConnection* socket_connection,
                 gboolean is_incoming,
                 DscussPeerDisconnectCallback disconn_callback,
                 gpointer disconn_data);

/**
 * Frees all memory allocated by the peer. Closes connection with
 * the default reason (@c DSCUSS_PEER_DISCONNECT_REASON_CLOSED).
 *
 * @param peer  Peer to free.
 */
void
dscuss_peer_free (DscussPeer* peer);

/**
 * Frees a peer with specified reason.
 *
 * @param peer         Peer to free.
 * @param reason       Explains why we are terminating connection with this
 *                     peer.
 * @param reason_data  Reason-specific data.
 *                     Must be the duplicate peer if @a reason is
 *                     @c DSCUSS_PEER_DISCONNECT_REASON_DUPLICATE.
 */
void
dscuss_peer_free_full (DscussPeer* peer,
                       DscussPeerDisconnectReason reason,
                       gpointer reason_data);

/**
 * Send an entity to the connected peer.
 *
 * @param peer       Peer send the entity to.
 * @param entity     Entity to send.
 * @param privkey    The private key of the user (for signing packet).
 * @param callback   Function to call with the result of the sending, will
 *                   only be called if the entity was successfully inserted
 *                   in the outgoing queue.
 * @param user_data  User data to pass to the callback.
 *
 * @return @c TRUE if the entity is successfully queued for delivery
 *                 (the callback will be called),
 *         @c FALSE if the attempt to send the entity has failed
 *                  (the callback will not be called).
 */
gboolean
dscuss_peer_send (DscussPeer* peer,
                  DscussEntity* entity,
                  DscussPrivateKey* privkey,
                  DscussPeerSendCallback callback,
                  gpointer user_data);

/**
 * Sets callback for notification about incoming entities and starts reading
 * data from the peer.
 *
 * @param peer       Peer to receive entities from.
 * @param callback   Function to call when new entity received.
 * @param user_data  User data to pass to the callback.
 */
void
dscuss_peer_set_receive_callback (DscussPeer* peer,
                                  DscussPeerReceiveCallback callback,
                                  gpointer user_data);

/**
 * Request performing handshake with the peer.
 *
 * @param peer           Peer to handshake with.
 * @param user           The user we are logged under.
 * @param privkey        The private key of the user (for signing packets).
 * @param subscriptions  List of the topics the user is subscribed to.
 * @param dbh            Handle for the database connection.
 * @param callback       Function to call when handshaking is over.
 * @param user_data      User data to pass to the callback.
 */
void
dscuss_peer_handshake (DscussPeer* peer,
                       const DscussUser* user,
                       DscussPrivateKey* privkey,
                       GSList* subscriptions,
                       DscussDb* dbh,
                       DscussPeerHandshakeCallback callback,
                       gpointer user_data);

/**
 * Shows whether peer is handshaked.
 * Peer's user is unknown until peer is handshaked.
 *
 * @param peer  Peer to investigate.
 *
 * @return @c TRUE is we've already handshaked with this peer,
 *         @c FALSE otherwise.
 */
gboolean
dscuss_peer_is_handshaked (DscussPeer* peer);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_PEER_H */
