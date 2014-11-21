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
 * @file header.h  Dscuss header definition.
 * @brief Header is what every packet starts with.
 */

#ifndef DSCUSS_HEADER_H
#define DSCUSS_HEADER_H

#include <glib.h>


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a header.
 */
typedef struct _DscussHeader DscussHeader;

/**
 * Create a new header using default type and size values.
 *
 * @return A new header.
 */
DscussHeader*
dscuss_header_new (void);

/**
 * Create a new header using specified type and size values.
 *
 * @param type  Packet type.
 * @param size  Packet size.
 *
 * @return A new header.
 */
DscussHeader*
dscuss_header_new_full (guint16 type, guint16 size);

/**
 * Destroy a header (free allocated memory).
 *
 * @param header  The header to free.
 */
void
dscuss_header_free (DscussHeader* header);

/**
 * Copy a header.
 *
 * @param header  The header to copy.
 *
 * @return A new header, which is identical to @a header.
 */
DscussHeader*
dscuss_header_copy (const DscussHeader* header);

/**
 * Create a header from raw data.
 *
 * @param raw_data  Raw data to parse
 *                  (must be dscuss_header_get_size () bytes in size).
 *
 * @return A new header.
 */
DscussHeader*
dscuss_header_deserialize (const gchar* raw_data);

/**
 * Convert a header to raw data, which can be transmitted via network.
 *
 * @param header    Header to serialize.
 * @param raw_data  Where to write serialized header (raw header in NBO).
 *                  It's always dscuss_header_get_size () bytes in size.
 */
void
dscuss_header_serialize (const DscussHeader* header,
                         gchar* raw_data);

/**
 * Get size of a header (which is constant).
 *
 * @return Header size in bytes.
 */
gssize
dscuss_header_get_size (void);

/**
 * Get packet type from a header.
 *
 * @param header  Header to get type from.
 *
 * @return Packet type in HBO.
 */
guint16
dscuss_header_get_packet_type (const DscussHeader* header);

/**
 * Get full packet size (including header and body) from a header.
 *
 * @param header  Header to get size from.
 *
 * @return Packet size in HBO.
 */
gssize
dscuss_header_get_packet_size (const DscussHeader* header);

/**
 * Compose a one-line text description of a header.
 *
 * @param header  Header to compose description for.
 *
 * @return  Text description of the header.
 */
const gchar*
dscuss_header_get_description (const DscussHeader* header);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_HEADER_H */
