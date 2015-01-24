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
 * @file entity.h  Dscuss user definition.
 * @brief User entity identifies and describes a user.
 * It's like a password in real life.
 */

#ifndef DSCUSS_USER_H
#define DSCUSS_USER_H

#include <glib.h>
#include "entity.h"
#include "crypto.h"
#include "crypto_hash.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a user entity.
 */
typedef struct _DscussUser DscussUser;

/**
 * Creates a new user entity.
 *
 * @param pubkey     User's public key.
 * @param proof      Proof-of-work for the @a pubkey.
 * @param nickname   User's nickname.
 * @param info       Additional information about the user.
 * @param datetime   Registration date and time.
 * @param signature  Signature of the entity.
 *
 * @return  Newly created user entity.
 */
DscussUser*
dscuss_user_new (const DscussPublicKey* pubkey,
                 guint64 proof,
                 const gchar* nickname,
                 const gchar* info,
                 GDateTime* datetime,
                 const struct DscussSignature* signature);

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
 * Destroys a user entity.
 * Frees all memory allocated by the entity.
 *
 * @ param user  User to be destroyed.
 */
void
dscuss_user_free (DscussUser* user);

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
 * Composes a one-line text description of a user.
 *
 * @param user  User to compose description for.
 *
 * @return  Text description of the user.
 */
const gchar*
dscuss_user_get_description (const DscussUser* user);

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
 * Returns ID of the user.
 *
 * @param user  User to get ID of.
 *
 * @return  ID of the user.
 */
const DscussHash*
dscuss_user_get_id (const DscussUser* user);

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
 * Returns nickname of the user.
 *
 * @param user  User to get nickname of.
 *
 * @return  Nickname of the user.
 */
const gchar*
dscuss_user_get_nickname (const DscussUser* user);

/**
 * Returns additional information of the user.
 *
 * @param user  User to get info of.
 *
 * @return  Information about the user.
 */
const gchar*
dscuss_user_get_info (const DscussUser* user);

/**
 * Returns date and time when the user was registered.
 *
 * @param user User to get registration date and time of..
 *
 * @return  Date and time when the user was registered.
 */
GDateTime*
dscuss_user_get_datetime (const DscussUser* user);

/**
 * Returns signature the user.
 *
 * @param user User to get signature of.
 *
 * @return  Signature of the user.
 */
const struct DscussSignature*
dscuss_user_get_signature (const DscussUser* user);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_USER_H */
