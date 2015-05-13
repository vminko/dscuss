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
 * @file include/user.h  Dscuss user definition.
 * @brief User entity identifies and describes a user.
 * It's like a password in real life.
 */

#ifndef DSCUSS_INCLUDE_USER_H
#define DSCUSS_INCLUDE_USER_H

#include <glib.h>
#include "crypto_hash.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a user entity.
 */
typedef struct _DscussUser DscussUser;


/**
 * Destroys a user entity.
 * Frees all memory allocated by the entity.
 *
 * @param user  User to be destroyed.
 */
void
dscuss_user_free (DscussUser* user);

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
 * Returns ID of the user.
 *
 * @param user  User to get ID of.
 *
 * @return  ID of the user.
 */
const DscussHash*
dscuss_user_get_id (const DscussUser* user);

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
dscuss_user_get_datetime (DscussUser* user);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_INCLUDE_USER_H */
