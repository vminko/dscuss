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

#include <glib.h>
#include "entity.h"
#include "user.h"
#include "message.h"


/**
 * Handle for an entity.
 */
struct _DscussEntity
{
  /**
   * Type of the entity.
   */
  DscussEntityType type;

  /**
   * Reference counter.
   */
  guint ref_count;
};


DscussEntityType
dscuss_entity_get_type (const DscussEntity* entity)
{
  return entity->type;
}


const gchar*
dscuss_entity_get_description (const DscussEntity* entity)
{
  switch (entity->type)
    {
    case DSCUSS_ENTITY_TYPE_USER:
      g_assert_not_reached ();
      /* TBD */
      break;

    case DSCUSS_ENTITY_TYPE_MSG:
      return dscuss_message_get_description ((DscussMessage*) entity);

    case DSCUSS_ENTITY_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }
}


static void
dscuss_entity_free (DscussEntity* entity)
{
  g_debug ("Freeing entity '%s'",
           dscuss_entity_get_description (entity));
  switch (entity->type)
    {
    case DSCUSS_ENTITY_TYPE_USER:
      dscuss_user_free ((DscussUser*) entity);
      break;

    case DSCUSS_ENTITY_TYPE_MSG:
      dscuss_message_free ((DscussMessage*) entity);
      break;

    case DSCUSS_ENTITY_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }
}


DscussEntity*
dscuss_entity_ref (DscussEntity* entity)
{
  g_return_val_if_fail (entity, NULL);
  g_return_val_if_fail (entity->ref_count > 0, NULL);

  entity->ref_count++;

  return entity;
}


void
dscuss_entity_unref (DscussEntity* entity)
{
  g_debug ("Unrefing entity '%d'", entity->ref_count);
  g_return_if_fail (entity);
  g_return_if_fail (entity->ref_count > 0);
  
  entity->ref_count--;
  if (!entity->ref_count)
    {
      dscuss_entity_free (entity);
    }
}
