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
#include <openssl/ec.h>
#include <openssl/pem.h>
#include "config.h"
#include "util.h"
#include "crypto.h"


/* Private key of this peer. */
static EC_KEY* privkey = NULL;


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
      generate_private_key (privkey_filename);
    }

  g_debug ("Using private key from the file '%s'", privkey_filename);
  eckey = read_private_key (privkey_filename);
  if (eckey == NULL)
    {
      g_error ("Failed to read private key from '%s'."
               " If you want to generate a new private key, remove this file.",
               privkey_filename);
    }

  if (privkey_filename != NULL)
    g_free (privkey_filename);

  return eckey;
}


gboolean
dscuss_crypto_init ()
{
  privkey = init_private_key ();
  if (privkey == NULL)
    {
      g_error ("Failed to initialize private key.");
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
