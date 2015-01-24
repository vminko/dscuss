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
#include "header.h"


#define DSCUSS_HEADER_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_HEADER_DESCRIPTION_MAX_LEN];


/**
 * Handle for a header.
 */
struct _DscussHeader
{
  /* Packet type in HBO */
  guint16 type;

  /* Size of the whole packet in HBO (in bytes) */
  guint16 size;
};

/**
 * RAW header struct.
 */
struct _DscussHeaderNBO
{
  /* Packet type in NBO */
  guint16 type;

  /* Size of the whole packet in NBO (in bytes) */
  guint16 size;
};


DscussHeader*
dscuss_header_new (void)
{
  DscussHeader* header = g_new0 (DscussHeader, 1);
  return header;
}


DscussHeader*
dscuss_header_new_full (guint16 type, guint16 size)
{
  DscussHeader* header = g_new0 (DscussHeader, 1);
  header->type = type;
  header->size = size;
  return header;
}


void
dscuss_header_free (DscussHeader* header)
{
  if (header == NULL)
    return;
  g_free (header);
}


DscussHeader*
dscuss_header_copy (const DscussHeader* header)
{
  return dscuss_header_new_full (dscuss_header_get_packet_type (header),
                                 dscuss_header_get_packet_size (header));
}


DscussHeader*
dscuss_header_deserialize (const gchar* raw_data)
{
  g_assert (raw_data != NULL);

  struct _DscussHeaderNBO* header_nbo = (struct _DscussHeaderNBO*) raw_data;
  DscussHeader* header = g_new0 (DscussHeader, 1);
  header->type = g_ntohs (header_nbo->type);
  header->size = g_ntohs (header_nbo->size);
  return header;
}


void
dscuss_header_serialize (const DscussHeader* header,
                         gchar* raw_data)
{
  g_assert (header != NULL);
  g_assert (raw_data != NULL);

  struct _DscussHeaderNBO* header_nbo = (struct _DscussHeaderNBO*) raw_data;
  header_nbo->type = g_htons (header->type);
  header_nbo->size = g_ntohs (header->size);
}


gsize
dscuss_header_get_size (void)
{
  return sizeof (struct _DscussHeaderNBO);
}


guint16
dscuss_header_get_packet_type (const DscussHeader* header)
{
  g_assert (header != NULL);
  return header->type;
}


gsize
dscuss_header_get_packet_size (const DscussHeader* header)
{
  g_assert (header != NULL);
  return header->size;
}

const gchar*
dscuss_header_get_description (const DscussHeader* header)
{
  g_assert (header != NULL);
  g_snprintf (description_buf, 
              DSCUSS_HEADER_DESCRIPTION_MAX_LEN,
              "type %d, size %" G_GSIZE_FORMAT,
              dscuss_header_get_packet_type (header),
              dscuss_header_get_packet_size (header));
  return description_buf;
}
