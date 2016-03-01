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

/**
 * @file handshake.h  Business logic for handshaking.
 */


#ifndef DSCUSS_HANDSHAKE_H
#define DSCUSS_HANDSHAKE_H

#include <glib.h>
#include "user.h"
#include "subscriptions.h"
#include "db.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for handshaking.
 */
typedef struct _DscussHandshakeHandle DscussHandshakeHandle;

/**
 * Callback used for notifying that handshaking is over.
 *
 * @param result               @c TRUE if the handshaking was successful,
 *                             @c FALSE otherwise.
 * @param peers_user           The user of the peer which we've handshaked with
 *                             or @c NULL if the result is negative.
 * @param peers_subscriptions  The list of topics the peer's user is subscribed
 *                             to or @c NULL if the result is negative.
 * @param user_data            The user data.
 */
typedef void (*DscussHandshakeCallback)(gboolean result,
                                        DscussUser* peers_user,
                                        GSList* peers_subscriptions,
                                        gpointer user_data);

/**
 * Start handshaking with other peer.
 *
 * @param connection          Network connection with the peer to handshake
 *                            with.
 * @param self                The user we are logged under.
 * @param self_privkey        The private key of the user (for signing packet).
 * @param self_subscriptions  List of the topics the user is subscribed to.
 * @param dbh                 Handle for the database connection.
 * @param callback            Function to call when handshaking is over.
 * @param user_data           User data to pass to the callback.
 *
 * @return  Newly created handle for handshaking. Handle will be automatically
 *          destroyed after invoking the callback.
 */
DscussHandshakeHandle*
dscuss_handshake_start (DscussConnection* connection,
                        const DscussUser* self,
                        DscussPrivateKey* self_privkey,
                        GSList* self_subscriptions,
                        DscussDb* dbh,
                        DscussHandshakeCallback callback,
                        gpointer user_data);

/**
 * Stop the handshaking process and destroy its handle.
 * Callback will not be called.
 *
 * @param handle  Handle for the handshaking process to cancel,
 */
void
dscuss_handshake_cancel (DscussHandshakeHandle* handle);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_HANDSHAKE_H */
