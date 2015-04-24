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
 * @file api.h  Defines the API of libdiscuss.
 * @brief Contains core subroutines, which should be moved to a separate file.
 */


#ifndef DSCUSS_API_H
#define DSCUSS_API_H

#include <glib.h>
#include "message.h"
#include "user.h"

#ifdef __cplusplus
extern "C" {
#endif


/* TBD */
typedef gpointer DscussOperation;


/**
 * Callback used for notification that registration is finished.
 *
 * @param result      The result of the registration (@c TRUE if success).
 * @param user_data   The user data.
 */
typedef void (*DscussRegisterCallback)(gboolean result,
                                       gpointer user_data);

/**
 * Callback used for notification about incoming messages.
 *
 * @param msg         The message received.
 * @param user_data   The user data.
 */
typedef void (*DscussNewMessageCallback)(DscussMessage* msg,
                                         gpointer user_data);

/**
 * Callback used for notification about incoming user entities.
 *
 * @param user        The user received.
 * @param user_data   The user data.
 */
typedef void (*DscussNewUserCallback)(DscussUser* user,
                                      gpointer user_data);

/**
 * Callback used for notification about incoming operations.
 *
 * @param operation   The operation received.
 * @param user_data   The user data.
 */
typedef void (*DscussNewOperationCallback)(DscussOperation* oper,
                                           gpointer user_data);

/**
 * Initializes the library.
 *
 * Initializes all the subsystems.
 *
 * @param data_dir      Path to the directory containing data files.
 *                      If @c NULL, the default directory will be used
 *                      (@c ~/.dscuss).
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_init (const gchar* data_dir);

/**
 * Uninitializes the library.
 *
 * Disconnects from other peers, closes database connection and frees allocated
 * memory.
 */
void
dscuss_uninit (void);

/**
 * Performs a single iteration for the event loop.
 *
 * Processes a pending event if any.
 */
void
dscuss_iterate (void);

/**
 * Register a new user.
 *
 * Creates private key for the user, find proof-of-work, stores user's profile
 * in the user's database.
 *
 * @param nickname   Nickname of the new user.
 * @param info       Additional information about the new user (may be
 *                   @c NULL).
 * @param callback   The function to be called when registration is
 *                   finished.
 * @param user_data  Additional data to be passed to the @a callback.
 *
 * @return @c TRUE if registration started successfully, or @c FALSE on error.
 */
gboolean
dscuss_register (const gchar* nickname,
                 const gchar* info,
                 DscussRegisterCallback callback,
                 gpointer user_data);

/**
 * Login into the network as user @a nickname.
 *
 * @param nickname      Username to login under.
 * @param msg_callback  The function to be called when a new message is
 *                      received.
 * @param msg_data      Additional data to be passed to the @a msg_callback.
 * @param user_callback The function to be called when a new user is
 *                      received.
 * @param user_data     Additional data to be passed to the @a user_callback.
 * @param oper_callback The function to be called when a new operation is
 *                      received.
 * @param oper_data     Additional data to be passed to the @a oper_callback.
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_login (const gchar* nickname,
              DscussNewMessageCallback msg_callback,
              gpointer msg_data,
              DscussNewUserCallback user_callback,
              gpointer user_data,
              DscussNewOperationCallback oper_callback,
              gpointer oper_data);

/**
 * Logout from the network.
 */
void
dscuss_logout (void);

/**
 * Show whether some user is logged in into the network.
 *
 * @return @c TRUE if a user is logged in, @c FALSE otherwise.
 */
gboolean
dscuss_is_logged_in (void);

/**
 * Returns list of connected peers. Should be called only after logging.
 *
 * @return  List of connected peers.
 */
const GSList*
dscuss_get_peers (void);

/**
 * Send a message to the network.
 *
 * Actually the message will be just copied the outgoing buffer. So it does not
 * guarantee that the message has been sent to a single peer when the function
 * returns.
 *
 * TBD: release version should have some sort of notification that the message
 * has been successfully sent to this or that peer.
 *
 * @param msg  Message to send.
 */
void
dscuss_send_message (DscussMessage* msg);

/**
 * Send a user to the network.
 *
 * Just like the @c dscuss_send_message, it does not guarantee that the
 * user has been sent to a single peer when the function returns.
 *
 * @param msg  Message to send.
 */
void
dscuss_send_user (DscussUser* user);

/**
 * Send an operation to the network.
 *
 * Just like the @c dscuss_send_message, it does not guarantee that the
 * operation has been sent to a single peer when the function returns.
 *
 * @param msg  Message to send.
 */
void
dscuss_send_operation (DscussOperation* oper);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_API_H */
