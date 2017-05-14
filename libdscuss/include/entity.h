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
 * @file include/entity.h  Dscuss entity definition.
 */

#ifndef DSCUSS_INCLUDE_ENTITY_H
#define DSCUSS_INCLUDE_ENTITY_H

#include <glib.h>
#include "crypto_hash.h"


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Dscuss entity types.
 */
typedef enum
{
  /*
   * User registers, post messages and performs operations.
   */
  DSCUSS_ENTITY_TYPE_USER = 0,
  /*
   * Some information published by a user.
   */
  DSCUSS_ENTITY_TYPE_MSG,
  /*
   * An action performed on a user or a message.
   */
  DSCUSS_ENTITY_TYPE_OPER,

} DscussEntityType;

/**
 * Handle for an entity.
 */
typedef struct _DscussEntity DscussEntity;

/**
 * Returns type of an entity.
 *
 * @param entity  Entity to get type of.
 *
 * @return Entity type.
 */
DscussEntityType
dscuss_entity_get_type (const DscussEntity* entity);

/**
 * Returns ID of an entity.
 *
 * @param entity  Entity to get ID of.
 *
 * @return  ID of the entity.
 */
const DscussHash*
dscuss_entity_get_id (const DscussEntity* entity);

/**
 * Composes a one-line text description of an entity.
 *
 * @param entity  Entity to compose description for.
 *
 * @return Text description of the entity.
 */
const gchar*
dscuss_entity_get_description (const DscussEntity* entity);

/**
 * Increases the reference count of an entity.
 *
 * @param entity  an entity.
 *
 * Returns: the same entity.
 */
DscussEntity*
dscuss_entity_ref (DscussEntity* entity);

/**
 * Decreases the reference count of an entity. When its reference count
 * drops to 0, the entity is freed.
 *
 * @param entity  an entity.
 */
void
dscuss_entity_unref (DscussEntity* entity);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_INCLUDE_ENTITY_H */
