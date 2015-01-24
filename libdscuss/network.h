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
 * @file network.h  Defines API of the network subsystem.
 * @brief The network subsystem is responsible for establishing connections
 * with other peers.
 */


#ifndef DSCUSS_NETWORK_H
#define DSCUSS_NETWORK_H

#include <glib.h>
#include "connection.h"
#include "peer.h"
#include "user.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Callback used for notification about newly connected peers.
 *
 * @param conn        Connected peer.
 * @param user_data   The user data.
 */
typedef void (*DscussNewPeerCallback)(DscussPeer* peer,
                                      gpointer user_data);

/**
 * Initializes the network subsystem.
 *
 * Opens listening port and establishes connection with known peer addresses.
 *
 * @param addr_file          The path to the file to read addresses of other
 *                           peers from.
 * @param user               The user entity (for handshaking).
 * @param privkey            The private key of the user (for signing packets).
 * @param new_peer_callback  The function to be called when a new peer connects.
 * @param new_peer_data      Additional data to be passed to the callback.
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
gboolean
dscuss_network_init (const gchar* addr_filename,
                     const DscussUser* user,
                     const DscussPrivateKey* privkey,
                     DscussNewPeerCallback new_peer_callback,
                     gpointer new_peer_data);

/**
 * Destroys the network subsystem.
 *
 * Closes listening port, closes all network connections and frees allocated memory.
 */
void
dscuss_network_uninit ();


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_NETWORK_H */
