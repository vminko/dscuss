/**
 * This file is part of Dscuss.
 * Copyright (C) 2015  Vitaly Minko
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
#include "payload_hello.h"
#include "topic.h"
#include "subscriptions.h"
#include "util.h"

#define DSCUSS_PAYLOAD_HELLO_TOPIC_DELMITER ";"


/**
 * Hello packet is used for handshaking.
 * When user A sends this packet to user B, he/she:
 * 1. notifies user B about topics of A's interests;
 * 2. proves that the user A actually has the A's private key;
 *
 * timestamp and id prevent from reusing hello sent to some other peer.
 */
struct _DscussPayloadHello
{
  /**
   * Id of the user this payload is designated for.
   */
  DscussHash receiver_id;

  /**
   * Date and time when the payload was composed.
   */
  GDateTime* datetime;

  /**
   * Subscriptions of the author of the payload.
   */
  GSList* subscriptions;

};


/**
 * RAW user struct. All fields are in NBO.
 */
struct _DscussPayloadHelloNBO
{
  /**
   * Id of the user this payload is designated for.
   */
  DscussHash receiver_id;
  /**
   * UNIX timestamp when payload was composed.
   */
  gint64 timestamp;
  /**
   * Length of the string of subscriptions.
   */
  guint16 subscriptions_len;
  /**
   * After this struct goes the string of the user's subscriptions.
   */
};


static gchar*
payload_hello_subscriptions_to_string (const GSList* subscriptions)
{
  const GSList* iter = NULL;
  GPtrArray* str_array = NULL;
  gchar* string = NULL;

  g_assert (subscriptions != NULL);

  str_array = g_ptr_array_new ();
  g_ptr_array_set_free_func (str_array, g_free);
  for (iter = subscriptions; iter; iter = iter->next)
    {
      g_ptr_array_add (str_array,
                       dscuss_topic_to_string ((const DscussTopic*) iter->data));
    }

  string = dscuss_strnjoinv (DSCUSS_PAYLOAD_HELLO_TOPIC_DELMITER,
                             (gchar **)str_array->pdata,
                             str_array->len);
  g_ptr_array_free (str_array, TRUE);

  return string;
}


static GSList*
payload_hello_subscriptions_from_string (const gchar* string)
{
  GSList* topics = NULL;
  DscussTopic* topic = NULL;
  gchar** subscriptions =  NULL;
  gchar* topic_str = NULL;
  guint i = 0;

  g_assert (string != NULL);

  /* TBD: validate string before parsing */

  subscriptions = g_strsplit (string,
                              DSCUSS_PAYLOAD_HELLO_TOPIC_DELMITER,
                              0);

  while ((topic_str = subscriptions[i]) != NULL)
    {
      topic = dscuss_topic_new (topic_str);
      if (topic == NULL)
        {
          /* Malformed topic string, clean up and return NULL. */
          g_slist_free_full (topics, (GDestroyNotify) dscuss_topic_free);
          topics = NULL;
          break;
        }
      topics = g_slist_append (topics, topic);
      i++;
    }

  if (topics == NULL)
    g_warning ("Malformed subscription list: '%s'.", string);

  g_strfreev (subscriptions);
  return topics;
}


static DscussPayloadHello*
payload_hello_new_full (const DscussHash* receiver_id,
                        GSList* subscriptions,
                        GDateTime* datetime)
{
  DscussPayloadHello* pld_hello = NULL;

  g_assert (receiver_id != NULL);
  g_assert (subscriptions != NULL);
  g_assert (datetime != NULL);

  pld_hello = g_new0 (DscussPayloadHello, 1);
  pld_hello->subscriptions = dscuss_subscriptions_copy (subscriptions);
  g_date_time_ref (datetime);
  pld_hello->datetime = datetime;
  memcpy (&pld_hello->receiver_id, receiver_id, sizeof (DscussHash));

  return pld_hello;
}


DscussPayloadHello*
dscuss_payload_hello_new (const DscussHash* receiver_id,
                          GSList* subscriptions)
{
  DscussPayloadHello* pld_hello = NULL;
  GDateTime* now = NULL;

  g_assert (receiver_id != NULL);
  g_assert (subscriptions != NULL);

  now = g_date_time_new_now_utc ();
  pld_hello = payload_hello_new_full (receiver_id,
                                      subscriptions,
                                      now);
  g_date_time_unref (now);

  return pld_hello;
}


void
dscuss_payload_hello_free (DscussPayloadHello* hello)
{
  if (hello == NULL)
    return;

  dscuss_free_non_null (hello->subscriptions, dscuss_subscriptions_free);
  dscuss_free_non_null (hello->datetime, g_date_time_unref);
  g_free (hello);
}


gboolean
dscuss_payload_hello_serialize (const DscussPayloadHello* hello,
                                gchar** data,
                                gsize* size)
{
  gchar* digest = NULL;
  struct _DscussPayloadHelloNBO* hello_nbo = NULL;
  gchar* subs_string = NULL;
  gsize subs_string_len = 0;

  g_assert (hello != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  subs_string = payload_hello_subscriptions_to_string (hello->subscriptions);
  if (subs_string == NULL)
    {
      g_warning ("Failed to serialize subscriptions.");
      return FALSE;
    }
  subs_string_len = strlen (subs_string) + 1;

  *size = sizeof (struct _DscussPayloadHelloNBO) + subs_string_len;
  digest = g_malloc0 (*size);
  *data = digest;

  hello_nbo = (struct _DscussPayloadHelloNBO*) digest;
  hello_nbo->subscriptions_len = g_htons (subs_string_len);
  hello_nbo->timestamp = dscuss_htonll (g_date_time_to_unix (hello->datetime));
  memcpy (&hello_nbo->receiver_id, &hello->receiver_id, sizeof (DscussHash));
  digest += sizeof (struct _DscussPayloadHelloNBO);
  memcpy (digest,
          subs_string,
          subs_string_len);
  g_free (subs_string);

  return TRUE;
}


DscussPayloadHello*
dscuss_payload_hello_deserialize (const gchar* data,
                                  gsize size)
{
  const gchar* digest = data;
  struct _DscussPayloadHelloNBO* hello_nbo = NULL;
  gsize subs_string_len = 0;
  GDateTime* datetime = NULL;
  DscussPayloadHello* hello = NULL;
  GSList* subs;

  g_assert (data != NULL);

  /* Validate raw data */
  if (size <= sizeof (struct _DscussPayloadHelloNBO))
    {
      g_warning ("Size of the raw data is too small."
                 " Actual size: %" G_GSIZE_FORMAT
                 ", expected: > %" G_GSIZE_FORMAT,
                 size,  sizeof (struct _DscussPayloadHelloNBO));
      return NULL;
    }
  hello_nbo = (struct _DscussPayloadHelloNBO*) data;

  subs_string_len = g_ntohs (hello_nbo->subscriptions_len);
  if (size != sizeof (struct _DscussPayloadHelloNBO) + subs_string_len)
    {
      g_warning ("Size of the raw data is wrong."
                 " Actual size: %" G_GSIZE_FORMAT
                 ", expected: %" G_GSIZE_FORMAT,
                 size,
                 sizeof (struct _DscussPayloadHelloNBO) + subs_string_len);
      return NULL;
    }

  /* Parse subscriptions */
  digest += sizeof (struct _DscussPayloadHelloNBO);
  subs = payload_hello_subscriptions_from_string (digest);
  if (subs == NULL)
    {
      g_warning ("Failed to parse subscriptions in the payload.");
      return NULL;
    }

  /* Parse timestamp */
  datetime = g_date_time_new_from_unix_utc (dscuss_ntohll (hello_nbo->timestamp));

  hello = payload_hello_new_full (&hello_nbo->receiver_id,
                                  subs,
                                  datetime);

  dscuss_free_non_null (subs, dscuss_subscriptions_free);
  g_date_time_unref (datetime);

  return hello;
}


const DscussHash*
dscuss_payload_hello_get_receiver_id (const DscussPayloadHello* hello)
{
  g_assert (hello != NULL);
  return &hello->receiver_id;
}


GDateTime*
dscuss_payload_hello_get_datetime (const DscussPayloadHello* hello)
{
  g_assert (hello != NULL);
  return hello->datetime;
}


GSList*
dscuss_payload_hello_get_subscriptions (DscussPayloadHello* hello)
{
  g_assert (hello != NULL);
  return hello->subscriptions;
}
