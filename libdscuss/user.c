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
#include "user.h"
#include "util.h"
#include "crypto.h"

#define DSCUSS_USER_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_USER_DESCRIPTION_MAX_LEN];


/**
 * Handle for a user.
 */
struct _DscussUser
{
  /**
   * Type of the entity.
   * Always equals to DSCUSS_ENTITY_TYPE_USER.
   */
  DscussEntityType type;
  /**
   * Reference counter.
   */
  guint ref_count;
  /**
   * User's public key.
   */
  DscussPublicKey* pubkey;
  /**
   * User id - hash of its public key (for convenience).
   */
  DscussHash id;
  /**
   * Proof-of-work for the pubkey.
   */
  guint64 proof;
  /**
   * User's nickname.
   */
  gchar* nickname;
  /**
   * Additional information about the user.
   */
  gchar* info;
  /**
   * Registration date and time..
   */
  GDateTime* datetime;
  /**
   * Signature of the serialized entity.
   */
  struct DscussSignature signature;
};

/**
 * RAW user struct. All fields are in NBO.
 */
struct _DscussUserNBO
{
  /**
   * Length of public key.
   */
  guint16 pubkey_len;
  /**
   * Proof of work.
   */
  guint64 proof;
  /**
   * Length of nickname (can't be zero).
   */
  guint16 nickname_len;
  /**
   * Length of additional information (can be zero).
   */
  guint16 info_len;
  /**
   * UNIX timestamp when user was registered.
   */
  gint64 timestamp;
  /**
   * After this struct go public key, nickname, (optionally) additional
   * information and finally signature of the entity covering everything from
   * the beginning of the struct.
   */
};

static DscussUser*
user_new_but_signature (const DscussPublicKey* pubkey,
                        guint64 proof,
                        const gchar* nickname,
                        const gchar* info,
                        GDateTime* datetime)
{
  gchar* pubkey_digest = NULL;
  gsize pubkey_digest_len = 0;

  g_assert (pubkey != NULL);
  g_assert (nickname != NULL);
  g_assert (datetime != NULL);

  DscussUser* user = g_new0 (DscussUser, 1);
  user->type = DSCUSS_ENTITY_TYPE_USER;
  user->ref_count = 1;
  user->pubkey = dscuss_crypto_public_key_copy (pubkey);
  user->proof = proof;
  user->nickname = g_strdup (nickname);
  if (user->info != NULL)
    user->info = g_strdup (info);
  user->datetime = g_date_time_ref (datetime);

  if (!dscuss_crypto_public_key_to_der (pubkey,
                                        &pubkey_digest,
                                        &pubkey_digest_len))
    {
      g_warning ("Failed to serialize public key.");
      goto err;
    }
  dscuss_crypto_hash_sha512 (pubkey_digest,
                             pubkey_digest_len,
                             &user->id);
  g_free (pubkey_digest);

  return user;

err:
  dscuss_user_free (user);
  return NULL;
}


static gboolean
user_serialize_all_but_signature (const DscussUser* user,
                                  gchar** data,
                                  gsize* size)
{
  gchar* digest = NULL;
  struct _DscussUserNBO* user_nbo = NULL;
  gchar* pubkey_digest = NULL;
  gsize pubkey_digest_len = 0;

  g_assert (user != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  if (!dscuss_crypto_public_key_to_der (user->pubkey,
                                        &pubkey_digest,
                                        &pubkey_digest_len))
    {
      g_warning ("Failed to serialize public key");
      return FALSE;
    }

  *size = sizeof (struct _DscussUserNBO)
        + pubkey_digest_len
        + strlen (user->nickname);
  if (user->info != NULL)
    *size += strlen (user->info);
  digest = g_malloc0 (*size);
  *data = digest;

  user_nbo = (struct _DscussUserNBO*) digest;
  user_nbo->pubkey_len = g_htons (pubkey_digest_len);
  user_nbo->proof = dscuss_htonll (user->proof);
  user_nbo->nickname_len = g_htons (strlen (user->nickname));
  user_nbo->info_len = (user->info == NULL)
                     ? 0 : g_htons (strlen (user->info));
  user_nbo->timestamp = dscuss_htonll (g_date_time_to_unix (user->datetime));
  digest += sizeof (struct _DscussUserNBO);

  memcpy (digest,
          pubkey_digest,
          pubkey_digest_len);
  g_free (pubkey_digest);
  digest += pubkey_digest_len;

  memcpy (digest,
          user->nickname,
          strlen (user->nickname));
  digest += strlen (user->nickname);

  if (user->info != NULL)
    {
      memcpy (digest,
              user->info,
              strlen (user->info));
      digest += strlen (user->info);
    }

  return TRUE;
}


DscussUser*
dscuss_user_new (const DscussPublicKey* pubkey,
                 guint64 proof,
                 const gchar* nickname,
                 const gchar* info,
                 GDateTime* datetime,
                 const struct DscussSignature* signature)
{
  g_assert (signature != NULL);
  DscussUser* user = user_new_but_signature (pubkey,
                                             proof,
                                             nickname,
                                             info,
                                             datetime);
  memcpy (&user->signature,
          signature,
          sizeof (struct DscussSignature));
  return user;
}


DscussUser*
dscuss_user_emerge (const DscussPrivateKey* privkey,
                    guint64 proof,
                    const gchar* nickname,
                    const gchar* info,
                    GDateTime* datetime)
{
  const DscussPublicKey* pubkey = NULL;
  gchar* all_but_signature = NULL;
  gsize all_but_signature_size = 0;
  DscussUser* user = NULL;

  g_assert (privkey != NULL);
  /* TBD: validate input */

  pubkey = dscuss_crypto_private_key_get_public (privkey);
  user = user_new_but_signature (pubkey,
                                 proof,
                                 nickname,
                                 info,
                                 datetime);
  if (user == NULL)
    {
      dscuss_user_free (user);
      return NULL;
    }

  if (!user_serialize_all_but_signature (user,
                                         &all_but_signature,
                                         &all_but_signature_size))
    {
      g_warning ("Failed to serialize a user entity");
      dscuss_user_free (user);
      return NULL;
    }
  if (!dscuss_crypto_sign (all_but_signature,
                           all_but_signature_size,
                           privkey,
                           &user->signature))
    {
      g_warning ("Failed to sign serialized user entity");
      g_free (all_but_signature);
      dscuss_user_free (user);
      return NULL;
    }
  g_free (all_but_signature);

  return user;
}


void
dscuss_user_free (DscussUser* user)
{
  if (user == NULL)
    return;

  dscuss_free_non_null (user->pubkey,
                        dscuss_crypto_public_key_free);
  dscuss_free_non_null (user->nickname, g_free);
  dscuss_free_non_null (user->info, g_free);
  dscuss_free_non_null (user->datetime, g_date_time_unref);
  g_free (user);
}


gboolean
dscuss_user_serialize (const DscussUser* user,
                       gchar** data,
                       gsize* size)
{
  gchar* all_but_signature = NULL;
  gsize all_but_signature_size = 0;
  gchar* digest = NULL;

  g_assert (user != NULL);
  g_assert (data != NULL);
  g_assert (size != NULL);

  /* Serialize everything except signature first */
  if (!user_serialize_all_but_signature (user,
                                         &all_but_signature,
                                         &all_but_signature_size))
    {
      g_warning ("Failed to serialize a user entity");
      return FALSE;
    }

  *size = all_but_signature_size + sizeof (struct DscussSignature);
  digest = g_malloc0 (*size);
  *data = digest;
  memcpy (digest,
          all_but_signature,
          all_but_signature_size);
  memcpy (digest + all_but_signature_size,
          &user->signature,
          sizeof (struct DscussSignature));

  return TRUE;
}


DscussUser*
dscuss_user_deserialize (const gchar* data,
                         gsize size)
{
  const gchar* digest = data;
  struct _DscussUserNBO* user_nbo = NULL;
  gsize pubkey_digest_len = 0;
  DscussPublicKey* pubkey = NULL;
  guint64 proof = 0;
  gsize nickname_len = 0;
  gchar* nickname = NULL;
  gsize info_len = 0;
  gchar* info = NULL;
  GDateTime* datetime = NULL;
  struct DscussSignature signature;
  DscussUser* user = NULL;

  g_assert (data != NULL);

  if (size <= sizeof (struct _DscussUserNBO))
    {
      g_warning ("Size of the raw data is too small."
                 " Actual size: %" G_GSIZE_FORMAT
                 ", expected: >= %" G_GSIZE_FORMAT,
                 size,  sizeof (struct _DscussUserNBO));
      return NULL;
    }
  user_nbo = (struct _DscussUserNBO*) data;
  digest += sizeof (struct _DscussUserNBO);

  /* Parse pubkey */
  pubkey_digest_len = g_ntohs (user_nbo->pubkey_len);
  pubkey = dscuss_crypto_public_key_from_der (digest, pubkey_digest_len);
  if (pubkey == NULL)
    {
      g_warning ("Failed to parse public key.");
      return NULL;
    }
  digest += pubkey_digest_len;

  /* Parse proof */
  proof = dscuss_ntohll (user_nbo->proof);

  /* Parse nickname */
  nickname_len = g_ntohs (user_nbo->nickname_len);
  nickname = g_malloc0 (nickname_len + 1);
  memcpy (nickname,
          digest,
          nickname_len);
  nickname[nickname_len] = '\0';
  digest += nickname_len;

  /* Parse info if any */
  info_len = g_ntohs (user_nbo->info_len);
  if (info_len > 0)
    {
      info = g_malloc0 (info_len + 1);
      memcpy (info, digest, info_len);
      info[info_len] = '\0';
      digest += info_len;
    }

  /* Parse timestamp */
  datetime = g_date_time_new_from_unix_utc (dscuss_ntohll (user_nbo->timestamp));

  /* Parse signature */
  memcpy (&signature,
          digest,
          sizeof (struct DscussSignature));

  user = dscuss_user_new (pubkey,
                          proof,
                          nickname,
                          info,
                          datetime,
                          &signature);

  dscuss_crypto_public_key_free (pubkey);
  g_free (nickname);
  if (info != info)
    g_free (info);
  g_date_time_unref (datetime);

  return user;
}


const gchar*
dscuss_user_get_description (const DscussUser* user)
{
  g_assert (user != NULL);
  g_snprintf (description_buf, 
              DSCUSS_USER_DESCRIPTION_MAX_LEN,
              "%s",
              user->nickname);
  return description_buf;
}


const DscussPublicKey*
dscuss_user_get_public_key (const DscussUser* user)
{
  g_assert (user != NULL);
  return user->pubkey;
}


const DscussHash*
dscuss_user_get_id (const DscussUser* user)
{
  g_assert (user != NULL);
  return &user->id;
}


guint64
dscuss_user_get_proof (const DscussUser* user)
{
  g_assert (user != NULL);
  return user->proof;
}


const gchar*
dscuss_user_get_nickname (const DscussUser* user)
{
  g_assert (user != NULL);
  return user->nickname;
}


const gchar*
dscuss_user_get_info (const DscussUser* user)
{
  g_assert (user != NULL);
  return user->info;
}


GDateTime*
dscuss_user_get_datetime (const DscussUser* user)
{
  g_assert (user != NULL);
  return user->datetime;
}


const struct DscussSignature*
dscuss_user_get_signature (const DscussUser* user)
{
  g_assert (user != NULL);
  return &user->signature;
}
