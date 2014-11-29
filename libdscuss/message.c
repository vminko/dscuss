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
#include "message.h"

#define DSCUSS_MESSAGE_DESCRIPTION_MAX_LEN 120

static gchar description_buf[DSCUSS_MESSAGE_DESCRIPTION_MAX_LEN];


/**
 * Handle for a message.
 */
struct _DscussMessage
{
  /**
   * Type of the entity.
   * Always equals to DSCUSS_ENTITY_TYPE_MSG.
   */
  DscussEntityType type;
  /**
   * Reference counter.
   */
  guint ref_count;
  /**
   * Message content. So far it's juts a plain text.
   */
  gchar* content;
};


DscussMessage*
dscuss_message_new (const gchar* content)
{
  DscussMessage *msg = g_new0 (DscussMessage, 1);
  msg->type = DSCUSS_ENTITY_TYPE_MSG;
  msg->ref_count = 1;
  msg->content = g_strdup (content);
  return msg;
}


void
dscuss_message_free (DscussMessage* msg)
{
  g_assert (msg != NULL);
  g_free (msg->content);
  g_free (msg);
}


const gchar*
dscuss_message_get_description (const DscussMessage* msg)
{
  g_assert (msg != NULL);
  g_snprintf (description_buf, 
              DSCUSS_MESSAGE_DESCRIPTION_MAX_LEN,
              "%s",
              msg->content);
  return description_buf;
}


const gchar*
dscuss_message_get_content (const DscussMessage* msg)
{
  return msg->content;
}
