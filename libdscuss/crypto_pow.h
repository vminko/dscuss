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
 * @file crypto_pow.h  Defines the functions for initialization proof-of-work.
 */


#ifndef DSCUSS_CRYPTO_POW_H
#define DSCUSS_CRYPTO_POW_H

#include <glib.h>
#include <openssl/sha.h>
#include "util.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Callback to notify that initialization of the proof-of-work is finished.
 *
 * @param result      @c TRUE if initialization was successful,
 *                    or @c FALSE otherwise.
 * @param proof       proof-of-work if @a result is @c TRUE, @c 0 otherwise.
 * @param user_data   The user data.
 */
typedef void (*DscussCryptoPowInitCallback)(gboolean result,
                                            guint64 proof,
                                            gpointer user_data);

/**
 * Initializes proof-of-work (PoW) for the crypto subsystem.
 *
 * Reads PoW from the PoW-file or generates a new one if there is no such file.
 *
 * @param pubkey     Public key to find proof for.
 * @param filename   Name of the file to read from or to store to if it does
 *                   not exist.
 * @param callback   The function to be called when initialization
 *                   is finished.
 * @param user_data  Additional data to be passed to @a callback.
 *
 * @return @c TRUE if initialization started successfully (the proof will be
 *         passes to the callback),
 *         or @c FALSE otherwise (callback will not be called at all).
 */
gboolean
dscuss_crypto_pow_init (const DscussPublicKey* pubkey,
                        const gchar* filename,
                        DscussCryptoPowInitCallback callback,
                        gpointer user_data);

/**
 * Uninitializes proof-of-work.
 *
 * Stops finding PoW if it hasn't been found yet and frees allocated memory.
 */
void
dscuss_crypto_pow_uninit ();


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CRYPTO_POW_H */
