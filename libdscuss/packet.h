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
 * @file packet.h  Packet is a blob of data for reading from and writing to
 *                 socket connections.
 * @brief Provides routines for converting entities to raw data and vice
 *        versa.
 */

#ifndef DSCUSS_PACKET_H
#define DSCUSS_PACKET_H

#include <glib.h>


#ifdef __cplusplus
extern "C" {
#endif

/**
 * Maximum size of a packet.
 */

#define DSCUSS_PACKET_MAX_SIZE 65535

/**
 * Dscuss packet types.
 */
typedef enum
{
  /*
   * Encapsulates a user entity.
   */
  DSCUSS_PACKET_TYPE_USER = 0,
  /*
   * Encapsulates a message entity.
   */
  DSCUSS_PACKET_TYPE_MSG,
  /*
   * Encapsulates an operation entity.
   */
  DSCUSS_PACKET_TYPE_OPER,
  /*
   * TBD
   */
  DSCUSS_ENTITY_TYPE_HELLO,
  DSCUSS_ENTITY_TYPE_GET,
  DSCUSS_PACKET_TYPE_END,

} DscussPacketType;

/**
 * Handle for a packet.
 */
typedef struct _DscussPacket DscussPacket;

/**
 * Header is what every packet stars with.
 */
typedef struct
{
  /* A DscussPacketType in NBO */
  guint16 type;

  /* Size of the whole packet in NBO (in bytes) */
  guint16 size;
} DscussPacketHeader;

/**
 * Returns packet type in HBO.
 *
 * @param packet  Packet to get type of.
 *
 * @return Packet type.
 */
DscussPacketType
dscuss_packet_get_type (const DscussPacket* packet);

/**
 * Returns full packet size (including header and body) 
 *
 * @param packet  Packet to get size of.
 *
 * @return Packet size in HBO.
 */
gssize
dscuss_packet_get_size (const DscussPacket* packet);

/**
 * Composes a one-line text description of a packet.
 *
 * @param packet  Packet to compose description for.
 *
 * @return Text description of the packet.
 */
const gchar*
dscuss_packet_get_description (const DscussPacket* packet);

/**
 * Use g_free to destroy a packet.
 */


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_PACKET_H */
