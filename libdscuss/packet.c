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
#include "packet.h"


#define DSCUSS_PACKET_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_PACKET_DESCRIPTION_MAX_LEN];


/**
 * Handle for a packet.
 */
struct _DscussPacket
{
  /* Every packet must start with a header */
  DscussHeader* header;

  /* Data which depends on the packet type. */
  gchar* payload;

  /* Every packet ends with a signature of header+payload */
  struct DscussSignature signature;

};

DscussPacket*
dscuss_packet_new (DscussPacketType type,
                   const gchar* payload,
                   gssize payload_size)
{
  gssize packet_size = 0;

  g_assert (payload != NULL);
  g_assert (payload_size >= 0);

  DscussPacket* packet = g_new0 (DscussPacket, 1);
  packet_size = dscuss_header_get_size ()
              + payload_size
              + sizeof (struct DscussSignature);
  packet->header = dscuss_header_new_full (type, packet_size);
  packet->payload = g_malloc0 (payload_size);
  memcpy (packet->payload, payload, payload_size);
  return packet;
}


DscussPacket*
dscuss_packet_full (DscussPacketType type,
                    const gchar* payload,
                    gssize payload_size,
                    const struct DscussSignature* signature)
{
  g_assert (payload != NULL);
  g_assert (payload_size >= 0);
  g_assert (signature != NULL);

  DscussPacket* packet = dscuss_packet_new (type, payload, payload_size);
  memcpy (&packet->signature,
          signature,
          sizeof (struct DscussSignature));
  return packet;
}


void
dscuss_packet_free (DscussPacket* packet)
{
  g_assert (packet != NULL);
  dscuss_header_free (packet->header);
  g_free (packet->payload);
  g_free (packet);
}


void
dscuss_packet_serialize (const DscussPacket* packet,
                         gchar** data,
                         gssize* size)
{
  gchar* digest = NULL;
  gssize payload_size = 0;

  g_assert (packet != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  /* Copy header */
  digest = g_malloc0 (dscuss_header_get_packet_size (packet->header));
  *data = digest;
  dscuss_header_serialize (packet->header, digest);
  digest += dscuss_header_get_size ();

  /* Copy payload */
  payload_size = dscuss_header_get_packet_size (packet->header)
               - dscuss_header_get_size ()
               - sizeof (struct DscussSignature);
  memcpy (digest,
          packet->payload,
          payload_size);
  digest += payload_size;

  /* Copy signature */
  memcpy (digest,
          &packet->signature,
          sizeof (struct DscussSignature));

  *size = dscuss_header_get_packet_size (packet->header);
}


DscussPacket*
dscuss_packet_deserialize (const DscussHeader* header,
                           const gchar* data)
{
  g_assert (header != NULL);
  g_assert (data != NULL);

  if (dscuss_header_get_packet_size (header) <=
      dscuss_header_get_size () + sizeof (struct DscussSignature))
    {
      g_warning ("Packet size is too small: '%" G_GSSIZE_FORMAT "'",
                 dscuss_header_get_packet_size (header));
      return NULL;
    }

  if (dscuss_header_get_packet_type (header) >= DSCUSS_PACKET_TYPE_LAST_TYPE)
    {
      g_warning ("Invalid packet type: '%u'",
                 dscuss_header_get_packet_type (header));
      return NULL;
    }

  DscussPacket* packet = g_new0 (DscussPacket, 1);
  packet->header = dscuss_header_copy (header);
  gssize data_size = dscuss_header_get_packet_size (header)
                   - dscuss_header_get_size ();
  gssize payload_size = data_size - sizeof (struct DscussSignature);
  packet->payload = g_malloc0 (payload_size);
  memcpy (packet->payload, data, payload_size);
  memcpy (&packet->signature,
          data + payload_size,
          sizeof (struct DscussSignature));

  return packet;
}


DscussPacketType
dscuss_packet_get_type (const DscussPacket* packet)
{
  g_assert (packet != NULL);
  return dscuss_header_get_packet_type (packet->header);
}


gssize
dscuss_packet_get_size (const DscussPacket* packet)
{
  g_assert (packet != NULL);
  return dscuss_header_get_packet_size (packet->header);
}


void
dscuss_packet_get_payload (const DscussPacket* packet,
                           const gchar** payload,
                           gssize* size)
{
  g_assert (packet != NULL);
  g_assert (payload != NULL);
  g_assert (size != NULL);

  *payload = packet->payload;
  *size = dscuss_header_get_packet_size (packet->header)
        - dscuss_header_get_size ()
        - sizeof (struct DscussSignature);
}


const struct DscussSignature*
dscuss_packet_get_signature (const DscussPacket* packet)
{
  g_assert (packet != NULL);
  return &packet->signature;
}


const gchar*
dscuss_packet_get_description (const DscussPacket* packet)
{
  g_assert (packet != NULL);
  g_snprintf (description_buf, 
              DSCUSS_PACKET_DESCRIPTION_MAX_LEN,
              "type %d, size %" G_GSSIZE_FORMAT,
              dscuss_packet_get_type (packet),
              dscuss_packet_get_size (packet));
  return description_buf;
}


void
dscuss_packet_sign (DscussPacket* packet,
                    DscussPrivateKey* privkey)
{
  gchar* digest = NULL;
  gssize digest_len = 0;

  g_assert (packet != NULL);
  g_assert (privkey != NULL);

  dscuss_packet_serialize (packet,
                           &digest,
                           &digest_len);
  digest_len -= sizeof (struct DscussSignature);
  dscuss_crypto_ecc_sign (digest,
                          digest_len,
                          privkey,
                          (struct DscussSignature*) digest + digest_len);
  g_free (digest);
}


gboolean
dscuss_packet_verify (const DscussPacket* packet,
                      const DscussPublicKey* pubkey)
{
  gchar* digest = NULL;
  gssize digest_len = 0;
  gboolean res = FALSE;

  g_assert (packet != NULL);
  g_assert (pubkey != NULL);

  dscuss_packet_serialize (packet,
                           &digest,
                           &digest_len);
  digest_len -= sizeof (struct DscussSignature);
  res = dscuss_crypto_ecc_verify (digest,
                                  digest_len,
                                  pubkey,
                                  (const struct DscussSignature* ) digest + digest_len);
  g_free (digest);

  return res;
}

