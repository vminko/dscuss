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

#include <string.h>
#include <glib.h>
#include "payload_announcement.h"
#include "util.h"


/*
 * Announcement packets are used for propagating new entities through the
 * network with low traffic overhead.
 * When user A sends this packet to user B, user A wants to let user B known
 * that user A has new entity, which may be interesting for user B.
 */
struct _DscussPayloadAnnouncement
{
  /**
   * Id of the entity to advertise.
   */
  DscussHash entity_id;

};


/**
 * RAW announcement struct. All fields are in NBO.
 */
struct _DscussPayloadAnnouncementNBO
{
  /**
   * Id of the entity to advertise.
   */
  DscussHash entity_id;

};


DscussPayloadAnnouncement*
dscuss_payload_announcement_new (const DscussHash* entity_id)
{
  DscussPayloadAnnouncement* pld_ann = NULL;

  g_assert (entity_id != NULL);

  pld_ann = g_new0 (DscussPayloadAnnouncement, 1);
  memcpy (&pld_ann->entity_id, entity_id, sizeof (DscussHash));

  return pld_ann;
}


void
dscuss_payload_announcement_free (DscussPayloadAnnouncement* pld_ann)
{
  if (pld_ann == NULL)
    return;
  g_free (pld_ann);
}


gboolean
dscuss_payload_announcement_serialize (const DscussPayloadAnnouncement* pld_ann,
                                        gchar** data,
                                        gsize* size)
{
  struct _DscussPayloadAnnouncementNBO* pld_ann_nbo = NULL;

  g_assert (pld_ann != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  *size = sizeof (struct _DscussPayloadAnnouncementNBO);
  pld_ann_nbo = g_new0 (struct _DscussPayloadAnnouncementNBO, 1);
  *data = (gchar*) pld_ann_nbo;
  memcpy (&pld_ann_nbo->entity_id, &pld_ann->entity_id, sizeof (DscussHash));

  return TRUE;
}


DscussPayloadAnnouncement*
dscuss_payload_announcement_deserialize (const gchar* data,
                                          gsize size)
{
  struct _DscussPayloadAnnouncementNBO* pld_ann_nbo = NULL;
  DscussPayloadAnnouncement* pld_ann = NULL;


  /* Validate raw data */
  if (size <= sizeof (struct _DscussPayloadAnnouncementNBO))
    {
      g_warning ("Size of the raw data is too small."
                 " Actual size: %" G_GSIZE_FORMAT
                 ", expected: > %" G_GSIZE_FORMAT,
                 size,  sizeof (struct _DscussPayloadAnnouncementNBO));
      return NULL;
    }

  g_assert (data != NULL);

  pld_ann_nbo = (struct _DscussPayloadAnnouncementNBO*) data;
  pld_ann = dscuss_payload_announcement_new (&pld_ann_nbo->entity_id);

  return pld_ann;
}


const DscussHash*
dscuss_payload_announcement_get_entity_id (const DscussPayloadAnnouncement* pld_ann)
{
  g_assert (pld_ann != NULL);
  return &pld_ann->entity_id;
}

