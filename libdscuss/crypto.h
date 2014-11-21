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
#include "connection.h"
#include "crypto_ecc.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Callback to notify that initialization of the crypto subsystem is finished.
 *
 * @param result      @c TRUE if initialization was successful,
 *                    or @c FALSE otherwise.
 * @param user_data   The user data.
 */
typedef void (*DscussCryptoInitCallback)(gboolean result,
                                         gpointer user_data);

/**
 * Initializes the crypto subsystem.
 *
 * Initializes the private key and proof-of-work. @a dscuss_crypto_uninit
 * should be called if @c TRUE was passed to the callback.
 *
 * @param callback   The function to be called when initialization
 *                   is finished.
 * @param user_data  Additional data to be passed to the callback.
 *
 * @return @c TRUE if initialization started successfully (init result will be
 *         passes to the callback),
 *         or @c FALSE otherwise (callback will not be called at all).
 */
gboolean
dscuss_crypto_init (DscussCryptoInitCallback callback,
                    gpointer user_data);

/**
 * Destroys the crypto subsystem.
 *
 * Frees allocated memory. Stops finding PoW if it's still is progress.
 */
void
dscuss_crypto_uninit ();

/**
 * Returns the user's private key. Should only be called when crypto
 * is initialized.
 *
 * @return The user's private key.
 */
DscussPrivateKey*
dscuss_crypto_get_privkey ();


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CRYPTO_H */
