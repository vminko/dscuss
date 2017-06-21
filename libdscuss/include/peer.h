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
 * @file include/peer.h  A peer connected to us.
 * @brief Peer provides a high-level API for communication with other nodes:
 * sending/receiving entities, syncing, etc.
 */


#ifndef DSCUSS_INCLUDE_PEER_H
#define DSCUSS_INCLUDE_PEER_H

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a peer.
 */
typedef struct _DscussPeer DscussPeer;

/**
 * Composes a one-line text description of a peer.
 *
 * @param peer  Peer to compose description for.
 *
 * @return Text description of the peer.
 */
const gchar*
dscuss_peer_get_description (DscussPeer* peer);

/**
 * Composes a one-line text description of a peer's connection
 * (host, port).
 *
 * @param peer  Peer to compose description for.
 *
 * @return Text description of the peer's connection.
 */
const gchar*
dscuss_peer_get_connecton_description (DscussPeer* peer);

/**
 * Returns peer's user. Should be called after handshaking.
 *
 * @param peer  Peer to fetch user from.
 *
 * @return Peer's user if peer is handshaked or @c NULL otherwise.
 */
const DscussUser*
dscuss_peer_get_user (const DscussPeer* peer);

/**
 * Returns peer's subscriptions. Should be called after handshaking.
 *
 * @param peer  Peer to fetch subscriptions from.
 *
 * @return Peer's subscriptions if peer is handshaked or @c NULL otherwise.
 */
const GSList*
dscuss_peer_get_subscriptions (const DscussPeer* peer);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_INCLUDE_PEER_H */
