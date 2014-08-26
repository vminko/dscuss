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

#include <glib.h>
#include "packet.h"


#define DSCUSS_PACKET_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_PACKET_DESCRIPTION_MAX_LEN];


/**
 * Packet is a unit of raw data for communication between peers.
 * All packet data must be stored in network byte order.
 */
struct _DscussPacket
{
  /* Every packet must start with a header */
  DscussPacketHeader header;

  /* Below may be located packet body,
   * which depends on header.type
   */
};


DscussPacketType
dscuss_packet_get_type (const DscussPacket* packet)
{
  g_assert (packet);
  return g_ntohs (packet->header.type);
}


gssize
dscuss_packet_get_size (const DscussPacket* packet)
{
  g_assert (packet);
  return g_ntohs (packet->header.size);
}


const gchar*
dscuss_packet_get_description (const DscussPacket* packet)
{
  g_assert (packet);
  g_snprintf (description_buf, 
              DSCUSS_PACKET_DESCRIPTION_MAX_LEN,
              "type %d, size %ld",
              dscuss_packet_get_type (packet),
              dscuss_packet_get_size (packet));
  return description_buf;
}

