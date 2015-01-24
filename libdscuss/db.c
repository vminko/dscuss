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

#include <string.h>
#include "config.h"
#include "util.h"
#include "db.h"


static gboolean
db_sqlite3_exec (sqlite3* sql_dbh, const gchar* sql_cmd)
{
  char* sql_err_msg = NULL;
  if (SQLITE_OK != sqlite3_exec (sql_dbh, sql_cmd, NULL, NULL, &sql_err_msg))
    {
      g_critical ("`sqlite3_exec' failed to execute '%s' with the following"
                  " error: %s", sql_cmd, sql_err_msg);
      sqlite3_free (sql_err_msg);
      return FALSE;
    }
  return TRUE;
}


static gint
db_sqlite3_prepare (sqlite3 * dbh,
                    const gchar *sql,
                    sqlite3_stmt ** stmt)
{
  char *dummy;
  return sqlite3_prepare (dbh,
                          sql, strlen (sql),
                          stmt, (const char **) &dummy);
}


DscussDb*
dscuss_db_open (const gchar* filename)
{
  sqlite3* sql_dbh = NULL;
  gint res = 0;

  res = sqlite3_open (filename, &sql_dbh);
  if (SQLITE_OK != res)
    {
      g_critical ("Unable to initialize SQLite: %s.", sqlite3_errmsg (sql_dbh));
      goto error;
    }

  if (!db_sqlite3_exec (sql_dbh, "PRAGMA temp_store=MEMORY") ||
      !db_sqlite3_exec (sql_dbh, "PRAGMA synchronous=OFF") ||
      !db_sqlite3_exec (sql_dbh, "PRAGMA locking_mode=EXCLUSIVE") ||
      !db_sqlite3_exec (sql_dbh, "PRAGMA page_size=4092"))
    {
      /* db_sqlite3_exec logs error messages. */
      goto error;
    }

  if (!db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS user ("
                        "  id              BLOB PRIMARY KEY,"
                        "  public_key      BLOB NOT NULL,"
                        "  proof           UNSIGNED BIG INT NOT NULL,"
                        "  nickname        TEXT NOT NULL,"
                        "  info            TEXT,"
                        "  timestamp       INTEGER NOT NULL,"
                        "  signature       BLOB NOT NULL)") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  message ("
                        "  id              BLOB PRIMARY KEY,"
                        "  subject         TEXT,"
                        "  content         TEXT,"
                        "  timestamp       UNSIGNED BIG INT NOT NULL,"
                        "  author_id       BLOB NOT NULL,"
                        "  in_reply_to     BLOB NOT NULL,"
                        "  signature       BLOB NOT NULL,"
                        "  FOREIGN KEY (author_id) REFERENCES user(id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  operation ("
                        "  id              BLOB PRIMARY KEY,"
                        "  type            INTEGER NOT NULL,"
                        "  reason          INTEGER NOT NULL,"
                        "  comment         TEXT,"
                        "  author_id       BLOB NOT NULL,"
                        "  timestamp       UNSIGNED BIG INT NOT NULL,"
                        "  signature       BLOB NOT NULL,"
                        "  FOREIGN KEY (author_id) REFERENCES user(id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  operation_on_user ("
                        "  operation_id    BLOB NOT NULL,"
                        "  user_id         BLOB NOT NULL,"
                        "  FOREIGN KEY (operation_id) REFERENCES operation(id),"
                        "  FOREIGN KEY (user_id) REFERENCES user(id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  operation_on_message ("
                        "  operation_id    BLOB NOT NULL,"
                        "  message_id      BLOB NOT NULL,"
                        "  FOREIGN KEY (operation_id) REFERENCES operation(id),"
                        "  FOREIGN KEY (message_id) REFERENCES message(id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  tag ("
                        "  id              INTEGER PRIMARY KEY AUTOINCREMENT,"
                        "  name            TEXT NOT NULL UNIQUE)") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  message_tag ("
                        "  tag_id          INTEGER NOT NULL,"
                        "  message_id      BLOB NOT NULL,"
                        "  FOREIGN KEY (tag_id) REFERENCES tag(id),"
                        "  FOREIGN KEY (message_id) REFERENCES message(id))"))
    {
      /* db_sqlite3_exec logs error messages. */
      goto error;
    }

  /* TBD: create indexes? */
  g_debug ("Database subsystem successfully initialized.");

  return (DscussDb*) sql_dbh;

error:
  dscuss_db_close ((DscussDb*) sql_dbh);
  return NULL;
}


void
dscuss_db_close (DscussDb* dbh)
{
  sqlite3* sql_dbh = (sqlite3*) dbh;
  sqlite3_stmt* stmt;
  gint res;

  g_debug ("Closing the database connection.");

  if (sql_dbh != NULL)
    {
      res = sqlite3_close (sql_dbh);
      if (res == SQLITE_BUSY)
        {
          g_warning ("Attempt to close SQLite without finalizing all prepared"
                     " statements.");
          stmt = sqlite3_next_stmt (sql_dbh, NULL); 
          while (stmt != NULL)
            {
              g_debug ("Closing statement %p", stmt);
              res = sqlite3_finalize (stmt);
              if (res != SQLITE_OK)
                g_warning ("Failed to close statement %p: %d", stmt, res);
              stmt = sqlite3_next_stmt (sql_dbh, NULL);
            }
          res = sqlite3_close (sql_dbh);
        }
      if (res != SQLITE_OK)
        g_critical ("Failed to close SQLite connection with error: %s",
                    sqlite3_errmsg (sql_dbh));
    }
}


gboolean
dscuss_db_put_user (DscussDb* dbh, const DscussUser* user)
{
  sqlite3_stmt *stmt;
  gchar* pubkey_digest = NULL;
  gsize pubkey_digest_len = 0;
  gboolean result = FALSE;
  gint64 timestamp = 0;

  g_assert (dbh != NULL);
  g_assert (user != NULL);

  g_debug ("Adding user `%s' to the database.",
           dscuss_user_get_nickname (user));


  if (!dscuss_crypto_public_key_to_der (dscuss_user_get_public_key(user),
                                        &pubkey_digest,
                                        &pubkey_digest_len))
    {
      g_warning ("Failed to serialize public key");
      goto out;
    }
  timestamp = g_date_time_to_unix (dscuss_user_get_datetime (user));

  if (db_sqlite3_prepare (dbh,
          "INSERT INTO user "
          "(id, public_key, proof, nickname, info, timestamp, signature) "
          "VALUES (?, ?, ?, ?, ?, ?, ?)", &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `put_user' statement with error: %s.",
                 sqlite3_errmsg(dbh));
      goto out;
    }

  if ( (SQLITE_OK != sqlite3_bind_blob  (stmt, 1,
                                         dscuss_user_get_id (user),
                                         sizeof (DscussHash),
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_blob  (stmt, 2,
                                         pubkey_digest,
                                         pubkey_digest_len,
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_int64 (stmt, 3,
                                         dscuss_user_get_proof (user))) ||
       (SQLITE_OK != sqlite3_bind_text  (stmt, 4,
                                         dscuss_user_get_nickname (user),
                                         -1,
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_text  (stmt, 5,
                                         dscuss_user_get_info (user),
                                         -1, SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_int64 (stmt, 6,
                                         timestamp)) ||
       (SQLITE_OK != sqlite3_bind_blob  (stmt, 7,
                                         dscuss_user_get_signature (user),
                                         sizeof (struct DscussSignature),
                                         SQLITE_TRANSIENT)) )
    {
      g_warning ("Failed to bind parameters to `put_user' statement"
                 " with error: %s.", sqlite3_errmsg(dbh));
      goto out;
    }

  if (SQLITE_DONE != sqlite3_step (stmt))
    {
      g_warning ("Failed to execute `put_user' statement with error: %s.",
                 sqlite3_errmsg(dbh));
      goto out;
    }

  result = TRUE;

out:
  dscuss_free_non_null (pubkey_digest, g_free);
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `put_user' statement with error: %s.",
                 sqlite3_errmsg(dbh));
    }

  return result;
}


DscussUser*
dscuss_db_get_user (DscussDb* dbh, const DscussHash* id)
{
  sqlite3_stmt *stmt;
  gsize pubkey_digest_len = 0;
  const gchar* pubkey_digest = NULL;
  DscussPublicKey* pubkey = NULL;
  GDateTime* datetime = NULL;
  DscussUser* user = NULL;

  g_assert (dbh != NULL);
  g_assert (id != NULL);

  g_debug ("Fetching user with id `%s' from the database.",
           dscuss_crypto_hash_to_string (id));
  if (db_sqlite3_prepare (dbh,
                  "SELECT public_key, proof, nickname, info, timestamp, signature "
                  "FROM user WHERE id=?",
                  &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `gut_user' statement with error: %s.",
                 sqlite3_errmsg(dbh));
      goto out;
    }
  if (SQLITE_OK != sqlite3_bind_blob (stmt, 1, id, sizeof (DscussHash),
                      SQLITE_TRANSIENT))
    {
      g_warning ("Failed to bind parameters to `get_user' statement"
                 " with error: %s.", sqlite3_errmsg(dbh));
      goto out;
    }
  if (SQLITE_ROW != sqlite3_step (stmt))
    {
      g_debug ("No such user in the database.");
      goto out;
    }
  if ( (sqlite3_column_bytes (stmt, 5) !=
        sizeof (struct DscussSignature)) )
    {
      g_warning ("Database is corrupted: wrong signature size.");
      goto out;
    }
  pubkey_digest_len = sqlite3_column_bytes (stmt, 0);
  if ( (pubkey_digest_len == 0) )
    {
      g_warning ("Database is corrupted: public key size is 0.");
      goto out;
    }
  pubkey_digest = sqlite3_column_blob (stmt, 0),
  pubkey = dscuss_crypto_public_key_from_der (pubkey_digest,
                                              pubkey_digest_len);
  if (pubkey == NULL)
    {
      g_warning ("Failed to parse public key.");
      goto out;
    }

  datetime = g_date_time_new_from_unix_utc (sqlite3_column_int64 (stmt, 4));
  user = dscuss_user_new (pubkey,
                          sqlite3_column_int64 (stmt, 1),
                          (const gchar*) sqlite3_column_text  (stmt, 2),
                          (const gchar*) sqlite3_column_text  (stmt, 3),
                          datetime,
                          (struct DscussSignature *) sqlite3_column_blob  (stmt, 5));
  if (user == NULL)
    g_warning ("Failed to create a user entity.");

out:
  dscuss_free_non_null (datetime, g_date_time_unref);
  dscuss_free_non_null (pubkey,
                        dscuss_crypto_public_key_free);
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `gut_user' statement with error: %s.",
                 sqlite3_errmsg(dbh));
    }

  return user;
}
