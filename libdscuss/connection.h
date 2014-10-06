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

/**
 * @file connection.h  Connection with a peer.
 * @brief Connection provides a convenient API for sending and receiving
 * Dscuss packets between peers.
 */


#ifndef DSCUSS_CONNECTION_H
#define DSCUSS_CONNECTION_H

#include <glib.h>
#include <gio/gio.h>
#include "entity.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Explains an error occurred during interaction via the socket connection.
 */
typedef enum
{
  /*
   * Connection was broken due to some foreign factor.
   */
  DSCUSS_CONNECTION_ERROR_BROKEN = 0,
  /*
   * This other side has violated the protocol.
   */
  DSCUSS_CONNECTION_ERROR_VIOLATION,

} DscussConnectionError;


/**
 * Handle for a network connection.
 */
typedef struct _DscussConnection DscussConnection;

/**
 * Callback returns result of a send operation.
 *
 * @param connection Connection via which we've been trying to send
 *                   the packet.
 * @param packet     Packet which we've been trying to send.
 * @param result     @c TRUE if the packet has been successfully sent,
 *                   @c FALSE otherwise.
 * @param user_data  The user data.
 */
typedef void (*DscussConnectionSendCallback)(DscussConnection* connection,
                                             const DscussPacket* packet,
                                             gboolean result,
                                             gpointer user_data);

/**
 * Callback used for notifying about incoming packets.
 *
 * @param connection Connection via which we've been trying to send
 *                   the packet.
 * @param packet     Received packet.
 * @param result     @c TRUE if the packet has been successfully sent,
 *                   @c FALSE otherwise.
 * @param user_data  The user data.
 */
typedef void (*DscussConnectionReceiveCallback)(DscussConnection* connection,
                                                const DscussPacket* packet,
                                                gboolean result,
                                                gpointer user_data);

/**
 * Creates a new connection.
 *
 * @param socket_connection Socket connection, which will be used for
 *                          receiving and transmitting data.
 * @param is_incoming       @c TRUE if @a socket_connection is incoming,
 *                          @c FALSE otherwise.
 *
 * @return New connection handle.
 */
DscussConnection*
dscuss_connection_new (GSocketConnection* socket_connection,
                       gboolean is_incoming);

/**
 * Frees all memory allocated by a connection. Closes the socket
 * connection in case it was open.
 *
 * @param connection A connection whose memory is going to be freed.
 */
void
dscuss_connection_free (DscussConnection* connection);

/**
 * Send a packet to the connected peer.
 *
 * @param connection Connection to send the packet via.
 * @param packet     Packet to send.
 * @param callback   Function to call with the result of the sending.
 * @param user_data  User data to pass to the callback.
 */
void
dscuss_connection_send (DscussConnection* connection,
                        const DscussPacket* packet,
                        DscussConnectionSendCallback callback,
                        gpointer user_data);

/**
 * Composes a one-line text description of a connection.
 *
 * @param connection Connection to compose description for.
 *
 * @return Text description of the connection.
 */
const gchar*
dscuss_connection_get_description (DscussConnection* connection);

/**
 * Sets callback for notification about incoming packets and starts reading
 * data from the connection.
 *
 * @param connection Connection to receive packets from.
 * @param callback   Function to call when new packet received.
 * @param user_data  User data to pass to the callback.
 */
void
dscuss_connection_set_receive_callback (DscussConnection* connection,
                                        DscussConnectionReceiveCallback callback,
                                        gpointer user_data);

/**
 * Shows if connection is incoming.
 *
 * @param connection Connection to get parameter value of.
 *
 * @return @c TRUE if connection is incoming, @c FALSE otherwise.
 */
gboolean
dscuss_connection_is_incoming (DscussConnection* connection);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CONNECTION_H */
