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
#include "util.h"


static gchar* default_data_dir = NULL;
static gchar* custom_data_dir = NULL;


void
dscuss_util_init (const gchar* data_dir)
{
  if (data_dir != NULL && *data_dir)
    custom_data_dir = g_strdup (data_dir);
}


void
dscuss_util_uninit (void)
{
  g_free (custom_data_dir);
  custom_data_dir = NULL;
  g_free (default_data_dir);
  default_data_dir = NULL;
}


const gchar*
dscuss_util_get_data_dir (void)
{
  if (custom_data_dir != NULL)
    return custom_data_dir;

  if (!default_data_dir)
    default_data_dir = g_build_filename (g_get_home_dir (), ".dscuss", NULL);
  return default_data_dir;
}
