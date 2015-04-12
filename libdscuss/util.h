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
 * @file util.h  Utility functions.
 */

#ifndef DSCUSS_UTIL_H
#define DSCUSS_UTIL_H

#include <glib.h>


#ifdef __cplusplus
extern "C" {
#endif


/**
 * Free the memory pointed to by ptr if ptr is not NULL
 * and sets the pointer to NULL.
 *
 * @param ptr   address of the memory to free.
 * @param func  function to call to free the memory.
 */
#define dscuss_free_non_null(ptr, func) \
  do { if (ptr != NULL) { func (ptr); ptr = NULL; } \
     } while(0)

/**
 * Converts a 64-bit integer value from host to network byte order.
 *
 * @param val  a 64-bit integer value in host byte order.
 */
#define dscuss_htonll(val) (GUINT64_TO_BE (val))

/**
 * Converts a 64-bit integer value from network to host byte order.
 *
 * @param val  a 64-bit integer value in network byte order.
 */
#define dscuss_ntohll(val) (GUINT64_FROM_BE (val))

/**
 * Initializes the utility subsystem. It must be initialized before any other
 * Dscuss subsystems.
 */
void
dscuss_util_init (const gchar* data_dir);

/**
 * Uninitializes the utility subsystem.
 */
void
dscuss_util_uninit (void);

/**
 * Returns the directory containing Dscuss data files.
 * Default value is ~/.dscuss.
 */
const gchar*
dscuss_util_get_data_dir (void);

/**
 * Converts a binary data to a hexademical string.
 *
 * @param data      Pointer to the data to convert.
 * @param data_len  Length of the data.
 * @param buffer    Optional buffer for writing the result,
 *                  must be at least @c data_len * 2 + 1 bytes in size.
 *                  If no buffer is specified, the memory for the string
 *                  will be allocated.
 *
 * @return  Address of the hexademical string (@a buffer of newly allocated
 *          string if buffer was not specified).
 * the data.
 */
gchar*
dscuss_data_to_hex (const gpointer data, gsize data_len, gchar* buffer);

/**
 * Joins a number of strings together to form one long string, with a
 * separator inserted between each of them.
 *
 * @param separator      A string to insert between each of the strings.
 * @param str_array      Array of strings to join.
 * @param str_array_len  Length of the @a str_array_len.
 *
 * @return  A newly-allocated string containing all of the strings joined
 * together (should be freed with g_free()).
 */
gchar *
dscuss_strnjoinv (const gchar *separator,
                  gchar **str_array,
                  gsize str_array_len);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_UTIL_H */
