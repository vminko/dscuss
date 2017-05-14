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
 * @file db.h  Defines API of the database subsystem.
 */


#ifndef DSCUSS_DB_H
#define DSCUSS_DB_H

#include <glib.h>
#include <sqlite3.h>
#include "user.h"
#include "message.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a database.
 */
typedef sqlite3 DscussDb;

/**
 * Function called to iterate over messages.
 *
 * @param success    @c TRUE if the message was successfully obtained,
 *                   @c FALSE otherwise.
 * @param msg        The message this notification is about,
 *                   @c NULL indicates end of list.
 * @param user_data  The user data.
 */
typedef void (*DscussDbIterateMessageCallback)(gboolean success,
                                               DscussMessage* msg,
                                               gpointer user_data);

/**
 * Opens connection with the database. Creates a new database if it does not
 * exist.
 *
 * @param filename  Filename used for the database.
 *
 * @return  Handle for the database in case of success, or @c NULL on error.
 */
DscussDb*
dscuss_db_open (const gchar* filename);

/**
 * Closes connection with the database. Frees allocated memory.
 *
 * @param dbh   Database to close.
 */
void
dscuss_db_close (DscussDb* dbh);

/**
 * Store a user in the database.
 *
 * @param dbh   Database handle.
 * @param user  The user to store.
 *
 * @return  @c TRUE on success, @c FALSE on error.
 */
gboolean
dscuss_db_put_user (DscussDb* dbh, DscussUser* user);

/**
 * Fetch a user from the database.
 *
 * @param id  Hash of the user's public key.
 *
 * @return  The fetched user
 *          or @c NULL if there is no such user in the database.
 */
DscussUser*
dscuss_db_get_user (DscussDb* dbh, const DscussHash* id);

/**
 * Store a message in the database.
 *
 * @param dbh   Database handle.
 * @param msg   The message to store.
 *
 * @return  @c TRUE on success, @c FALSE on error.
 */
gboolean
dscuss_db_put_message (DscussDb* dbh, DscussMessage* msg);

/**
 * Get a message by ID from the database.
 *
 * @param dbh     Database handle.
 * @param msg_id  ID of the message to fetch.
 *
 * @return  The fetched message or @c NULL in case of a failure
 *          (no such message in the database or internal error).
 */
DscussMessage*
dscuss_db_get_message (DscussDb* dbh,
                       const DscussHash* id);

/**
 * Fetch all root messages sorted by timestamp (from newest to oldest).
 * Root messages is a message with empty In_reply_to.
 *
 * @param dbh        Database handle.
 * @param callback   Function to call for each root message.
 * @param user_data  Data to be passed to the @a callback.
 */
void
dscuss_db_get_root_messages (DscussDb* dbh,
                             DscussDbIterateMessageCallback callback,
                             gpointer user_data);

/**
 * Fetch all replies to a parent message.
 * Replies will be sorted by timestamp (from newest to oldest).
 *
 * @param dbh        Database handle.
 * @param parent_id  ID of the parent message.
 * @param callback   Function to call for each root message.
 * @param user_data  Data to be passed to the @a callback.
 */
void
dscuss_db_get_message_replies (DscussDb* dbh,
                               const DscussHash* parent_id,
                               DscussDbIterateMessageCallback callback,
                               gpointer user_data);

/**
 * Test whether DB contains entity with the given ID.
 *
 * @param dbh        Database handle.
 * @param parent_id  ID of the entity to look for.
 */
gboolean
dscuss_db_has_entity (DscussDb* dbh, const DscussHash* id);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_DB_H */
