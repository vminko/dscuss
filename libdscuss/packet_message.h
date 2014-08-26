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
 * @file packet_message.h  Packet encapsulating a message entity.
 */

#ifndef DSCUSS_PACKET_MESSAGE_H
#define DSCUSS_PACKET_MESSAGE_H

#include <glib.h>
#include "packet.h"
#include "message.h"


#ifdef __cplusplus
extern "C" {
#endif


/*
 * This packet is used for transferring message entities between peers.
 */
typedef struct _DscussPacketMessage DscussPacketMessage;

/**
 * Creates new packet encapsulating a message entity.
 *
 * @param msg  Message to encapsulate in the packet.
 *
 * @return New packet.
 */
DscussPacketMessage*
dscuss_packet_message_new (const DscussMessage* msg);

/**
 * Extracts message from message packet.
 *
 * @param pckt_msg  Packet to extract message from.
 *
 * @return Extracted message entity.
 */
DscussMessage*
dscuss_packet_message_to_message (const DscussPacketMessage* pckt_msg);

#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_PACKET_MESSAGE_H */
