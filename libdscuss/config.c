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
#include "config.h"
#include "util.h"


static GKeyFile* key_file = NULL;
static gchar* conf_filename = NULL;


gboolean
dscuss_config_init (void)
{
  GError* error = NULL;

  conf_filename = g_build_filename (dscuss_util_get_data_dir (),
                                     "config", NULL);

  if (g_file_test (conf_filename, G_FILE_TEST_EXISTS)) 
    {
      g_debug ("Using config file '%s'", conf_filename);
      key_file = g_key_file_new ();
      if (!g_key_file_load_from_file (key_file, conf_filename,
                                      G_KEY_FILE_KEEP_COMMENTS, &error))
        {
          g_warning ("Couldn't read '%s' : %s",
                     conf_filename, error->message);
          g_error_free (error);
          goto error;
        }
    }
  else
    {
      g_debug ("Config file '%s' not found", conf_filename);
    }

  return TRUE;

error:
  dscuss_config_uninit ();
  return FALSE;
}


void
dscuss_config_uninit (void)
{
  /**
   * TBD, g_key_file_save_to_file available since glib-2.40

  GError *error = NULL;

  if (key_file && conf_filename)
    {
      if (!g_key_file_save_to_file (key_file, conf_filename, &error))
        {
          g_warning ("Couldn't save settings to '%s' : %s",
                     conf_filename, error->message);
          g_error_free (error);
        }
    }
   */

  dscuss_free_non_null (key_file, g_key_file_free);
  dscuss_free_non_null (conf_filename, g_free);
}


gint
dscuss_config_get_integer (const gchar* group,
                           const gchar* param,
                           gint default_value)
{
  GError* error = NULL;
  gint res = 0;

  if (key_file == NULL)
    return default_value;
  
  res = g_key_file_get_integer (key_file, group, param, &error);
  if (error)
    {
      g_debug ("Couldn't get integer value of the key '%s' of the group '%s' : %s",
               param, group, error->message);
      g_error_free (error);
      return default_value;
    }

  return res;
}

