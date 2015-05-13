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

/**
 * @file include/crypto_hash.h  Public API of the hash.
 */


#ifndef DSCUSS_INCLUDE_CRYPTO_HASH_H
#define DSCUSS_INCLUDE_CRYPTO_HASH_H

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

#endif /* DSCUSS_INCLUDE_CRYPTO_HASH_H */
