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

/**
 * @file subscriptions.h  Provides API for managing topics, to which the user
 * is subscribed.
 */

#ifndef DSCUSS_SUBSCRIPTIONS_H
#define DSCUSS_SUBSCRIPTIONS_H

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Reads user subscriptions from a file.
 *
 * @param filename  Name of the file to read subsciptions from.
 *
 * @return  The list of topics, which the user is subscribed to
 *          or @c NULL on error (empty list is an error).
 */
GSList*
dscuss_subscriptions_read (const gchar* filename);

/**
 * Destroys a list of user subscriptions.
 *
 * @param subscriptions  Subsciptions to be destroyed.
 */
void
dscuss_subscriptions_free (GSList* subscriptions);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_SUBSCRIPTIONS_H */
