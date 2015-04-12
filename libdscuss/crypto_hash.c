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

#include <string.h>
#include <openssl/evp.h>
#include "crypto_hash.h"
#include "util.h"

#define DSCUSS_CRYPTO_HASH_DESCRIPTION_MAX_LEN 1024

static gchar description_buf[DSCUSS_CRYPTO_HASH_DESCRIPTION_MAX_LEN];


void
dscuss_crypto_hash_sha512 (const gchar* digest,
                           gsize digest_len,
                           DscussHash* hash)
{
  SHA512 ((unsigned char*) digest,
          digest_len,
          (unsigned char*) hash);
}


gboolean
dscuss_crypto_hash_pbkdf2_hmac_sha512 (const gchar* password,
                                       gsize password_len,
                                       const gchar* salt,
                                       guint iter,
                                       DscussHash* hash)
{
  int result = 0;
  result = PKCS5_PBKDF2_HMAC (password,
                              password_len,
                              (const unsigned char *) salt,
                              strlen (salt),
                              iter,
                              EVP_sha512 (),
                              SHA512_DIGEST_LENGTH,
                              (unsigned char*) hash);
  return (result == 1);
}


gint
dscuss_crypto_hash_get_bit (const DscussHash* hash,
                            guint bit)
{
  g_assert (bit < 8 * sizeof (DscussHash));
  return (((unsigned char *) hash)[bit >> 3] & (1 << (bit & 7))) > 0;
}


guint
dscuss_crypto_hash_count_leading_zeroes (const DscussHash* hash)
{
  guint hash_count = 0;
  while ((0 == dscuss_crypto_hash_get_bit (hash, hash_count)))
    hash_count++;
  return hash_count;
}


const gchar*
dscuss_crypto_hash_to_string (const DscussHash* hash)
{
  g_assert (hash != NULL);
  dscuss_data_to_hex ((const gpointer)hash, sizeof (DscussHash), description_buf);
  return description_buf;
}
