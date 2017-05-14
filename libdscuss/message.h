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
 * @file message.h  Internal API for message entity.
 * @brief Message is some text information published by a user.
 */

#ifndef DSCUSS_MESSAGE_H
#define DSCUSS_MESSAGE_H

#include <glib.h>
#include "crypto.h"
#include "entity.h"
#include "include/message.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Creates a new message entity with all fields supplied.
 * Either @a topic or @a parent_id must be @c NULL.
 *
 * @param topic          Topic the message will belong to.
 * @param parent_id      ID of the parent message.
 * @param subject        Subject of the message.
 * @param text           Plain next message content.
 * @param author_id      ID of the author of the message.
 * @param datetime       Date and time when the message was written.
 * @param signature      Signature of the message.
 * @param signature_len  Length of the @c signature.
 *
 * @return  Newly created message entity or
 *          @c NULL in case of incorrect parameters.
 */
DscussMessage*
dscuss_message_new_full (DscussTopic* topic,
                         const DscussHash* parent_id,
                         const gchar* subject,
                         const gchar* text,
                         const DscussHash* author_id,
                         GDateTime* datetime,
                         const struct DscussSignature* signature,
                         gsize signature_len);

/**
 * Convert a message to raw data, which can be transmitted via network.
 *
 * @param msg   Message to serialize.
 * @param data  Where to store address of the serialized mesage.
 * @param size  @a data size (output parameter).
 *
 * @return  @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_message_serialize (const DscussMessage* msg,
                          gchar** data,
                          gsize* size);

/**
 * Create message from raw data.
 *
 * @param data  Raw data to parse.
 * @param size  Size of @a data.
 *
 * @return  A new message in case of success or @c NULL on error.
 */
DscussMessage*
dscuss_message_deserialize (const gchar* data,
                            gsize size);

/**
 * Verify signature of the message.
 *
 * @param msg         Message to verify.
 * @param pubkey      Public key of the message author.
 *
 * @return  @c TRUE if signature is valid, or @c FALSE on error.
 */
gboolean
dscuss_message_verify_signature (const DscussMessage* msg,
                                 const DscussPublicKey* pubkey);

/**
 * Returns signature the message.
 *
 * @param msg  Message to get signature of.
 *
 * @return  Signature of the message.
 */
const struct DscussSignature*
dscuss_message_get_signature (const DscussMessage* msg);

/**
 * Returns length of the signature of the message.
 *
 * @param msg  Message to get length from.
 *
 * @return  The length of the signature of the message.
 */
gsize
dscuss_message_get_signature_length (const DscussMessage* msg);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_MESSAGE_H */
