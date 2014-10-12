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
 * @file config.h  Used to get parameters from the configuration file and save
 * them to the file.
 */

#ifndef DSCUSS_CONFIG_H
#define DSCUSS_CONFIG_H

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif


/**
 * Initializes the configuration subsystem.
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_config_init (void);

/**
 * Uninitializes the configuration subsystem.
 */
void
dscuss_config_uninit (void);

/**
 * Requests value of an integer parameter from the config file.
 *
 * @param group          A parameter group name.
 * @param param          A parameter name.
 * @param default_value  Default value to use in case the parameter was not found
 *                       or the value not be parsed.
 * @return parameter value in case of success, or @c default_value on error.
 */
gint
dscuss_config_get_integer (const gchar *group,
                           const gchar *param,
                           gint default_value);


#ifdef __cplusplus
}
#endif


#endif /* DSCUSS_CONFIG_H */
