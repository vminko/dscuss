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

#include <sqlite3.h>
#include "config.h"
#include "util.h"
#include "db.h"


/* SQLite database handle. */
static sqlite3* sql_dbh = NULL;

/* Precompiled SQL statement for getting user. */
//static sqlite3_stmt* sql_stmt_get_user = NULL;


static gboolean
db_sqlite3_exec (const gchar* sql_cmd)
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


gboolean
dscuss_db_init ()
{
  gchar* db_filename = NULL;
  gint res = 0;

  db_filename = g_build_filename (dscuss_util_get_data_dir (),
                                  "db", NULL);
  res = sqlite3_open (db_filename, &sql_dbh);
  g_free (db_filename);
  if (SQLITE_OK != res)
    {
      g_critical ("Unable to initialize SQLite: %s.", sqlite3_errmsg (sql_dbh));
      goto error;
    }

  if (!db_sqlite3_exec ("PRAGMA temp_store=MEMORY") ||
      !db_sqlite3_exec ("PRAGMA synchronous=OFF") ||
      !db_sqlite3_exec ("PRAGMA locking_mode=EXCLUSIVE") ||
      !db_sqlite3_exec ("PRAGMA page_size=4092"))
    {
      /* db_sqlite3_exec logs error messages. */
      goto error;
    }

  if (!db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS user ("
                        "  id              BLOB PRIMARY KEY,"
                        "  public_key      BLOB NOT NULL,"
                        "  proof_of_work   BLOB NOT NULL,"
                        "  nickname        TEXT NOT NULL,"
                        "  additional_info TEXT,"
                        "  timestamp       INTEGER NOT NULL,"
                        "  signature       BLOB NOT NULL)") ||
      !db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS  message ("
                        "  id              BLOB PRIMARY KEY,"
                        "  subject         TEXT,"
                        "  content         TEXT,"
                        "  timestamp       INTEGER NOT NULL,"
                        "  author_id       BLOB NOT NULL,"
                        "  in_reply_to     BLOB NOT NULL,"
                        "  signature       BLOB NOT NULL,"
                        "  FOREIGN KEY (author_id) REFERENCES user(id))") ||
      !db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS  operation ("
                        "  id              BLOB PRIMARY KEY,"
                        "  type            INTEGER NOT NULL,"
                        "  reason          INTEGER NOT NULL,"
                        "  comment         TEXT,"
                        "  author_id       BLOB NOT NULL,"
                        "  timestamp       INTEGER NOT NULL,"
                        "  signature       BLOB NOT NULL,"
                        "  FOREIGN KEY (author_id) REFERENCES user(id))") ||
      !db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS  operation_on_user ("
                        "  operation_id    BLOB NOT NULL,"
                        "  user_id         BLOB NOT NULL,"
                        "  FOREIGN KEY (operation_id) REFERENCES operation(id),"
                        "  FOREIGN KEY (user_id) REFERENCES user(id))") ||
      !db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS  operation_on_message ("
                        "  operation_id    BLOB NOT NULL,"
                        "  message_id      BLOB NOT NULL,"
                        "  FOREIGN KEY (operation_id) REFERENCES operation(id),"
                        "  FOREIGN KEY (message_id) REFERENCES message(id))") ||
      !db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS  tag ("
                        "  id              INTEGER PRIMARY KEY AUTOINCREMENT,"
                        "  name            TEXT NOT NULL UNIQUE)") ||
      !db_sqlite3_exec ("CREATE TABLE IF NOT EXISTS  message_tag ("
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

  return TRUE;

error:
  dscuss_db_uninit ();
  return FALSE;
}


void
dscuss_db_uninit ()
{
  gint res;
  sqlite3_stmt* stmt;

  g_debug ("Uninitializing the database subsystem.");

  //if (sql_stmt_get_user != NULL)
  //  sqlite3_finalize (sql_stmt_get_user);

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
      sql_dbh = NULL;
    }
}
