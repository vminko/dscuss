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
 * Callback to notify that the search of proof-of-work is over.
 *
 * @param result      @c TRUE if proof-of-work is found, @c FALSE otherwise.
 * @param proof       proof-of-work if @a result is @c TRUE, @c 0 otherwise.
 * @param user_data   The user data.
 */
typedef void (*DscussCryptoPowFindCallback)(gboolean result,
                                            guint64 proof,
                                            gpointer user_data);

/**
 * Finds proof-of-work (PoW) for the specified public key.
 *
 * Continues searching from the PoW-file or starts from scratch if there is no
 * such file.
 *
 * @param pubkey     Public key to find proof for.
 * @param filename   Name of the file to store progress.
 * @param callback   The function to be called when the search is over.
 * @param user_data  Additional data to be passed to @a callback.
 *
 * @return @c TRUE if the search started successfully (the proof will be
 *         passes to the callback),
 *         or @c FALSE otherwise (callback will not be called at all).
 */
gboolean
dscuss_crypto_pow_find (const DscussPublicKey* pubkey,
                        const gchar* filename,
                        DscussCryptoPowFindCallback callback,
                        gpointer user_data);

/**
 * Stop the search of proof-of-work.
 *
 * Stops finding PoW and frees allocated memory.
 */
void
dscuss_crypto_pow_stop_finding (void);


/**
 * Validates proof of work.
 *
 * @param pubkey     Public key to find proof for.
 * @param proof       proof-of-work if @a result is @c TRUE, @c 0 otherwise.
 *
 * @return @c TRUE if given @a proof is valid for specified @a pubkey,
 *         @c FALSE otherwise.
 */
gboolean
dscuss_crypto_pow_validate (const DscussPublicKey* pubkey,
                            guint64 proof);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CRYPTO_POW_H */
