/**
 * This file is part of Dscuss.
 * Copyright (C) 2016  Vitaly Minko
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
 * @file payload_announcement.h  Payload of the packet for advertising new
 * entities.
 */

#ifndef DSCUSS_PAYLOAD_ANNOUNCEMENT_H
#define DSCUSS_PAYLOAD_ANNOUNCEMENT_H

#include <glib.h>
#include "crypto_hash.h"


#ifdef __cplusplus
extern "C" {
#endif


/*
 * Announcement packets are used for propagating new entities through the
 * network with low traffic overhead.
 * When user A sends this packet to user B, user A wants to let user B known
 * that user A has new entity, which may be interesting for user B.
 */
typedef struct _DscussPayloadAnnouncement DscussPayloadAnnouncement;

/**
 * Creates new payload for an announcement packet.
 *
 * @param id  ID of the entity to advertise.
 *
 * @return New payload.
 */
DscussPayloadAnnouncement*
dscuss_payload_announcement_new (const DscussHash* entity_id);

/**
 * Destroys an announcement payload.
 * Frees all memory allocated by the payload.
 *
 * @param pld_ann  Payload to be destroyed.
 */
void
dscuss_payload_announcement_free (DscussPayloadAnnouncement* pld_ann);

/**
 * Convert an Announcement payload to raw data, which can be transmitted via
 * network.
 *
 * @param pld_ann  Payload to serialize.
 * @param data     Where to store address of the serialized payload.
 * @param size     @a data size (output parameter).
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_payload_announcement_serialize (const DscussPayloadAnnouncement* pld_ann,
                                       gchar** data,
                                       gsize* size);

/**
 * Create an Announcement payload from raw data.
 *
 * @param data  Raw data to parse.
 * @param size  Size of @a data.
 *
 * @return  A new Announcement payload in case of success or @c NULL on error.
 */
DscussPayloadAnnouncement*
dscuss_payload_announcement_deserialize (const gchar* data,
                                         gsize size);

/**
 * Returns ID of the advertised entity.
 *
 * @param pld_ann  Payload to fetch entity ID from.
 *
 * @return  ID of the new entity.
 */
const DscussHash*
dscuss_payload_announcement_get_entity_id (const DscussPayloadAnnouncement* pld_ann);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_PAYLOAD_ANNOUNCEMENT_H */
