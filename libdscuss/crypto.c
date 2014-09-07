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
#include <openssl/ec.h>
#include <openssl/pem.h>
#include <openssl/sha.h>
#include "config.h"
#include "util.h"
#include "crypto.h"

#define REQUIRED_ZERO_NUM 30
#define SALT "dscuss-proof-of-work"


/* Private key of this peer. */
static EC_KEY* privkey = NULL;

/* Proof-of-work for the peer's public key. */
guint64 proof = 0;


static gboolean
serialize_public_key (EC_KEY* eckey, gchar** digest, gsize* digest_len);


/*** PROOF-OF-WORK ***********************************************************/
/* TODO: 1. Probably it's better to use scrypt instead of PBKDF2 for hashing;
 *       2. Store current progress in some temporary file. */


/**
 * Read value of the specified bit in hash.
 *
 * @param hash  Hash to read bit from.
 * @param hash  Number of the bit to read.
 * @return      The value of the bit: 0 or 1.
 */
static gint
hash_get_bit (const DscussHash* hash,
              guint bit)
{
  g_assert (bit < 8 * sizeof (DscussHash));
  return (((unsigned char *) hash)[bit >> 3] & (1 << (bit & 7))) > 0;
}


/**
 * Count the leading zero bits in hash.
 *
 * @param hash  Hash to count leading zeros in.
 * @return      The number of leading zeros.
 */
static guint
count_leading_zeroes (const DscussHash* hash)
{
  guint hash_count = 0;
  while ((0 == hash_get_bit (hash, hash_count)))
    hash_count++;
  return hash_count;
}


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
hash_pow (const char* pubkey_digest,
          gsize digest_len,
          guint64 proof,
          DscussHash* hash)
{
  int result = 0;
  gchar* to_hash = g_malloc0 (digest_len + sizeof (guint64));

  guint64 nproof = dscuss_htonll (proof);
  memcpy (&to_hash[0], pubkey_digest, digest_len);
  memcpy (&to_hash[digest_len], &nproof, sizeof (guint64));

  result = PKCS5_PBKDF2_HMAC (to_hash,
                              digest_len + sizeof (guint64),
                              (const unsigned char *) SALT,
                              strlen (SALT),
                              1, // iterations
                              EVP_sha512 (),
                              SHA512_DIGEST_LENGTH,
                              (unsigned char*) hash);

  g_free (to_hash);

  return (result == 1);
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
is_pow_valid (const char* pubkey_digest, gsize digest_len, guint64 proof)
{
  static DscussHash hash;

  if (!hash_pow (pubkey_digest, digest_len, proof, &hash))
    {
      g_warning ("Failed to calculate PoW-hash");
      return FALSE;
    }

  return (count_leading_zeroes (&hash) >= REQUIRED_ZERO_NUM);
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
write_pow (const gchar* filename,
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

  g_debug ("Proof-of-work successfully written!");
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
read_pow (const gchar* filename, guint64* proof)
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


/* Find proof of work for a public key.
 *
 * @param eckey       Public key to find proof for.
 * @param start_from  Initial proof value to start from (to continue
 *                    terminated calculations). Should be 0 when searching
 *                    from scratch.
 * @param proof       Where to store proof of work (in case of success).
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
static gboolean
find_pow (EC_KEY* eckey,
          guint64 start_from,
          guint64* proof)
{
  guint64 counter = 0;
  char* pubkey_digest = NULL;
  gsize digest_len = 0;
  gboolean result = FALSE;

  if (!serialize_public_key (eckey, &pubkey_digest, &digest_len))
    {
      g_warning ("Failed to serialize public key");
      goto out;
    }

  for (counter = start_from;
       counter < G_MAXUINT64;
       counter++)
    {
      if (is_pow_valid (pubkey_digest, digest_len, counter))
        {
          g_debug ("Proof of work found: %" G_GUINT64_FORMAT, counter);
          *proof = counter;
          result = TRUE;
          break;
        }
    }

out:
  if (pubkey_digest != NULL)
    g_free (pubkey_digest);

  return result;
}


/**
 * Initializes proof-of-work (PoW) for the crypto subsystem.
 *
 * Reads PoW from the PoW-file or generates a new one if there is no such file.
 * Should only be called when private key is initialized.
 */
static gboolean
init_pow (guint64* proof)
{
  gchar *pow_filename = NULL;
  gboolean result = FALSE;

  g_assert (privkey != NULL);

  pow_filename = g_build_filename (dscuss_util_get_data_dir (),
                                   "proof_of_work", NULL);
  if (!g_file_test (pow_filename, G_FILE_TEST_EXISTS))
    {
      g_debug ("Proof-of-work file '%s' not found, generating new one."
               " This will take a while.",
               pow_filename);
      if (find_pow (privkey, 0, proof))
        {
          if (!write_pow (pow_filename, *proof))
            {
              g_error ("Failed to write proof-of-work to '%s'", pow_filename);
              goto out;
            }
        }
      else
        {
          g_error ("Failed to find proof-of-work.");
          goto out;
        }
    }
  else
    {
      g_debug ("Using proof-of-work from the file '%s'", pow_filename);
      if (!read_pow (pow_filename, proof))
        {
          g_error ("Failed to read proof-of-work from '%s'."
                   " If you want to generate a new proof-of-work, remove this file.",
                   pow_filename);
          goto out;
        }
    }
  result = TRUE;
  g_debug ("Proof-of-work successfully initialized!");

out:
  if (pow_filename != NULL)
    g_free (pow_filename);

  return result;
}

/*** END OF PROOF-OF-WORK ****************************************************/


/*** KEY MANAGEMENT **********************************************************/

static gboolean
serialize_public_key (EC_KEY* eckey, gchar** digest, gsize* digest_len)
{
  gboolean result = FALSE;
  BIO* bio = NULL;

  bio = BIO_new (BIO_s_mem ());
  if (bio == NULL)
    {
      g_warning ("Failed to create new BIO");
      goto out;
    }

  if (i2d_EC_PUBKEY_bio (bio, eckey) != 1)
    {
      g_warning ("Failed to write public key to BIO");
      goto out;
    }

  *digest_len = BIO_pending (bio);
  *digest = g_malloc0 (*digest_len);
  if (BIO_read (bio, *digest, *digest_len) <= 0)
    {
      g_warning ("Failed to read key digest from BIO");
      goto out;
    }

  result = TRUE;

out:
  if (bio != NULL)
    BIO_free_all (bio);

  return result;
}


/* Generates new private key and writes it to a file. */
static gboolean
generate_private_key (const gchar* filename)
{
  gboolean result = FALSE;
  EC_KEY* eckey = NULL;
  BIO* bio = NULL;
  gsize keylen;
  gchar *keystr = NULL;
  GError *error = NULL;

  
  eckey = EC_KEY_new_by_curve_name (NID_secp224r1);
  if (NULL == eckey)
    {
      g_warning ("Failed to create new EC key");
      goto out;
    }

  if (EC_KEY_generate_key (eckey) != 1)
    {
      g_warning ("Failed to generate EC key");
      goto out;
    }

  bio = BIO_new (BIO_s_mem ());
  if (bio == NULL)
    {
      g_warning ("Failed to create new BIO");
      goto out;
    }
  
  EC_KEY_set_asn1_flag (eckey, OPENSSL_EC_NAMED_CURVE);
  if (!PEM_write_bio_ECPrivateKey (bio, eckey,
                                   NULL, NULL, 0, NULL, NULL))
    {
      g_warning ("Failed to write EC key to BIO");
      goto out;
    }

  keylen = BIO_pending (bio);
  keystr = g_malloc0 (keylen + 1);
  if (BIO_read (bio, keystr, keylen) <= 0)
    {
      g_warning ("Failed to read EC key from BIO");
      goto out;
    }

  g_debug ("Successfully generated new private key:\n%s", keystr);

  if (!g_file_set_contents (filename,
                            keystr,
                            keylen,
                            &error))
    {
      g_warning ("Couldn't write generated private key to '%s' : %s",
                 filename, error->message);
      g_error_free (error);
      goto out;
    }

  result = TRUE;

out:
  if (bio != NULL)
    BIO_free_all (bio);

  if (eckey != NULL)
    EC_KEY_free (eckey);

  if (keystr != NULL)
    g_free (keystr);

  return result;
}


/* Reads private key from a file. */
static EC_KEY*
read_private_key (const gchar* filename)
{
  EC_KEY* eckey = NULL;
  gsize keylen;
  gchar *keystr;
  BIO* bio = NULL;
  GError *error = NULL;

  if (!g_file_get_contents (filename,
                            &keystr,
                            &keylen,
                            &error))
    {
      g_warning ("Couldn't read private key from '%s' : %s",
                 filename, error->message);
      g_error_free (error);
      goto out;
    }

  bio = BIO_new (BIO_s_mem ());
  if (BIO_write (bio, keystr, keylen) <= 0)
    {
      g_warning ("Failed to read EC key to  BIO");
      goto out;
    }

  eckey = PEM_read_bio_ECPrivateKey (bio, NULL, NULL, NULL);
  if (eckey == NULL)
    {
      g_warning ("Unable to load EC key from BIO");
      goto out;
    }
  
  g_debug ("EC key successfully loaded!");

out:
  if (bio != NULL)
    BIO_free_all (bio);

  if (keystr != NULL)
    g_free (keystr);

  return eckey;
}


static EC_KEY*
init_private_key ()
{
  EC_KEY* eckey = NULL;
  gchar *privkey_filename = NULL;

  privkey_filename = g_build_filename (dscuss_util_get_data_dir (),
                                       "privkey.pem", NULL);
  if (!g_file_test (privkey_filename, G_FILE_TEST_EXISTS)) 
    {
      g_debug ("Private key file '%s' not found, generating new one",
               privkey_filename);
      if (!generate_private_key (privkey_filename))
        {
          g_error ("Failed to generate new private key.");
          goto out;
        }
    }

  g_debug ("Using private key from the file '%s'", privkey_filename);
  eckey = read_private_key (privkey_filename);
  if (eckey == NULL)
    {
      g_error ("Failed to read private key from '%s'."
               " If you want to generate a new private key, remove this file.",
               privkey_filename);
    }

out:
  if (privkey_filename != NULL)
    g_free (privkey_filename);

  return eckey;
}

/**** END OF KEY MANAGEMENT **************************************************/


gboolean
dscuss_crypto_init ()
{
  privkey = init_private_key ();
  if (privkey == NULL)
    {
      g_error ("Failed to initialize private key.");
      goto error;
    }
  if (!init_pow (&proof))
    {
      g_error ("Failed to initialize proof-of-work.");
      goto error;
    }
  return TRUE;

error:
  dscuss_crypto_uninit ();
  return FALSE;
}


void
dscuss_crypto_uninit ()
{
  if (privkey != NULL)
    EC_KEY_free(privkey);
}
