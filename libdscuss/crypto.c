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
#include <openssl/sha.h>
#include "config.h"
#include "util.h"
#include "crypto.h"
#include "crypto_pow.h"


/* Private key of this peer. */
static DscussPrivateKey* privkey = NULL;

/* Proof-of-work for the peer's public key. */
static guint64 proof = 0;

/* Function to call when initialization is finished. */
static DscussCryptoInitCallback init_callback;

/* User data to pass to @a init_callback. */
static gpointer init_data;


static void
on_pow_init_finished (gboolean result,
                      guint64 proof_,
                      gpointer user_data)
{
  g_debug ("Proof-of-work initialization finished.");

  if (result)
    proof = proof_;
  else
    dscuss_crypto_uninit ();

  init_callback (result, init_data);
}


gboolean
dscuss_crypto_init (DscussCryptoInitCallback callback,
                    gpointer user_data)
{
  gchar* privkey_filename = NULL;
  gchar* pow_filename = NULL;
  const DscussPublicKey* pubkey = NULL;
  gboolean result = FALSE;

  init_callback = callback;
  init_data = user_data;

  privkey_filename = g_build_filename (dscuss_util_get_data_dir (),
                                       "privkey.pem", NULL);
  privkey = dscuss_crypto_ecc_private_key_init (privkey_filename);
  g_free (privkey_filename);
  if (privkey == NULL)
    {
      g_error ("Failed to initialize private key.");
      goto error;
    }

  pubkey = dscuss_crypto_ecc_private_key_get_public (privkey);
  pow_filename = g_build_filename (dscuss_util_get_data_dir (),
                                   "proof_of_work", NULL);
  result = dscuss_crypto_pow_init (pubkey,
                                   pow_filename,
                                   on_pow_init_finished,
                                   NULL);
  g_free (pow_filename);
  if (!result)
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
  g_debug ("Uninitializing the crypto subsystem");

  dscuss_crypto_pow_uninit ();

  dscuss_free_non_null (privkey,
                        dscuss_crypto_ecc_private_key_free);
}


DscussPrivateKey*
dscuss_crypto_get_privkey ()
{
  g_assert (privkey != NULL);
  return privkey;
}
