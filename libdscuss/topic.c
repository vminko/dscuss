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
#include "util.h"
#include "topic.h"

#define TAG_REGEXP "[a-zA-Z0-9_]+"

/* Regexp for validating string representation of a topic. */
static GRegex* topic_regex = NULL;

/* Regexp for extracting tag from topic string. */
static GRegex* tag_extr_regex = NULL;

/* Regexp for validating single tags. */
static GRegex* tag_valid_regex = NULL;


static gboolean
topic_find_tag (const DscussTopic* topic, const gchar* tag, guint* index)
{
  guint i;

  g_assert (topic != NULL);
  g_assert (tag != NULL);

  for (i = 0; i < ((GPtrArray*) topic)->len; i++)
    {
      if (g_strcmp0 (g_ptr_array_index ((GPtrArray*) topic, i), tag) == 0)
        {
          if (index != NULL)
            *index = i;
          return TRUE;
        }
    }

  return FALSE;
}


static gint
topic_tag_compare (const gchar **a, const gchar **b)
{
  return g_strcmp0 (*a, *b);
}


void
dscuss_topic_cache_init ()
{
  g_debug ("Initializing topic cache.");
  topic_regex = g_regex_new ("^ *(" TAG_REGEXP ", *)*" TAG_REGEXP " *$", 0, 0, NULL);
  tag_extr_regex = g_regex_new (TAG_REGEXP, 0, 0, NULL);
  tag_valid_regex = g_regex_new ("^" TAG_REGEXP "$", 0, 0, NULL);
}


void
dscuss_topic_cache_uninit ()
{
  g_debug ("Uninitializing topic cache.");
  dscuss_free_non_null (topic_regex, g_regex_unref);
  dscuss_free_non_null (tag_extr_regex, g_regex_unref);
  dscuss_free_non_null (tag_valid_regex, g_regex_unref);
}


DscussTopic*
dscuss_topic_new_empty (void)
{
  GPtrArray* topic = NULL;
  topic = g_ptr_array_new ();
  g_ptr_array_set_free_func ((GPtrArray*) topic, g_free);
  return topic;
}


DscussTopic*
dscuss_topic_new (const gchar* topic_str)
{
  GPtrArray* topic = NULL;
  GMatchInfo *match_info;

  g_assert (topic_str != NULL);

  if (!g_regex_match (topic_regex, topic_str, 0, NULL))
    {
      g_warning ("This is not a valid topic string: '%s'",
                 topic_str);
      return NULL;
    }

  topic = (GPtrArray*)dscuss_topic_new_empty();
  g_regex_match (tag_extr_regex, topic_str, 0, &match_info);
  while (g_match_info_matches (match_info))
    {
      gchar* tag = g_match_info_fetch (match_info, 0);
      g_debug ("Found the following tag: '%s'", tag);

      if (topic_find_tag (topic, tag, NULL))
        {
          g_warning ("Duplicated tag found: '%s', ignoring it.", tag);
          g_free (tag);
          goto next;
        }
      g_ptr_array_add (topic, tag);

next:
      g_match_info_next (match_info, NULL);
    }
  g_match_info_free (match_info);

  g_ptr_array_sort (topic, (GCompareFunc) topic_tag_compare);

  return topic;
}


void
dscuss_topic_free (DscussTopic* topic)
{
  if (topic == NULL)
    return;

  g_ptr_array_unref ((GPtrArray*) topic);
}


DscussTopic*
dscuss_topic_copy (DscussTopic* topic)
{
  g_assert (topic != NULL);

  g_ptr_array_ref ((GPtrArray*) topic);
  return topic;
}


gchar*
dscuss_topic_to_string (const DscussTopic* topic)
{
  g_assert (topic != NULL);

  GPtrArray* topic2 = (GPtrArray*)topic;

  return dscuss_strnjoinv (", ", (gchar**)topic2->pdata, topic2->len);
}


gboolean
dscuss_topic_add_tag (DscussTopic* topic, const gchar* tag)
{
  g_assert (topic != NULL);
  g_assert (tag != NULL);

  if (!g_regex_match (tag_valid_regex, tag, 0, NULL))
    {
      g_debug ("Attempt to add invalid tag: '%s'", tag);
      return FALSE;
    }

  if (topic_find_tag (topic, tag, NULL))
    {
      g_debug ("Attempt to add duplicate tag: '%s'", tag);
      return FALSE;
    }

  g_ptr_array_add ((GPtrArray*) topic, g_strdup (tag));
  g_ptr_array_sort ((GPtrArray*) topic, (GCompareFunc) topic_tag_compare);

  return TRUE;
}


gboolean
dscuss_topic_remove_tag (DscussTopic* topic, const gchar* tag)
{
  guint index;

  g_assert (topic != NULL);
  g_assert (tag != NULL);

  if (!topic_find_tag (topic, tag, &index))
    return FALSE;

  g_ptr_array_remove_index (topic, index);

  return TRUE;
}


gboolean
dscuss_topic_contains_topic (const DscussTopic* main_topic,
                             const DscussTopic* sub_topic)
{
  guint i;

  g_assert (main_topic != NULL);
  g_assert (sub_topic != NULL);

  for (i = 0; i < ((GPtrArray*) main_topic)->len; i++)
    {
      if (!topic_find_tag (sub_topic,
                          g_ptr_array_index ((GPtrArray*) main_topic, i),
                          NULL))
        {
          return FALSE;
        }
    }

  return TRUE;
}


gboolean
dscuss_topic_is_empty (DscussTopic* topic)
{
  g_assert (topic != NULL);

  return (((GPtrArray*) topic)->len == 0);
}


gint
dscuss_topic_compare (const DscussTopic* topic1, const DscussTopic* topic2)
{
  gint res = 0;
  gchar* topic1_str = dscuss_topic_to_string (topic1);
  gchar* topic2_str = dscuss_topic_to_string (topic2);
  res = g_strcmp0 (topic1_str, topic2_str);
  g_free (topic1_str);
  g_free (topic2_str);
  return res;
}


void
dscuss_topic_foreach (const DscussTopic* topic,
                      DscussTopicIteratorCallback callback,
                      gpointer user_data)
{
  g_assert (topic != NULL);
  g_assert (callback != NULL);

  g_ptr_array_foreach ((GPtrArray*) topic,
                       (GFunc) callback,
                       user_data);
}
