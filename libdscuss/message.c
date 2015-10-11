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
#include "core.h"
#include "crypto_hash.h"
#include "util.h"
#include "message.h"

#define DSCUSS_MESSAGE_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_MESSAGE_DESCRIPTION_MAX_LEN];


/**
 * Handle for a message.
 */
struct _DscussMessage
{
  /**
   * Type of the entity.
   * Always equals to DSCUSS_ENTITY_TYPE_MSG.
   */
  DscussEntityType type;
  /**
   * Reference counter.
   */
  guint ref_count;
  /**
   * Message id - hash of serialized message except signature and signature
   * length.
   */
  DscussHash id;
  /**
   * Topic the message belongs to.
   * If topic is specified, then @c parent_id must be empty.
   */
  DscussTopic* topic;
  /**
   * ID of the parent message.
   * If this field is not empty, then @c topic must be NULL.
   */
  DscussHash parent_id;
  /**
   * Plain text message subject.
   */
  gchar* subject;
  /**
   * Message content. So far it's juts a plain text.
   */
  gchar* text;
  /**
   * User ID of the author of the message.
   */
  DscussHash author_id;
  /**
   * Date and time when the message was written.
   */
  GDateTime* datetime;
  /**
   * Length of the signature.
   */
  gsize signature_len;
  /**
   * Signature of the serialized entity.
   */
  struct DscussSignature signature;
};

/**
 * RAW message struct. All fields are in NBO.
 */
struct _DscussMessageNBO
{
  /**
   * Length of serialized topic.
   * Will be zero if this message is a reply to another message.
   */
  guint16 topic_len;
  /**
   * Length of message subject.
   */
  guint16 subject_len;
  /**
   * Proof of message text.
   */
  guint16 text_len;
  /**
   * UNIX timestamp when the message was written.
   */
  gint64 timestamp;
  /**
   * User ID of the author of the message.
   */
  DscussHash author_id;
  /**
   * ID of the parent message. Empty if this message is a new thread.
   */
  DscussHash parent_id;
  /**
   * After this structure go serialized topic (optional), subject, message
   * text and finally the signature (both length and the signature itself) of
   * the message covering everything from the beginning of the structure.
   */
};


static void
message_dump (const DscussMessage* msg)
{
  gchar* datetime_str = NULL;
  gchar* signature_str = NULL;
  gchar* topic_str = NULL;

  g_assert (msg != NULL);

  datetime_str = g_date_time_format (msg->datetime, "%F %T");
  signature_str = dscuss_data_to_hex ((const gpointer) &msg->signature,
                                      sizeof (struct DscussSignature),
                                      NULL);
  if (msg->topic != NULL)
    topic_str = dscuss_topic_to_string (msg->topic);

  g_debug ("Dumping Message entity:");
  g_debug ("  type = %d", msg->type);
  g_debug ("  ref_count = %u", msg->ref_count);
  g_debug ("  id = %s", dscuss_crypto_hash_to_string (&msg->id));
  if (topic_str != NULL)
    g_debug ("  topic = '%s'", topic_str);
  g_debug ("  parent_id = %s", dscuss_crypto_hash_to_string (&msg->parent_id));
  g_debug ("  subject = '%s'", msg->subject);
  g_debug ("  text = '%s'", msg->text);
  g_debug ("  author_id = %s", dscuss_crypto_hash_to_string (&msg->author_id));
  g_debug ("  datetime = '%s'", datetime_str);
  g_debug ("  signature = %s", signature_str);
  g_debug ("  signature_len = %" G_GSIZE_FORMAT, msg->signature_len);

  g_free (datetime_str);
  g_free (topic_str);
  g_free (signature_str);
}


static void
message_serialize_all_but_signature (const DscussMessage* msg,
                                     gchar** data,
                                     gsize* size)
{
  gchar* digest = NULL;
  struct _DscussMessageNBO* msg_nbo = NULL;
  gchar* topic_str = NULL;

  g_assert (msg != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  if (msg->topic != NULL)
    topic_str = dscuss_topic_to_string (msg->topic);
  *size = sizeof (struct _DscussMessageNBO);
  if (msg->topic != NULL)
    *size += strlen (topic_str);
  *size += strlen (msg->subject);
  *size += strlen (msg->text);
  digest = g_malloc0 (*size);
  *data = digest;

  msg_nbo = (struct _DscussMessageNBO*) digest;
  msg_nbo->topic_len = (msg->topic != NULL) ? g_htons (strlen (topic_str)) : 0;
  msg_nbo->subject_len = g_htons (strlen (msg->subject));
  msg_nbo->text_len = g_htons (strlen (msg->text));
  msg_nbo->timestamp = dscuss_htonll (g_date_time_to_unix (msg->datetime));
  memcpy (&msg_nbo->author_id, &msg->author_id, sizeof (DscussHash));
  memcpy (&msg_nbo->parent_id, &msg->parent_id, sizeof (DscussHash));
  digest += sizeof (struct _DscussMessageNBO);

  if (topic_str != NULL)
    {
      memcpy (digest, topic_str, strlen (topic_str));
      digest += strlen (topic_str);
      g_free (topic_str);
    }

  memcpy (digest, msg->subject, strlen (msg->subject));
  digest += strlen (msg->subject);

  memcpy (digest, msg->text, strlen (msg->text));
  digest += strlen (msg->text);
}


static DscussMessage*
message_new_but_signature (DscussTopic* topic,
                           const DscussHash* parent_id,
                           const gchar* subject,
                           const gchar* text,
                           const DscussHash* author_id,
                           GDateTime* datetime)
{
  gchar* all_but_signature = NULL;
  gsize all_but_signature_size = 0;

  g_assert (topic != NULL || parent_id != NULL);
  g_assert (subject != NULL);
  g_assert (text != NULL);
  g_assert (author_id != NULL);
  g_assert (datetime != NULL);
  g_assert (datetime != NULL);

  DscussMessage* msg = g_new0 (DscussMessage, 1);
  msg->type = DSCUSS_ENTITY_TYPE_MSG;
  msg->ref_count = 1;
  msg->topic = (topic != NULL) ? dscuss_topic_copy (topic) : NULL;
  msg->subject = g_strdup (subject);
  msg->text = g_strdup (text);
  memcpy (&msg->author_id,
          author_id,
          sizeof (DscussHash));
  if (parent_id != NULL)
    memcpy (&msg->parent_id,
            parent_id,
            sizeof (DscussHash));
  msg->datetime = g_date_time_ref (datetime);

  /* Calculate message ID */
  message_serialize_all_but_signature (msg,
                                       &all_but_signature,
                                       &all_but_signature_size);

  dscuss_crypto_hash_sha512 (all_but_signature,
                             all_but_signature_size,
                             &msg->id);
  g_free (all_but_signature);

  return msg;
}


DscussMessage*
dscuss_message_new_thread (DscussTopic* topic,
                           const gchar* subject,
                           const gchar* text)
{
  if (!dscuss_is_logged_in ())
    return NULL;

  const DscussHash* author_id = dscuss_user_get_id (dscuss_get_logged_user ());
  const DscussPrivateKey* privkey = dscuss_get_logged_user_private_key ();

  DscussMessage* msg = dscuss_message_new_int (topic,
                                               NULL, /* parent_id */
                                               subject,
                                               text,
                                               author_id,
                                               privkey);
  message_dump (msg);
  return msg;
}


DscussMessage*
dscuss_message_new_reply (const DscussHash* parent_id,
                          const gchar* subject,
                          const gchar* text)
{
  if (!dscuss_is_logged_in ())
    return NULL;

  const DscussHash* author_id = dscuss_user_get_id (dscuss_get_logged_user ());
  const DscussPrivateKey* privkey = dscuss_get_logged_user_private_key ();

  DscussMessage* msg = dscuss_message_new_int (NULL, /* topic */
                                               parent_id,
                                               subject,
                                               text,
                                               author_id,
                                               privkey);
  message_dump (msg);
  return msg;
}


DscussMessage*
dscuss_message_new_int (DscussTopic* topic,
                        const DscussHash* parent_id,
                        const gchar* subject,
                        const gchar* text,
                        const DscussHash* author_id,
                        const DscussPrivateKey* privkey)
{
  gchar* all_but_signature = NULL;
  gsize all_but_signature_size = 0;
  GDateTime* datetime = NULL;

  g_assert (author_id != NULL);
  g_assert (privkey != NULL);

  datetime = g_date_time_new_now_utc ();
  DscussMessage* msg = message_new_but_signature (topic,
                                                  parent_id,
                                                  subject,
                                                  text,
                                                  author_id,
                                                  datetime);
  g_date_time_unref (datetime);

  /* Serializing for the second time. TBD: optimize. */
  message_serialize_all_but_signature (msg,
                                       &all_but_signature,
                                       &all_but_signature_size);
  if (!dscuss_crypto_sign (all_but_signature,
                           all_but_signature_size,
                           privkey,
                           &msg->signature,
                           &msg->signature_len))
    {
      g_warning ("Failed to sign serialized message entity");
      g_free (all_but_signature);
      dscuss_message_free (msg);
      return NULL;
    }
  g_free (all_but_signature);

  return msg;
}


DscussMessage*
dscuss_message_new_full (DscussTopic* topic,
                         const DscussHash* parent_id,
                         const gchar* subject,
                         const gchar* text,
                         const DscussHash* author_id,
                         GDateTime* datetime,
                         const struct DscussSignature* signature,
                         gsize signature_len)
{
  g_assert (signature != NULL);
  g_assert (signature_len != 0);

  if (topic != NULL && parent_id != NULL)
    {
      g_warning ("Malformed message entity:"
                 " both topic and parent_id are specified.");
      //return NULL;
    }
  if (topic == NULL && parent_id == NULL)
    {
      g_warning ("Malformed message entity:"
                 " both topic and parent_id are not specified.");
      return NULL;
    }

  DscussMessage* msg = message_new_but_signature (topic,
                                                  parent_id,
                                                  subject,
                                                  text,
                                                  author_id,
                                                  datetime);
  memcpy (&msg->signature,
          signature,
          sizeof (struct DscussSignature));
  msg->signature_len = signature_len;

  message_dump (msg);

  return msg;
}


void
dscuss_message_free (DscussMessage* msg)
{
  if (msg == NULL)
    return;

  dscuss_free_non_null (msg->topic,
                        dscuss_topic_free);
  dscuss_free_non_null (msg->subject, g_free);
  dscuss_free_non_null (msg->text, g_free);
  dscuss_free_non_null (msg->datetime, g_date_time_unref);
  g_free (msg);
}


gboolean
dscuss_message_serialize (const DscussMessage* msg,
                          gchar** data,
                          gsize* size)
{
  gchar* all_but_signature = NULL;
  gsize all_but_signature_size = 0;
  gchar* digest = NULL;
  guint16 signature_len_nbo = 0;

  g_assert (msg != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  message_dump (msg);

  message_serialize_all_but_signature (msg,
                                       &all_but_signature,
                                       &all_but_signature_size);

  *size = all_but_signature_size
        + sizeof (signature_len_nbo)
        + sizeof (struct DscussSignature);
  g_debug ("Message size is %" G_GSIZE_FORMAT, *size);
  digest = g_malloc0 (*size);
  *data = digest;
  memcpy (digest,
          all_but_signature,
          all_but_signature_size);
  signature_len_nbo = g_htons (msg->signature_len);
  memcpy (digest + all_but_signature_size,
          &signature_len_nbo,
          sizeof (signature_len_nbo));
  memcpy (digest + all_but_signature_size + sizeof (signature_len_nbo),
          &msg->signature,
          sizeof (struct DscussSignature));
  g_free (all_but_signature);

  return TRUE;
}


DscussMessage*
dscuss_message_deserialize (const gchar* data,
                            gsize size)
{
  const gchar* digest = data;
  struct _DscussMessageNBO* msg_nbo = NULL;
  gsize topic_len = 0;
  gchar* topic_str = NULL;
  DscussTopic* topic = NULL;
  gsize subject_len = 0;
  gchar* subject = NULL;
  gsize text_len = 0;
  gchar* text = NULL;
  GDateTime* datetime = NULL;
  struct DscussSignature signature;
  guint16 signature_len_nbo = 0;
  DscussMessage* msg = NULL;

  g_assert (data != NULL);

  if (size < sizeof (struct _DscussMessageNBO))
    {
      g_warning ("Size of the raw data is too small."
                 " Actual size: %" G_GSIZE_FORMAT
                 ", expected: >= %" G_GSIZE_FORMAT,
                 size,  sizeof (struct _DscussMessageNBO));
      return NULL;
    }
  msg_nbo = (struct _DscussMessageNBO*) data;
  digest += sizeof (struct _DscussMessageNBO);

  /* Parse topic */
  topic_len = g_ntohs (msg_nbo->topic_len);
  if (topic_len > 0)
    {
      topic_str = g_malloc0 (topic_len + 1);
      memcpy (topic_str, digest, topic_len);
      topic_str[topic_len] = '\0';
      topic = dscuss_topic_new (topic_str);
      if (topic == NULL)
        {
          g_warning ("Malformed topic in the message: '%s'.", topic_str);
          g_free (topic_str);
          return NULL;
        }
      digest += topic_len;
    }

  /* Parse subject */
  subject_len = g_ntohs (msg_nbo->subject_len);
  subject = g_malloc0 (subject_len + 1);
  memcpy (subject, digest, subject_len);
  subject[subject_len] = '\0';
  digest += subject_len;

  /* Parse text */
  text_len = g_ntohs (msg_nbo->text_len);
  text = g_malloc0 (text_len + 1);
  memcpy (text, digest, text_len);
  text[text_len] = '\0';
  digest += text_len;

  /* Parse timestamp */
  datetime = g_date_time_new_from_unix_utc (dscuss_ntohll (msg_nbo->timestamp));

  /* Parse signature */
  memcpy (&signature_len_nbo,
          digest,
          sizeof (signature_len_nbo));
  digest += sizeof (signature_len_nbo);
  memcpy (&signature,
          digest,
          sizeof (struct DscussSignature));

  msg = dscuss_message_new_full (topic,
                                 &msg_nbo->parent_id,
                                 subject,
                                 text,
                                 &msg_nbo->author_id,
                                 datetime,
                                 &signature,
                                 g_ntohs (signature_len_nbo));

  g_date_time_unref (datetime);
  g_free (text);
  g_free (subject);
  g_free (topic_str);

  if (msg != NULL)
    message_dump (msg);

  return msg;
}


gboolean
dscuss_message_verify_signature (const DscussMessage* msg,
                                 const DscussPublicKey* pubkey)
{
  gboolean result = FALSE;
  gchar* all_but_signature = NULL;
  gsize all_but_signature_size = 0;

  message_serialize_all_but_signature (msg,
                                       &all_but_signature,
                                       &all_but_signature_size);

  result = dscuss_crypto_verify (all_but_signature,
                                 all_but_signature_size,
                                 pubkey,
                                 &msg->signature,
                                 msg->signature_len);
  if (!result);
    {
      g_debug ("Invalid signature of the message");
    }

  g_free (all_but_signature);

  return result;
}


const gchar*
dscuss_message_get_description (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  g_snprintf (description_buf, 
              DSCUSS_MESSAGE_DESCRIPTION_MAX_LEN,
              "%s",
              msg->text);
  return description_buf;
}


const DscussHash*
dscuss_message_get_id (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return &msg->id;
}


const DscussTopic*
dscuss_message_get_topic (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return msg->topic;
}


const gchar*
dscuss_message_get_subject (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return msg->subject;
}


const gchar*
dscuss_message_get_content (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return msg->text;
}


GDateTime*
dscuss_message_get_datetime (DscussMessage* msg)
{
  g_assert (msg != NULL);
  return msg->datetime;
}


const DscussHash*
dscuss_message_get_author_id (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return &msg->author_id;
}


const DscussHash*
dscuss_message_get_parent_id (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return &msg->parent_id;
}


const struct DscussSignature*
dscuss_message_get_signature (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return &msg->signature;
}


gsize
dscuss_message_get_signature_length (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  return msg->signature_len;
}
