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

/**
 * @file user.h  Internal API for Dscuss User.
 */

#ifndef DSCUSS_USER_H
#define DSCUSS_USER_H

#include <glib.h>
#include "entity.h"
#include "crypto.h"
#include "include/user.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Creates a new user entity.
 *
 * @param pubkey         User's public key.
 * @param proof          Proof-of-work for the @a pubkey.
 * @param nickname       User's nickname.
 * @param info           Additional information about the user.
 * @param datetime       Registration date and time.
 * @param signature      Signature of the entity.
 * @param signature_len  Length of the signature.
 *
 * @return  Newly created user entity.
 */
DscussUser*
dscuss_user_new (const DscussPublicKey* pubkey,
                 guint64 proof,
                 const gchar* nickname,
                 const gchar* info,
                 GDateTime* datetime,
                 const struct DscussSignature* signature,
                 gsize signature_len);

/**
 * Emerge a new user entity. It should only be called when signature is not
 * known yet.  Signature will be created using the provided private key.
 *
 * @param privkey   User's private key (required for making signature).
 * @param proof     Proof-of-work for the @a pubkey.
 * @param nickname  User's nickname.
 * @param info      Additional information about the user.
 * @param datetime  Registration date and time.
 *
 * @return  Newly created user entity, or @c NULL on error.
 */
DscussUser*
dscuss_user_emerge (const DscussPrivateKey* privkey,
                    guint64 proof,
                    const gchar* nickname,
                    const gchar* info,
                    GDateTime* datetime);

/**
 * Convert a user to raw data, which can be transmitted via network.
 *
 * @param user  User to serialize.
 * @param data  Where to store address of the serialized user.
 * @param size  @a data size (output parameter).
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_user_serialize (const DscussUser* user,
                       gchar** data,
                       gsize* size);

/**
 * Create user from raw data.
 *
 * @param data  Raw data to parse.
 * @param size  Size of @a data.
 *
 * @return  A new user in case of success or @c NULL on error.
 */
DscussUser*
dscuss_user_deserialize (const gchar* data,
                         gsize size);

/**
 * Returns public key of the user.
 *
 * @param user  User to get public key of.
 *
 * @return  Public key of the user.
 */
const DscussPublicKey*
dscuss_user_get_public_key (const DscussUser* user);

/**
 * Returns proof-of-work of the user.
 *
 * @param user  User to get proof of.
 *
 * @return  Proof-of-work of the user.
 */
guint64
dscuss_user_get_proof (const DscussUser* user);

/**
 * Returns signature the user.
 *
 * @param user User to get signature of.
 *
 * @return  Signature of the user.
 */
const struct DscussSignature*
dscuss_user_get_signature (const DscussUser* user);

/**
 * Returns length of the signature of the user.
 *
 * @param user  User to get length from.
 *
 * @return  The length of the signature of the user.
 */
gsize
dscuss_user_get_signature_length (const DscussUser* user);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_USER_H */
