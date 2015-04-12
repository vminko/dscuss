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

#include <glib.h>
#include <gio/gio.h>
#include "util.h"
#include "topic.h"
#include "subscriptions.h"


GSList*
dscuss_subscriptions_read (const gchar* filename)
{
  GFile* file;
  GFileInputStream* file_in = NULL;
  GDataInputStream* data_in = NULL;
  gchar* line;
  DscussTopic* topic = NULL;
  GSList* topics = NULL;
  GError* error = NULL;
  gboolean res = FALSE;

  g_assert (filename != NULL);
  g_debug ("Reading subscriptions from '%s'.", filename);

  file = g_file_new_for_path (filename);
  file_in = g_file_read (file, NULL, &error);
  if (error != NULL)
    {
      g_critical ("Failed to open file '%s': %s",
                  filename, error->message);
      g_error_free (error);
      g_object_unref (file);
      return NULL;
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
                     filename, error->message);
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

  if (!res)
    {
      dscuss_subscriptions_free (topics);
      topics = NULL;
    }

  return topics;
}


void
dscuss_subscriptions_free (GSList* subscriptions)
{
  g_debug ("Destroying user subscriptions.");
  if (subscriptions != NULL)
    g_slist_free_full (subscriptions, (GDestroyNotify) dscuss_topic_free);
}


GSList*
dscuss_subscriptions_copy (GSList* subscriptions)
{
  GSList* subscriptions_copy = NULL;

  g_assert (subscriptions != NULL);
  g_debug ("Copying user subscriptions.");
  subscriptions_copy = g_slist_copy_deep (subscriptions,
                                          (GCopyFunc) dscuss_topic_copy, NULL);

  return subscriptions_copy;
}
