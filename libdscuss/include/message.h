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
 * @file include/message.h  Dscuss message definition.
 * @brief Message is some text information published by a user.
 */

#ifndef DSCUSS_INCLUDE_MESSAGE_H
#define DSCUSS_INCLUDE_MESSAGE_H

#include <glib.h>
#include "topic.h"
#include "crypto_hash.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a message.
 */
typedef struct _DscussMessage DscussMessage;

/**
 * Creates a new thread.
 * Date and time will be current date and time.
 *
 * @param topic    Topic the thread will belong to.
 * @param subject  Subject of the message.
 * @param text     Plain next message content.
 *
 * @return  Newly created message entity.
 */
DscussMessage*
dscuss_message_new_thread (DscussTopic* topic,
                           const gchar* subject,
                           const gchar* text);

/**
 * Creates a reply to another message.
 * Date and time will be current date and time.
 *
 * @param parent_id  ID of the parent message.
 * @param subject    Subject of the message.
 * @param text       Plain next message content.
 *
 * @return  Newly created message entity.
 */
DscussMessage*
dscuss_message_new_reply (const DscussHash* parent_id,
                          const gchar* subject,
                          const gchar* text);

/**
 * Destroys a message entity.
 * Frees all memory allocated by the entity.
 *
 * @param msg  Message to be destroyed.
 */
void
dscuss_message_free (DscussMessage* msg);

/**
 * Composes a one-line text description of a message.
 *
 * @param msg Message to compose description for.
 *
 * @return  Text description of the message.
 */
const gchar*
dscuss_message_get_description (const DscussMessage* msg);

/**
 * Returns ID of a mesasge.
 *
 * @param msg  Message to get ID of.
 *
 * @return  ID of the message.
 */
const DscussHash*
dscuss_message_get_id (const DscussMessage* msg);

/**
 * Returns topic of a message.
 *
 * @param msg  Message to get topic from.
 *
 * @return  Topic of the message.
 */
const DscussTopic*
dscuss_message_get_topic (const DscussMessage* msg);

/**
 * Returns subject of a message.
 *
 * @param msg  Message to get subject from.
 *
 * @return  Subject of the message.
 */
const gchar*
dscuss_message_get_subject (const DscussMessage* msg);

/**
 * Returns content of a message.
 *
 * @param msg  Message to get content from.
 *
 * @return  Content of the message.
 */
const gchar*
dscuss_message_get_content (const DscussMessage* msg);

/**
 * Returns date and time when the message was written.
 *
 * @param msg  Message to get creation date and time of..
 *
 * @return  Date and time when the message was created.
 */
GDateTime*
dscuss_message_get_datetime (DscussMessage* msg);

/**
 * Returns author ID of a mesasge.
 *
 * @param msg  Message to get author ID of.
 *
 * @return  ID of the author of the message.
 */
const DscussHash*
dscuss_message_get_author_id (const DscussMessage* msg);

/**
 * Returns author ID of a mesasge.
 *
 * @param msg  Message to get author ID of.
 *
 * @return  ID of the author of the message.
 */
const DscussHash*
dscuss_message_get_parent_id (const DscussMessage* msg);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_INCLUDE_MESSAGE_H */
