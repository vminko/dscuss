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

#include <string.h>
#include <glib.h>
#include "packet_message.h"


/**
 * This packet is used for transferring message entities between peers.
 * All data in a packet must be stored in network byte order.
 * Can be freed by g_free (packet)
 */
struct _DscussPacketMessage
{
  /* Packet type will be DSCUSS_PACKET_TYPE_MSG */
  DscussPacketHeader header;

  /* After this struct, the remaining bytes are the actual message in
   * plaintext */
};


DscussPacketMessage*
dscuss_packet_message_new (const DscussMessage* msg)
{
  DscussPacketMessage* pckt_msg = NULL;

  const gchar* cont = dscuss_message_get_content (msg);
  gsize size = sizeof (DscussPacketHeader) + strlen (cont) + 1;
  pckt_msg = g_malloc (size);
  pckt_msg->header.type = g_htons (DSCUSS_PACKET_TYPE_MSG);
  pckt_msg->header.size = g_htons (size);
  g_stpcpy ((char*)pckt_msg + sizeof (DscussPacketHeader), cont);

  return pckt_msg;
}


DscussMessage*
dscuss_packet_message_to_message (const DscussPacketMessage* pckt_msg)
{
  g_assert (pckt_msg);

  gssize size = g_ntohs (pckt_msg->header.size);
  gssize type = g_ntohs (pckt_msg->header.type);

  g_assert (type == DSCUSS_PACKET_TYPE_MSG);
  g_assert (*((char*)pckt_msg + size) == '\0');

  return dscuss_message_new ((char*)pckt_msg + sizeof (DscussPacketHeader));
}
