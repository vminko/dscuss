
/**
 * This file is part of Dscuss.
 * Copyright (C) 2015  Vitaly Minko
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
 * @file payload_hello.h  Payload of the packet for introducing peers during
 * handshake.
 */

#ifndef DSCUSS_PAYLOAD_HELLO_H
#define DSCUSS_PAYLOAD_HELLO_H

#include <glib.h>
#include "crypto_hash.h"


#ifdef __cplusplus
extern "C" {
#endif


/*
 * Hello packet is used for handshaking.
 * When user A sends this packet to user B, he/she:
 * 1. notifies user B about topics of A's interests;
 * 2. proves that the user A actually has the A's private key;
 */
typedef struct _DscussPayloadHello DscussPayloadHello;

/**
 * Creates new payload for a hello packet.
 *
 * @param id             ID of the user we are sending Hello to.
 * @param subscriptions  List of topics, which are interesting for the user.
 *
 * @return New payload.
 */
DscussPayloadHello*
dscuss_payload_hello_new (const DscussHash* receiver_id,
                          GSList* subscriptions);

/**
 * Destroys a hello payload.
 * Frees all memory allocated by the payload.
 *
 * @param hello  Payload to be destroyed.
 */
void
dscuss_payload_hello_free (DscussPayloadHello* hello);

/**
 * Convert a Hello payload to raw data, which can be transmitted via network.
 *
 * @param hello  Payload to serialize.
 * @param data   Where to store address of the serialized payload.
 * @param size   @a data size (output parameter).
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_payload_hello_serialize (const DscussPayloadHello* hello,
                                gchar** data,
                                gsize* size);

/**
 * Create a Hello payload from raw data.
 *
 * @param data  Raw data to parse.
 * @param size  Size of @a data.
 *
 * @return  A new Hello payload in case of success or @c NULL on error.
 */
DscussPayloadHello*
dscuss_payload_hello_deserialize (const gchar* data,
                                  gsize size);

/**
 * Returns ID of the receiver.
 *
 * @param hello  Payload to fetch ID from.
 *
 * @return  ID of the receiver.
 */
const DscussHash*
dscuss_payload_hello_get_receiver_id (const DscussPayloadHello* hello);

/**
 * Returns date and time when the payload was composed.
 *
 * @param hello  Payload to fetch date and time from.
 *
 * @return  Date and time when the payload was composed.
 */
GDateTime*
dscuss_payload_hello_get_datetime (const DscussPayloadHello* hello);

/**
 * Returns the list of the sender's subscriptions.
 *
 * @param hello  Payload to fetch subscriptions from.
 *
 * @return  Subscriptions of the sender.
 */
GSList*
dscuss_payload_hello_get_subscriptions (DscussPayloadHello* hello);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_PAYLOAD_HELLO_H */
