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
#include <gio/gio.h>
#include "util.h"
#include "topic.h"
#include "subscriptions.h"


/* List of the topics the user is subscribed to. */
static GSList* topics = NULL;


gboolean
dscuss_subscriptions_init (void)
{
  gchar* path;
  GFile* file;
  GFileInputStream* file_in = NULL;
  GDataInputStream* data_in = NULL;
  gchar* line;
  DscussTopic* topic = NULL;
  GError* error = NULL;
  gboolean res = TRUE;

  g_debug ("Initializing subscriptions.");

  path = g_build_filename (dscuss_util_get_data_dir (), "subscriptions", NULL);
  file = g_file_new_for_path (path);

  file_in = g_file_read (file, NULL, &error);
  if (error != NULL)
    {
      g_critical ("Failed to open file '%s': %s",
                  path, error->message);
      g_error_free (error);
      g_object_unref (file);
      g_free (path);
      return FALSE;
    }
  
  data_in = g_data_input_stream_new ((GInputStream*) file_in);
  error = NULL;
  while (TRUE)
    {
      line = g_data_input_stream_read_line_utf8 (G_DATA_INPUT_STREAM (data_in),
                                                 NULL,
                                                 NULL,
                                                 &error);
      if (error != NULL)
        {
          g_warning ("Failed to read topics from '%s': %s",
                     path, error->message);
          g_error_free (error);
          res = FALSE;
          break;
        }

      if (line == NULL)
        {
          /* Whole file processed successfully,
           * return TRUE if there was at least one topic. */
          res = (topics != NULL);
          break;
        }

      topic = dscuss_topic_new (line);
      if (topic == NULL)
        {
          g_warning ("Malformed line in the subscriptions file: '%s'."
                     " Ignoring it.", line);
          g_free (line);
          continue;
        }

      if (g_slist_find_custom (topics,
                               topic,
                               (GCompareFunc) dscuss_topic_compare) != NULL)
        {
          g_warning ("Duplicated topic in the subscriptions file: '%s'!",
                     line);
          dscuss_topic_free (topic);
        }
      else
        {
          topics = g_slist_append (topics, topic);
        }

      g_free (line);
    }

  g_object_unref (data_in);
  g_object_unref (file_in);
  g_object_unref (file);
  g_free (path);

  if (!res)
    dscuss_subscriptions_uninit ();

  return res;
}


void
dscuss_subscriptions_uninit (void)
{
  g_debug ("Uninitializing subscriptions.");
  if (topics != NULL)
    {
      g_slist_free_full (topics, (GDestroyNotify) dscuss_topic_free);
      topics = NULL;
    }
}


const GSList*
dscuss_subscriptions_get (void)
{
  return topics;
}
