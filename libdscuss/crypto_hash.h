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

/**
 * @file crypto_pow.h  API of the hash functions (SHA-512).
 */


#ifndef DSCUSS_CRYPTO_HASH_H
#define DSCUSS_CRYPTO_HASH_H

#include <glib.h>
#include <openssl/sha.h>


#ifdef __cplusplus
extern "C" {
#endif


/* 512-bit hash digest. */
typedef struct
{
  unsigned char digest[SHA512_DIGEST_LENGTH];
} DscussHash;



/**
 * Creates digest using password based derivation function with salt
 * and iteration count. Uses SHA512 for hashing.
 *
 * @param data      Data to hash.
 * @param data_len  Length of @a data.
 * @param hash      Where to write the result.
 */
void
dscuss_crypto_hash_sha512 (const gchar* data,
                           gsize data_len,
                           DscussHash* hash);

/**
 * Creates hash using password based derivation function with salt
 * and iteration count. Uses SHA512 for hashing.
 *
 * @param password      Password used in the derivation.
 * @param password_len  Length of @a password.
 * @param salt          Null-terminated salt used for the derivation.
 * @param iter          Number or iterations (must be @c>=1 ).
 * @param hash          Where to write the result (derived key).
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
gboolean
dscuss_crypto_hash_pbkdf2_hmac_sha512 (const gchar* password,
                                       gsize password_len,
                                       const gchar* salt,
                                       guint iter,
                                       DscussHash* hash);

/**
 * Read value of the specified bit in hash.
 *
 * @param hash  Hash to read bit from.
 * @param hash  Number of the bit to read.
 *
 * @return  The value of the bit: 0 or 1.
 */
gint
dscuss_crypto_hash_get_bit (const DscussHash* hash,
                            guint bit);

/**
 * Count the leading zero bits in hash.
 *
 * @param hash  Hash to count leading zeros in.
 *
 * @return  The number of leading zeros.
 */
guint
dscuss_crypto_hash_count_leading_zeroes (const DscussHash* hash);

/**
 * Converts the first 4 bytes of a hash to a string.
 *
 * @param user  User to compose description for.
 *
 * @return  Text description of the user.
 */
const gchar*
dscuss_crypto_hash_to_string (const DscussHash* hash);

#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CRYPTO_HASH_H */
