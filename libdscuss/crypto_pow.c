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

/* TODO: 1. Probably it's better to use scrypt instead of PBKDF2 for hashing;
 *       2. Store current progress in some temporary file. */


#include <string.h>
#include <glib.h>
#include <glib/gstdio.h>
#include "crypto.h"
#include "crypto_hash.h"
#include "crypto_pow.h"

#define SALT "dscuss-proof-of-work"
#define REQUIRED_ZERO_NUM 10


typedef struct PowFindContext PowFindContext;

/* ID of the event source for finding PoW. */
static guint find_pow_id = 0;

/* Name of the file for storing current progress (counter value). */
static gchar* tmp_filename = NULL;

/* Context for @a pow_find. */
static PowFindContext* find_ctx = NULL;

/* Function to call when initialization is finished. */
static DscussCryptoPowFindCallback find_callback;

/* User data to pass to @a find_callback. */
static gpointer find_data;


/**
 * Calculate proof-of-work hash (PBKDF2_HMAC using SHA512) of the specified
 * digest and proof.
 *
 * @param pubkey_digest  Serialized public key.
 * @param digest_len     Size of @pubkey_digest.
 * @param proof          Proof of work.
 * @param hash           Where to write the result (in case of success).
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
static gboolean
pow_hash (const gchar* pubkey_digest,
          gsize digest_len,
          guint64 proof,
          DscussHash* hash)
{
  gboolean result = FALSE;
  gchar* to_hash = g_malloc0 (digest_len + sizeof (guint64));

  guint64 nproof = dscuss_htonll (proof);
  memcpy (&to_hash[0], pubkey_digest, digest_len);
  memcpy (&to_hash[digest_len], &nproof, sizeof (guint64));

  result = dscuss_crypto_hash_pbkdf2_hmac_sha512 (to_hash,
                                                  digest_len + sizeof (guint64),
                                                  SALT,
                                                  1, //iterations
                                                  hash);
  g_free (to_hash);
  return result;
}


/**
 * Validates proof-of-work.
 * Checks whether hash of public key and proof of work has at least
 * REQUIRED_ZERO_NUM leading zeros.
 *
 * @param pubkey_digest  Serialized public key.
 * @param digest_len     Size of @pubkey_digest.
 * @param proof          Proof of work.
 *
 * @return @c TRUE if proof or work is valid, or @c FALSE otherwise.
 */
static gboolean
pow_is_valid (const gchar* pubkey_digest, gsize digest_len, guint64 proof)
{
  static DscussHash hash;

  if (!pow_hash (pubkey_digest, digest_len, proof, &hash))
    {
      g_warning ("Failed to calculate PoW-hash");
      return FALSE;
    }

  return (dscuss_crypto_hash_count_leading_zeroes (&hash) >= REQUIRED_ZERO_NUM);
}


/**
 * Writes proof-of-work to a files.
 *
 * @param filename  Where to store proof of work.
 * @param proof     Proof of work.
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
#define MAX_GUINT64_DEC_LEN 10
static gboolean
pow_write (const gchar* filename,
           guint64 proof)
{
  gchar proofstr[MAX_GUINT64_DEC_LEN + 1];
  GError *error = NULL;

  if (sprintf (proofstr, "%" G_GUINT64_FORMAT, proof) <= 0)
    {
      g_warning ("Failed to convert proof-of-work"
                 " '%" G_GUINT64_FORMAT "' to a string.",
                 proof);
      return FALSE;
    }

  if (!g_file_set_contents (filename,
                            proofstr,
                            strlen(proofstr),
                            &error))
    {
      g_warning ("Couldn't write proof-of-work to '%s' : %s",
                 filename, error->message);
      g_error_free (error);
      return FALSE;
    }

  return TRUE;
}


/* Reads proof-of-work from a file.
 *
 * @param filename  File to read proof of work from.
 * @param proof     Where to store proof of work (in case of success).
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
static gboolean
pow_read (const gchar* filename, guint64* proof)
{
  gboolean result = FALSE;
  gsize prooflen;
  gchar *proofstr;
  GError *error = NULL;

  if (!g_file_get_contents (filename,
                            &proofstr,
                            &prooflen,
                            &error))
    {
      g_warning ("Couldn't read private proof-of-work from '%s' : %s",
                 filename, error->message);
      g_error_free (error);
      goto out;
    }

  if (sscanf (proofstr, "%" G_GUINT64_FORMAT, proof) != 1)
    {
      g_warning ("Couldn't parse content of '%s'", filename);
      goto out;
    }

  g_debug ("Proof-of-work successfully read!");
  result = TRUE;

out:
  if (proofstr != NULL)
    g_free (proofstr);

  return result;
}


/*** PowFindContext **********************************************************/

typedef void (*PowFindCallback)(gboolean result,
                                guint64 proof,
                                gpointer user_data);

typedef struct PowFindContext
{
  gchar* digest;
  gsize digest_len;
  guint64 counter;
  const gchar* tmp_filename;
  PowFindCallback callback;
  gpointer user_data;
} PowFindContext;


static PowFindContext*
pow_find_context_new (gchar* digest,
                      gsize digest_len,
                      guint64 counter,
                      const gchar* tmp_filename,
                      PowFindCallback callback,
                      gpointer user_data)
{
  PowFindContext* ctx = g_new0 (PowFindContext, 1);
  ctx->digest = digest;
  ctx->digest_len = digest_len;
  ctx->counter = counter;
  ctx->tmp_filename = tmp_filename;
  ctx->callback = callback;
  ctx->user_data = user_data;
  return ctx;
}


static void
pow_find_context_free (PowFindContext* ctx)
{
  g_free (ctx->digest);
  g_free (ctx);
}

/*** End of PowFindContext ***************************************************/


static void
on_pow_found (gboolean find_result,
              guint64 proof,
              gpointer user_data)
{
  gboolean result = find_result;

  if (g_file_test (tmp_filename, G_FILE_TEST_EXISTS))
    {
      if (g_unlink (tmp_filename) != 0)
        {
          g_critical ("Failed to remove temporary file '%s'", tmp_filename);
          result = FALSE;
        }
    }

  dscuss_free_non_null (tmp_filename, g_free);
  find_pow_id = 0;

  find_callback (result, proof, find_data);
}


/* Find proof of work for a public key.
 *
 * @param user_data  Public key to find proof for.
 *
 * @return @c TRUE to call this function again,
 *         or @c FALSE to stop finding proof.
 */
#define POW_PROBES_PER_ITERATION 100
#define POW_PROBES_BETWEEN_WRITES 1000000
static gboolean
pow_find (gpointer user_data)
{
  PowFindContext* ctx = (PowFindContext*) user_data;
  guint probes = 0;

  for (;
       ctx->counter < G_MAXUINT64 && probes < POW_PROBES_PER_ITERATION;
       ctx->counter++, probes++)
    {
      if (pow_is_valid (ctx->digest,
                        ctx->digest_len,
                        ctx->counter))
        {
          g_debug ("Proof of work found: %" G_GUINT64_FORMAT, ctx->counter);
          ctx->callback (TRUE,
                         ctx->counter,
                         ctx->user_data);
          pow_find_context_free (ctx);
          return FALSE;
        }
    }

  if (!(ctx->counter % POW_PROBES_BETWEEN_WRITES))
    {
      g_debug ("Saving current PoW counter %" G_GUINT64_FORMAT " to %s",
               ctx->counter,
               tmp_filename);
      if (!pow_write (tmp_filename, ctx->counter))
        {
          g_warning ("Failed to save proof-of-work to '%s'", tmp_filename);
        }
    }

  if (ctx->counter == G_MAXUINT64)
    {
      g_warning ("Failed to find proof of work");
      ctx->callback (FALSE,
                     0,
                     ctx->user_data);
      pow_find_context_free (ctx);
      return FALSE;
    }

  return TRUE;
}




static gboolean
pow_start_finding (const DscussPublicKey* pubkey,
                   const gchar* filename_)
{
  gchar* digest = NULL;
  gsize digest_len;
  guint64 start_from = 0;

  if (!dscuss_crypto_public_key_to_der (pubkey, &digest, &digest_len))
    {
      g_warning ("Failed to serialize public key");
      goto error;
    }

  tmp_filename = g_strdup (filename_);
  if (g_file_test (tmp_filename, G_FILE_TEST_EXISTS))
    {
      if (!pow_read (tmp_filename, &start_from))
        {
          g_critical ("Failed to read current progress of finding"
                      " proof-of-work from '%s'. Remove this file if you want"
                      " to start finding proof-of-work from scratch.",
                      tmp_filename);
          goto error;
        }
    }

  find_ctx = pow_find_context_new (digest,
                                   digest_len,
                                   start_from,
                                   tmp_filename,
                                   on_pow_found,
                                   NULL);
  find_pow_id = g_idle_add (pow_find, find_ctx);
  return TRUE;

error:
  dscuss_free_non_null (tmp_filename, g_free);
  dscuss_free_non_null (digest, g_free);
  return FALSE;
}


gboolean
dscuss_crypto_pow_find (const DscussPublicKey* pubkey,
                        const gchar* filename,
                        DscussCryptoPowFindCallback callback,
                        gpointer user_data)
{
  g_assert (pubkey != NULL);
  g_assert (filename != NULL);

  if (find_pow_id != 0)
    {
      g_warning ("Pow finding is already in progress.");
      return FALSE;
    }

  find_callback = callback;
  find_data = user_data;

  if (!pow_start_finding (pubkey, filename))
    {
      g_warning ("Failed to start finding proof-of-work");
      find_callback = NULL;
      find_data = NULL;
      return FALSE;
    }

  return TRUE;
}


void
dscuss_crypto_pow_stop_finding ()
{
  dscuss_free_non_null (tmp_filename, g_free);
  if (find_pow_id != 0)
    {
      g_source_remove (find_pow_id);
      find_pow_id = 0;
    }
  dscuss_free_non_null (find_ctx,
                        pow_find_context_free);
  find_callback = NULL;
  find_data = NULL;
}


gboolean
dscuss_crypto_pow_validate (const DscussPublicKey* pubkey,
                            guint64 proof)
{
  gchar* digest = NULL;
  gsize digest_len;
  gboolean result = FALSE;

  if (!dscuss_crypto_public_key_to_der (pubkey, &digest, &digest_len))
    {
      g_warning ("Failed to serialize public key");
      return FALSE;
    }
  result = pow_is_valid (digest, digest_len, proof);
  g_free (digest);

  return result;
}