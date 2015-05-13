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
 * @file topic.h  Internal API for topics.
 */

#ifndef DSCUSS_TOPIC_H
#define DSCUSS_TOPIC_H

#include <glib.h>
#include "include/topic.h"

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Initializes topic cache. The cache contains regular expressions used for
 * validating topic strings and extracting single tags.
 */
void
dscuss_topic_cache_init (void);

/**
 * Uninitializes topic cache. Frees allocated regular expressions.
 */
void
dscuss_topic_cache_uninit (void);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_TOPIC_H */
