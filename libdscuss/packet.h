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
 * @file packet.h  Packet is a unit of raw data for communication between peers.
 */

#ifndef DSCUSS_PACKET_H
#define DSCUSS_PACKET_H

#include <glib.h>
#include "header.h"
#include "crypto.h"


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
   * Used for introducing users during handshake.
   */
  DSCUSS_PACKET_TYPE_HELLO,
  /*
   * Used for advertising new entities.
   */
  DSCUSS_PACKET_TYPE_ANNOUNCE,
  /*
   * Acknowledgment for an announcement.
   */
  DSCUSS_PACKET_TYPE_ACK,
  /*
   * Request for an entity.
   */
  DSCUSS_PACKET_TYPE_REQ,

  /*
   * Used for checking packet type validity.
   * Must be the last type in the list.
   * If you need to add a new type, add it above this one.
   */
  DSCUSS_PACKET_TYPE_LAST_TYPE,

} DscussPacketType;

/**
 * Handle for a packet.
 */
typedef struct _DscussPacket DscussPacket;

/**
 * Create new packet with no signature specified.
 * Such packet must be signed explicitly after creation.
 *
 * @param type          Packet type.
 * @param payload       Payload of the packet.
 * @param payload_size  Size of @a payload.
 *
 * @return  The newly created packet with empty signature.
 */
DscussPacket*
dscuss_packet_new (DscussPacketType type,
                   const gchar* payload,
                   gsize payload_size);

/**
 * Destroy a packet (free allocated memory).
 *
 * @param  The packet to free.
 */
void
dscuss_packet_free (DscussPacket* packet);

/**
 * Convert a packet to raw data, which can be transmitted via network.
 *
 * @param packet  Packet to serialize.
 * @param data    Where to store address of the serialized packet.
 * @param size    @a data size (output parameter).
 */
void
dscuss_packet_serialize (const DscussPacket* packet,
                         gchar** data,
                         gsize* size);

/**
 * Create packet from raw data and header.
 * Parse signature, check packet type and size, etc.
 *
 * @param header  Packet header (defines packet type and size).
 * @param data    Packet data (payload and signature).
 *
 * @return  A new packet in case of success or @c NULL on error.
 */
DscussPacket*
dscuss_packet_deserialize (const DscussHeader* header,
                           const gchar* data);

/**
 * Return packet type.
 *
 * @param packet  Packet to get type of.
 *
 * @return  Packet type.
 */
DscussPacketType
dscuss_packet_get_type (const DscussPacket* packet);

/**
 * Return full packet size (including header and body)
 *
 * @param packet  Packet to get size of.
 *
 * @return  Packet size.
 */
gsize
dscuss_packet_get_size (const DscussPacket* packet);

/**
 * Get payload of a packet.
 *
 * @param packet  Packet to get payload.
 * @param payload Where to store address of the payload.
 * @param size    @a payload size (output parameter).
 */
void
dscuss_packet_get_payload (const DscussPacket* packet,
                           const gchar** payload,
                           gsize* size);

/**
 * Get packet signature.
 *
 * @param packet  Packet to get signature from.
 *
 * @return  The signature of the packet.
 */
const struct DscussSignature*
dscuss_packet_get_signature (const DscussPacket* packet);

/**
 * Get length of the signature of the packet.
 *
 * @param packet  Packet to get length from.
 *
 * @return  The length of the signature of the packet.
 */
gsize
dscuss_packet_get_signature_length (const DscussPacket* packet);

/**
 * Compose a one-line text description of a packet.
 *
 * @param packet  Packet to compose description for.
 *
 * @return  Text description of the packet.
 */
const gchar*
dscuss_packet_get_description (const DscussPacket* packet);

/**
 * Sign a packet.
 *
 * @param packet   Packet to sign.
 * @param privkey  Private key to use for signing.
 */
void
dscuss_packet_sign (DscussPacket* packet,
                    DscussPrivateKey* privkey);

/**
 * Verify signature of a packet.
 *
 * @param packet  Packet to verify.
 * @param pubkey  Public key to use for verification.
 */
gboolean
dscuss_packet_verify (const DscussPacket* packet,
                      const DscussPublicKey* pubkey);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_PACKET_H */
