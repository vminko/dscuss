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

#include <string.h>
#include <stdio.h>
#include <glib.h>
#include "util.h"


static gchar* default_data_dir = NULL;
static gchar* custom_data_dir = NULL;


void
dscuss_util_init (const gchar* data_dir)
{
  g_debug ("Initializing utils.");
  if (data_dir != NULL && *data_dir)
    custom_data_dir = g_strdup (data_dir);
}


void
dscuss_util_uninit (void)
{
  g_debug ("Uninitializing utils.");
  dscuss_free_non_null (custom_data_dir, g_free);
  dscuss_free_non_null (default_data_dir, g_free);
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


gchar*
dscuss_data_to_hex (const gpointer data, gsize data_len, gchar* buffer)
{
    gchar* result = buffer;
    gsize i = 0;
    guint tmp = 0;

    if (result == NULL)
      result = g_malloc (data_len * 2 + 1);

    for (i = 0; i < data_len; i++) {
        tmp = *((guint8 *) (data + i));
        g_snprintf (result + i * 2 * sizeof(char), 3, "%02X", tmp);
    }

    return result;
}


gboolean
dscuss_data_from_hex (const gchar* hex_str, gpointer* data, gsize* data_len)
{
    gsize hex_str_len  = 0;
    const gchar* hex_str_pos = hex_str;
    gpointer result    = NULL;
    gsize result_len   = 0;
    gsize i            = 0;

    g_assert (hex_str != NULL);

    hex_str_len = strlen (hex_str);
    if (hex_str_len % 2)
      {
        g_debug ("Malformed hex string '%s': odd length.", hex_str);
        goto error;
      }

    result_len = hex_str_len / 2;
    result = g_malloc (result_len);
    for (i = 0; i < result_len; i++)
      {
        if (sscanf (hex_str_pos, "%2hhX", (gchar*)result + i) != 1)
          {
            g_debug ("Malformed hex string '%s': failed at %" G_GSIZE_FORMAT ".",
                     hex_str, i);
            goto error;
          }
        hex_str_pos += 2 * sizeof(char);
      }

    *data = result;
    *data_len = result_len;
    return TRUE;

error:
    if (result != NULL)
      g_free (result);
    return FALSE;
}


/* Slightly modified version of Glib's g_strjoinv */
gchar *
dscuss_strnjoinv (const gchar *separator,
                  gchar **str_array,
                  gsize str_array_len)
{
  gchar *string;
  gchar *ptr;

  g_return_val_if_fail (str_array != NULL, NULL);
  g_return_val_if_fail (str_array_len > 0, NULL);
  g_assert (separator != NULL);

  if (*str_array)
    {
      gint i;
      gsize len;
      gsize separator_len;

      separator_len = strlen (separator);

      /* First part, getting length */
      len = 1 + strlen (str_array[0]);
      for (i = 1; i < str_array_len; i++)
        len += strlen (str_array[i]);
      len += separator_len * (i - 1);

      /* Second part, building string */
      string = g_new (gchar, len);
      ptr = g_stpcpy (string, *str_array);
      for (i = 1; i < str_array_len; i++)
        {
          ptr = g_stpcpy (ptr, separator);
          ptr = g_stpcpy (ptr, str_array[i]);
        }
      }
  else
    string = g_strdup ("");

  return string;
}
