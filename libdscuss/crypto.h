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
 * @file crypto.h  Defines API of the crypto subsystem.
 * @brief The crypto subsystem includes public key cryptography (ECDSA),
 * hashing and symmetric encryption. Based on OpenSSL.
 */


#ifndef DSCUSS_CRYPTO_H
#define DSCUSS_CRYPTO_H

#include <glib.h>
#include <openssl/sha.h>
#include "connection.h"
#include "peer.h"

#ifdef __cplusplus
extern "C" {
#endif

/* 512-bit hash digest. */
typedef struct
{
  unsigned char digest[SHA512_DIGEST_LENGTH];
} DscussHash;

/**
 * Initializes the crypto subsystem.
 *
 * Initializes the private key.
 *
 * @return @c TRUE in case of success, or @c FALSE otherwise.
 */
gboolean
dscuss_crypto_init ();

/**
 * Destroys the crypto subsystem.
 *
 * Frees allocated memory.
 */
void
dscuss_crypto_uninit ();


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CRYPTO_H */
