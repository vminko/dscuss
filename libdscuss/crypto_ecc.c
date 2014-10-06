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


#include "crypto_ecc.h"


#define CURVE NID_secp224r1


DscussPrivateKey*
dscuss_crypto_ecc_private_key_new ()
{
  EC_KEY* eckey = NULL;

  eckey = EC_KEY_new_by_curve_name (NID_secp224r1);
  if (NULL == eckey)
    {
      g_warning ("Failed to create new EC key");
      return NULL;
    }

  if (EC_KEY_generate_key (eckey) != 1)
    {
      g_warning ("Failed to generate EC key");
      EC_KEY_free (eckey);
      return NULL;
    }

  EC_KEY_set_asn1_flag (eckey, OPENSSL_EC_NAMED_CURVE);

  return (DscussPrivateKey*)eckey;
}


void
dscuss_crypto_ecc_private_key_free (DscussPrivateKey* privkey)
{
  EC_KEY_free ((EC_KEY*)privkey);
}


gboolean
dscuss_crypto_ecc_private_key_write (const DscussPrivateKey* privkey,
                                     const gchar* filename)
{
  gboolean result = FALSE;
  EC_KEY* eckey = (EC_KEY*)privkey;
  BIO* bio = NULL;
  gsize keylen;
  gchar *keystr = NULL;
  GError *error = NULL;

  g_assert (privkey != NULL);
  g_assert (filename != NULL);

  bio = BIO_new (BIO_s_mem ());
  if (bio == NULL)
    {
      g_warning ("Failed to create new BIO");
      goto out;
    }
  
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

  if (keystr != NULL)
    g_free (keystr);

  return result;
}


DscussPrivateKey*
dscuss_crypto_ecc_private_key_read (const gchar* filename)
{
  EC_KEY* eckey = NULL;
  gsize keylen;
  gchar *keystr;
  BIO* bio = NULL;
  GError *error = NULL;

  g_assert (filename != NULL);

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

  return (DscussPrivateKey*)eckey;
}


DscussPrivateKey*
dscuss_crypto_ecc_private_key_init (const gchar* filename)
{
  DscussPrivateKey* privkey = NULL;

  if (!g_file_test (filename, G_FILE_TEST_EXISTS)) 
    {
      g_debug ("Private key file '%s' not found, generating new one",
               filename);
      privkey = dscuss_crypto_ecc_private_key_new ();
      if (privkey == NULL)
        {
          g_critical ("Failed to generate new private key.");
          return NULL;
        }
      if (!dscuss_crypto_ecc_private_key_write (privkey,
                                                filename))
        {
          g_critical ("Failed to write the new private key to '%s'.",
                      filename);
          dscuss_crypto_ecc_private_key_free (privkey);
          return NULL;
        }
    }
  else
    {
      g_debug ("Using private key from the file '%s'", filename);
      privkey = dscuss_crypto_ecc_private_key_read (filename);
      if (privkey == NULL)
        {
          g_critical ("Failed to read private key from '%s'."
                      " If you want to generate a new private key,"
                      " remove this file.",
                      filename);
          return NULL;
        }
    }

  return privkey;
}


const DscussPublicKey*
dscuss_crypto_ecc_private_key_get_public (const DscussPrivateKey* privkey)
{
  g_assert (privkey != NULL);

  return EC_KEY_get0_public_key ((const EC_KEY*)privkey);
}


gboolean
dscuss_crypto_ecc_public_key_to_der (const DscussPublicKey* pubkey,
                                     gchar** digest,
                                     gsize* digest_len)
{
  gboolean result = FALSE;
  BIO* bio = NULL;
  EC_KEY* eckey = NULL;

  g_assert (pubkey != NULL);
  g_assert (digest != NULL);
  g_assert (digest_len != NULL);

  eckey = EC_KEY_new_by_curve_name (CURVE);
  if (eckey == NULL)
    {
      g_warning ("Failed to create new EC_KEY");
      goto out;
    }

  if (EC_KEY_set_public_key (eckey, (const EC_POINT*) pubkey) != 1)
    {
      g_warning ("Failed to set the public key for the EC_KEY");
      goto out;
    }

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
  if (eckey != NULL)
    EC_KEY_free (eckey);

  if (bio != NULL)
    BIO_free_all (bio);

  return result;
}

