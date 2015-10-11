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
 * @file include/topic.h  Topic is a group of tags, describing subject of a message or
 * area of user's interests
 * @brief  Topic tags are sorted in alphabetical order and match the following
 * pattern: [a-zA-Z0-9_].
 */

#ifndef DSCUSS_INCLUDE_TOPIC_H
#define DSCUSS_INCLUDE_TOPIC_H

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Handle for a topic.
 */
typedef void DscussTopic;

/**
 * Callback used for iterating over topics.
 *
 * @param tag        A tag.
 * @param user_data  The user data.
 */
typedef void (*DscussTopicIteratorCallback)(const gchar* tag,
                                            gpointer user_data);

/**
 * Create an empty topic.
 *
 * @return  Newly created topic.
 */
DscussTopic*
dscuss_topic_new_empty (void);

/**
 * Create a new topic from its string representation.
 *
 * @param topic_str  Comma-separated string of tags.
 *
 * @return  Newly created topic or @c NULL on error.
 */
DscussTopic*
dscuss_topic_new (const gchar* topic_str);

/**
 * Destroy a topic. Frees all memory allocated by the topic.
 *
 * @param topic  Topic to destroy.
 */
void
dscuss_topic_free (DscussTopic* topic);

/**
 * Copy a topic. Increments reference count to the object.
 *
 * @param topic  Topic to copy.
 *
 * @return  Copy of the topic.
 */
DscussTopic*
dscuss_topic_copy (DscussTopic* topic);

/**
 * Convert a topic to a string of comma-separated tags.
 *
 * @param topic  The topic to convert.
 *
 * @return  String representation of the topic.
 */
gchar*
dscuss_topic_to_string (const DscussTopic* topic);

/**
 * Add new tag to the topic.
 *
 * @param topic  The topic to modify.
 * @param tag    The tag to add to the topic.
 *
 * @return  @c TRUE in case of success, @c FALSE on error (the tag contains
 *          invalid characters or is already in the topic).
 */
gboolean
dscuss_topic_add_tag (DscussTopic* topic, const gchar* tag);

/**
 * Remove a tag from the topic.
 *
 * @param topic  The topic to modify.
 * @param tag    The tag to remove from the topic.
 *
 * @return  @c TRUE in case of success, @c FALSE on error (no such tag in the
 *          topic).
 */
gboolean
dscuss_topic_remove_tag (DscussTopic* topic, const gchar* tag);

/**
 * Shows whether one topic contains another.
 *
 * Topic A contains topic B if the list of tags of the topic B contains all
 * tags of the topic A.
 *
 * @param main_topic  A topic.
 * @param sub_topic   Another topic.
 *
 * @return  @c TRUE if @a main_topic contains @a sub_topic,
 *          @c FALSE otherwise.
 */
gboolean
dscuss_topic_contains_topic (const DscussTopic* main_topic,
                             const DscussTopic* sub_topic);

/**
 * Shows whether a topic is empty.
 *
 * @param topic  A topic.
 *
 * @return  @c TRUE if @a topic is empty, @c FALSE otherwise.
 */
gboolean
dscuss_topic_is_empty (DscussTopic* topic);

/**
 * Compares two topics.
 *
 * @param topic1      A topic.
 * @param topic2  Topic to compare with.
 *
 * @return  -1 if @a topic1 < @a topic2 ;
 *           0 if @a topic1 = @a topic2;
 *           1 if @a topic1 > @a topic2.
 */
gint
dscuss_topic_compare (const DscussTopic* topic1, const DscussTopic* topic2);

/**
 * Calls a function for each tag of a topic.
 *
 * @param topic      The topic to iterate.
 * @param callback   Function to call.
 * @param user_data  User data to be passed to the @c callback.
 */
void
dscuss_topic_foreach (const DscussTopic* topic,
                      DscussTopicIteratorCallback callback,
                      gpointer user_data);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_INCLUDE_TOPIC_H */