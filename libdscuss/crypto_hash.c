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

#include "crypto_hash.h"


gint
dscuss_crypto_hash_get_bit (const DscussHash* hash,
                            guint bit)
{
  g_assert (bit < 8 * sizeof (DscussHash));
  return (((unsigned char *) hash)[bit >> 3] & (1 << (bit & 7))) > 0;
}


guint
dscuss_crypto_hash_count_leading_zeroes (const DscussHash* hash)
{
  guint hash_count = 0;
  while ((0 == dscuss_crypto_hash_get_bit (hash, hash_count)))
    hash_count++;
  return hash_count;
}
