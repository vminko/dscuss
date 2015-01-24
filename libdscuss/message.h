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
 * @file entity.h  Dscuss message definition.
 * @brief Message is some information published by a user.
 */

#ifndef DSCUSS_MESSAGE_H
#define DSCUSS_MESSAGE_H

#include <glib.h>
#include "entity.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a message.
 */
typedef struct _DscussMessage DscussMessage;

/**
 * Creates a new message entity.
 *
 * @param content Plain text message content.
 *
 * @return newly created message entity.
 */
DscussMessage*
dscuss_message_new (const gchar* content);

/**
 * Destroys a message entity.
 * Frees all memory allocated by the entity.
 *
 * @ param msg Message to be destroyed.
 */
void
dscuss_message_free (DscussMessage* msg);

/**
 * Returns message content.
 *
 * @param msg a Message.
 *
 * @return Content of the message.
 */
const gchar*
dscuss_message_get_content (const DscussMessage* msg);

/**
 * Composes a one-line text description of a message.
 *
 * @param msg Message to compose description for.
 *
 * @return Text description of the message.
 */
const gchar*
dscuss_message_get_description (const DscussMessage* msg);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_MESSAGE_H */
