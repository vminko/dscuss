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
#include "payload_advertisement.h"
#include "util.h"


/*
 * Advertisement packets are used for propagating new entities through the
 * network with low traffic overhead.
 * When user A sends this packet to user B, user A wants to let user B known
 * that user A has new entity, which may be interesting for user B.
 */
struct _DscussPayloadAdvertisement
{
  /**
   * Id of the entity to advertise.
   */
  DscussHash entity_id;

  /**
   * Date and time when the payload was composed.
   */
  GDateTime* datetime;

};


/**
 * RAW advertisement struct. All fields are in NBO.
 */
struct _DscussPayloadAdvertisementNBO
{
  /**
   * Id of the entity to advertise.
   */
  DscussHash entity_id;

  /**
   * UNIX timestamp when payload was composed.
   */
  gint64 timestamp;

};


static DscussPayloadAdvertisement*
payload_advertisement_new_full (const DscussHash* entity_id,
                                GDateTime* datetime)
{
  DscussPayloadAdvertisement* pld_advert = NULL;

  g_assert (entity_id != NULL);
  g_assert (datetime != NULL);

  pld_advert = g_new0 (DscussPayloadAdvertisement, 1);
  g_date_time_ref (datetime);
  pld_advert->datetime = datetime;
  memcpy (&pld_advert->entity_id, entity_id, sizeof (DscussHash));

  return pld_advert;
}


DscussPayloadAdvertisement*
dscuss_payload_advertisement_new (const DscussHash* entity_id)
{
  DscussPayloadAdvertisement* pld_advert = NULL;
  GDateTime* now = NULL;

  g_assert (entity_id != NULL);

  now = g_date_time_new_now_utc ();
  pld_advert = payload_advertisement_new_full (entity_id,
                                               now);
  g_date_time_unref (now);

  return pld_advert;
}


void
dscuss_payload_advertisement_free (DscussPayloadAdvertisement* advert)
{
  if (advert == NULL)
    return;

  dscuss_free_non_null (advert->datetime, g_date_time_unref);
  g_free (advert);
}


gboolean
dscuss_payload_advertisement_serialize (const DscussPayloadAdvertisement* advert,
                                        gchar** data,
                                        gsize* size)
{
  struct _DscussPayloadAdvertisementNBO* advert_nbo = NULL;

  g_assert (advert != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  *size = sizeof (struct _DscussPayloadAdvertisementNBO);
  advert_nbo = g_new0 (struct _DscussPayloadAdvertisementNBO, 1);
  *data = (gchar*) advert_nbo;

  advert_nbo->timestamp = dscuss_htonll (g_date_time_to_unix (advert->datetime));
  memcpy (&advert_nbo->entity_id, &advert->entity_id, sizeof (DscussHash));

  return TRUE;
}


DscussPayloadAdvertisement*
dscuss_payload_advertisement_deserialize (const gchar* data,
                                          gsize size)
{
  struct _DscussPayloadAdvertisementNBO* advert_nbo = NULL;
  GDateTime* datetime = NULL;
  DscussPayloadAdvertisement* advert = NULL;

  g_assert (data != NULL);

  /* Validate raw data */
  if (size <= sizeof (struct _DscussPayloadAdvertisementNBO))
    {
      g_warning ("Size of the raw data is too small."
                 " Actual size: %" G_GSIZE_FORMAT
                 ", expected: > %" G_GSIZE_FORMAT,
                 size,  sizeof (struct _DscussPayloadAdvertisementNBO));
      return NULL;
    }
  advert_nbo = (struct _DscussPayloadAdvertisementNBO*) data;

  /* Parse timestamp */
  datetime = g_date_time_new_from_unix_utc (dscuss_ntohll (advert_nbo->timestamp));
  advert = payload_advertisement_new_full (&advert_nbo->entity_id,
                                           datetime);
  g_date_time_unref (datetime);

  return advert;
}


const DscussHash*
dscuss_payload_advertisement_get_entity_id (const DscussPayloadAdvertisement* advert)
{
  g_assert (advert != NULL);
  return &advert->entity_id;
}


GDateTime*
dscuss_payload_advertisement_get_datetime (const DscussPayloadAdvertisement* advert)
{
  g_assert (advert != NULL);
  return advert->datetime;
}

