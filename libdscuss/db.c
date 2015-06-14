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

#include <string.h>
#include "config.h"
#include "util.h"
#include "topic.h"
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
                        "CREATE TABLE IF NOT EXISTS  User ("
                        "  Id              BLOB PRIMARY KEY,"
                        "  Public_key      BLOB NOT NULL,"
                        "  Proof           UNSIGNED BIG INT NOT NULL,"
                        "  Nickname        TEXT NOT NULL,"
                        "  Info            TEXT,"
                        "  Timestamp       INTEGER NOT NULL,"
                        "  Signature_len   INTEGER NOT NULL,"
                        "  Signature       BLOB NOT NULL)") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  Message ("
                        "  Id              BLOB PRIMARY KEY,"
                        "  Subject         TEXT,"
                        "  Content         TEXT,"
                        "  Timestamp       UNSIGNED BIG INT NOT NULL,"
                        "  Author_id       BLOB NOT NULL,"
                        "  In_reply_to     BLOB NOT NULL,"
                        "  Signature_len   INTEGER NOT NULL,"
                        "  Signature       BLOB NOT NULL,"
                        "  FOREIGN KEY (Author_id) REFERENCES User(Id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  Operation ("
                        "  Id              BLOB PRIMARY KEY,"
                        "  Type            INTEGER NOT NULL,"
                        "  Reason          INTEGER NOT NULL,"
                        "  Comment         TEXT,"
                        "  Author_id       BLOB NOT NULL,"
                        "  Timestamp       UNSIGNED BIG INT NOT NULL,"
                        "  Signature_len   INTEGER NOT NULL,"
                        "  Signature       BLOB NOT NULL,"
                        "  FOREIGN KEY (Author_id) REFERENCES User(Id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  Operation_on_User ("
                        "  Operation_id    BLOB NOT NULL,"
                        "  User_id         BLOB NOT NULL,"
                        "  FOREIGN KEY (Operation_id) REFERENCES Operation(Id),"
                        "  FOREIGN KEY (User_id) REFERENCES User(Id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  Operation_on_Message ("
                        "  Operation_id    BLOB NOT NULL,"
                        "  Message_id      BLOB NOT NULL,"
                        "  FOREIGN KEY (Operation_id) REFERENCES Operation(Id),"
                        "  FOREIGN KEY (Message_id) REFERENCES Message(Id))") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  Tag ("
                        "  Id              INTEGER PRIMARY KEY AUTOINCREMENT,"
                        "  Name            TEXT NOT NULL UNIQUE ON CONFLICT IGNORE)") ||
      !db_sqlite3_exec (sql_dbh,
                        "CREATE TABLE IF NOT EXISTS  Message_Tag ("
                        "  Tag_id          INTEGER NOT NULL,"
                        "  Message_id      BLOB NOT NULL,"
                        "  FOREIGN KEY (Tag_id) REFERENCES Tag(Id),"
                        "  FOREIGN KEY (Message_id) REFERENCES Message(Id),"
                        "  UNIQUE (Tag_id, Message_id))"))
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
dscuss_db_put_user (DscussDb* dbh, DscussUser* user)
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
          "INSERT INTO User "
          "( Id,"
          "  Public_key,"
          "  Proof,"
          "  Nickname,"
          "  Info,"
          "  Timestamp,"
          "  Signature_len,"
          "  Signature) "
          "VALUES (?, ?, ?, ?, ?, ?, ?, ?)", &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `put_user' statement with error: %s.",
                 sqlite3_errmsg (dbh));
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
       (SQLITE_OK != sqlite3_bind_int   (stmt, 7,
                                         dscuss_user_get_signature_length (user))) ||
       (SQLITE_OK != sqlite3_bind_blob  (stmt, 8,
                                         dscuss_user_get_signature (user),
                                         sizeof (struct DscussSignature),
                                         SQLITE_TRANSIENT)) )
    {
      g_warning ("Failed to bind parameters to `put_user' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }

  if (SQLITE_DONE != sqlite3_step (stmt))
    {
      g_warning ("Failed to execute `put_user' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  result = TRUE;

out:
  dscuss_free_non_null (pubkey_digest, g_free);
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `put_user' statement with error: %s.",
                 sqlite3_errmsg (dbh));
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
                  "SELECT Public_key,"
                  "       Proof,"
                  "       Nickname,"
                  "       Info,"
                  "       Timestamp,"
                  "       Signature_len,"
                  "       Signature "
                  "FROM User WHERE Id=?",
                  &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `gut_user' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }
  if (SQLITE_OK != sqlite3_bind_blob (stmt, 1, id, sizeof (DscussHash),
                      SQLITE_TRANSIENT))
    {
      g_warning ("Failed to bind parameters to `get_user' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }
  if (SQLITE_ROW != sqlite3_step (stmt))
    {
      g_debug ("No such user in the database.");
      goto out;
    }
  if ( (sqlite3_column_bytes (stmt, 6) !=
        sizeof (struct DscussSignature)) )
    {
      g_warning ("Database is corrupted: wrong size of user.signature.");
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
                          (struct DscussSignature *) sqlite3_column_blob  (stmt, 6),
                          sqlite3_column_int (stmt, 5));
  if (user == NULL)
    g_warning ("Failed to create a user entity.");

out:
  dscuss_free_non_null (datetime, g_date_time_unref);
  dscuss_free_non_null (pubkey,
                        dscuss_crypto_public_key_free);
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `get_user' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }

  return user;
}


/*** IterateMessageTagsContext ***********************************************/

typedef struct IterateMessageTagsContext
{
  DscussDb* dbh;
  const DscussHash* msg_id;
} IterateMessageTagsContext;


static IterateMessageTagsContext*
iterate_message_tags_context_new (DscussDb* dbh,
                                  const DscussHash* msg_id)
{
  IterateMessageTagsContext* ctx = g_new0 (IterateMessageTagsContext, 1);
  ctx->dbh = dbh;
  ctx->msg_id = msg_id;
  return ctx;
}


static void
iterate_message_tags_context_free (IterateMessageTagsContext* ctx)
{
  g_assert (ctx != NULL);
  g_free (ctx);
}

/*** End of IterateMessageTagsContext ****************************************/


static gboolean
db_put_tag (DscussDb* dbh, const gchar* tag)
{
  sqlite3_stmt *stmt;
  gboolean result = FALSE;

  g_assert (dbh != NULL);
  g_assert (tag != NULL);

  g_debug ("Adding tag `%s' to the database.", tag);

  if (db_sqlite3_prepare (dbh,
          "INSERT INTO Tag (Name) VALUES (?)",
          &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `put_tag' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  if (SQLITE_OK != sqlite3_bind_text  (stmt, 1, tag,
                                       -1, SQLITE_TRANSIENT))
    {
      g_warning ("Failed to bind parameters to `put_tag' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }

  if (SQLITE_DONE != sqlite3_step (stmt))
    {
      g_warning ("Failed to execute `put_tag' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  result = TRUE;

out:
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `put_tag' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }

  return result;
}


static gboolean
db_put_message_tag (DscussDb* dbh,
                    const gchar* tag,
                    const DscussHash* message_id)
{
  sqlite3_stmt *stmt;
  gboolean result = FALSE;

  g_assert (dbh != NULL);
  g_assert (tag != NULL);
  g_assert (message_id != NULL);

  g_debug ("Adding tag `%s' for the message `%s' to the database.",
           tag, dscuss_crypto_hash_to_string (message_id));

  if (db_sqlite3_prepare (dbh,
          "INSERT INTO Message_Tag "
          "( Message_id, Tag_id ) "
          "VALUES (?, (SELECT Id FROM Tag WHERE Name=?))", &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `put_message_tag' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  if ( (SQLITE_OK != sqlite3_bind_blob  (stmt, 1, message_id,
                                         sizeof (DscussHash),
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_text  (stmt, 2, tag, -1,
                                         SQLITE_TRANSIENT)) )
    {
      g_warning ("Failed to bind parameters to `put_message_tag' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }

  if (SQLITE_DONE != sqlite3_step (stmt))
    {
      g_warning ("Failed to execute `put_message_tag' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  result = TRUE;

out:
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `put_message_tag' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }

  return result;
}


static void
db_iterate_message_tags (const gchar* tag,
                         gpointer user_data)
{
  IterateMessageTagsContext* ctx = user_data;

  if (!db_put_tag (ctx->dbh, tag))
    {
      g_critical ("Failed to store tag '%s' in the DB. DB may be corrupted!",
                  tag);
      return;
    }

  if (!db_put_message_tag (ctx->dbh, tag, ctx->msg_id))
    {
      g_critical ("Failed to store tag '%s' in the DB. DB may be corrupted!",
                  tag);
      return;
    }
}


static void
db_put_message_topic (DscussDb* dbh,
                      const DscussMessage* msg)
{
  IterateMessageTagsContext* ctx = NULL;

  ctx = iterate_message_tags_context_new (dbh,
                                          dscuss_message_get_id (msg));
  dscuss_topic_foreach (dscuss_message_get_topic (msg),
                        db_iterate_message_tags,
                        ctx);
  iterate_message_tags_context_free (ctx);
}


static DscussTopic*
db_get_message_topic (DscussDb* dbh,
                      const DscussHash* msg_id)
{
  sqlite3_stmt *stmt;
  DscussTopic* topic = NULL;

  g_assert (msg_id != NULL);
  g_debug ("Fetching message topic from the database.");

  if (db_sqlite3_prepare (dbh,
                  "SELECT Name "
                  "FROM Tag "
                  "INNER JOIN Message_Tag "
                  "ON Tag.Id = Message_Tag.Tag_id AND Message_Tag.Message_id = ?",
                  &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `get_message_topic' statement with"
                 " error: %s.", sqlite3_errmsg (dbh));
      return NULL;
    }

  if ( (SQLITE_OK != sqlite3_bind_blob  (stmt, 1,
                                         msg_id,
                                         sizeof (DscussHash),
                                         SQLITE_TRANSIENT)) )
    {
      g_warning ("Failed to bind parameters to `get_message_topic' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }

  topic = dscuss_topic_new_empty ();
  while (SQLITE_ROW == sqlite3_step (stmt))
    {
      g_debug ("Found a message tag matching the request.");
      dscuss_topic_add_tag (topic,
                            (const gchar*) sqlite3_column_text  (stmt, 0));
    }

out:
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `get_recent_messages' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }

  if (dscuss_topic_is_empty (topic))
    {
      dscuss_topic_free (topic);
      topic = NULL;
    }

  return topic;
}


gboolean
dscuss_db_put_message (DscussDb* dbh, DscussMessage* msg)
{
  sqlite3_stmt *stmt;
  gboolean result = FALSE;
  gint64 timestamp = 0;
  DscussHash parent_id;

  g_assert (dbh != NULL);
  g_assert (msg != NULL);

  g_debug ("Adding message `%s' to the database.",
           dscuss_message_get_description (msg));

  timestamp = g_date_time_to_unix (dscuss_message_get_datetime (msg));
  if (db_sqlite3_prepare (dbh,
          "INSERT INTO Message "
          "( Id,"
          "  Subject,"
          "  Content,"
          "  Timestamp,"
          "  Author_id,"
          "  In_reply_to,"
          "  Signature_len,"
          "  Signature) "
          "VALUES (?, ?, ?, ?, ?, ?, ?, ?)", &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `put_message' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  memset (&parent_id, 0, sizeof (DscussHash));

  if ( (SQLITE_OK != sqlite3_bind_blob  (stmt, 1,
                                         dscuss_message_get_id (msg),
                                         sizeof (DscussHash),
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_text  (stmt, 2,
                                         dscuss_message_get_subject (msg),
                                         -1,
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_text  (stmt, 3,
                                         dscuss_message_get_content (msg),
                                         -1, SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_int64 (stmt, 4,
                                         timestamp)) ||
       (SQLITE_OK != sqlite3_bind_blob  (stmt, 5,
                                         dscuss_message_get_author_id (msg),
                                         sizeof (DscussHash),
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_blob  (stmt, 6,
                                         &parent_id,
                                         sizeof (DscussHash),
                                         SQLITE_TRANSIENT)) ||
       (SQLITE_OK != sqlite3_bind_int   (stmt, 7,
                                         dscuss_message_get_signature_length (msg))) ||
       (SQLITE_OK != sqlite3_bind_blob  (stmt, 8,
                                         dscuss_message_get_signature (msg),
                                         sizeof (struct DscussSignature),
                                         SQLITE_TRANSIENT)) )
    {
      g_warning ("Failed to bind parameters to `put_message' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }

  if (SQLITE_DONE != sqlite3_step (stmt))
    {
      g_warning ("Failed to execute `put_message' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }

  result = TRUE;

  /* TBD: rollback DB state in case of a failure? */
  db_put_message_topic (dbh, msg);

out:
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `put_message' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }

  return result;
}


void
dscuss_db_get_recent_messages (DscussDb* dbh,
                               DscussDbIterateMessageCallback callback,
                               gpointer user_data)
{
  sqlite3_stmt *stmt;
  GDateTime* datetime = NULL;
  DscussMessage* msg = NULL;
  DscussTopic* topic = NULL;

  g_assert (dbh != NULL);
  g_assert (callback != NULL);

  g_debug ("Fetching latest messages from the database.");

  if (db_sqlite3_prepare (dbh,
                  "SELECT Subject,"
                  "       Content,"
                  "       Timestamp,"
                  "       Author_id,"
                  "       Signature_len,"
                  "       Signature, "
                  "       Id "
                  "FROM Message "
                  "ORDER BY Timestamp DESC",
                  &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `get_recent_messages' statement with"
                 " error: %s.", sqlite3_errmsg (dbh));
      callback (FALSE, NULL, user_data);
      return;
    }

  while (SQLITE_ROW == sqlite3_step (stmt))
    {
      g_debug ("Found a message matching the request.");
      if ( (sqlite3_column_bytes (stmt, 3) !=
            sizeof (DscussHash)) )
        {
          g_warning ("Database is corrupted: wrong size of message.author_id.");
          continue;
        }
      if ( (sqlite3_column_bytes (stmt, 5) !=
            sizeof (struct DscussSignature)) )
        {
          g_warning ("Database is corrupted: wrong size of message.signature.");
          continue;
        }
      topic = db_get_message_topic (dbh,
                                    (DscussHash *) sqlite3_column_blob (stmt, 6));
      if (topic == NULL)
        {
          g_warning ("Database is corrupted: failed to fetch message topic.");
          continue;
        }
      datetime = g_date_time_new_from_unix_utc (sqlite3_column_int64 (stmt, 2));
      msg = dscuss_message_new_full (
                topic,
                (const gchar*) sqlite3_column_text  (stmt, 0),             /* subject */
                (const gchar*) sqlite3_column_text  (stmt, 1),             /* content */
                (DscussHash *) sqlite3_column_blob  (stmt, 3),             /* author_id */
                datetime,
                (struct DscussSignature *) sqlite3_column_blob  (stmt, 5), /* signature */
                sqlite3_column_int (stmt, 4));                             /* signature_len */
      callback (TRUE, msg, user_data);
      dscuss_free_non_null (topic, dscuss_topic_free);
      dscuss_free_non_null (datetime, g_date_time_unref);
    }
  callback (TRUE, NULL, user_data);

  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `get_recent_messages' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }
}


DscussMessage*
dscuss_db_get_message (DscussDb* dbh, const DscussHash* id)
{
  sqlite3_stmt *stmt;
  GDateTime* datetime = NULL;
  DscussTopic* topic = NULL;
  DscussMessage* msg = NULL;

  g_assert (dbh != NULL);
  g_assert (id != NULL);

  g_debug ("Fetching message with id `%s' from the database.",
           dscuss_crypto_hash_to_string (id));
  if (db_sqlite3_prepare (dbh,
                  "SELECT Subject,"
                  "       Content,"
                  "       Timestamp,"
                  "       Author_id,"
                  "       Signature_len,"
                  "       Signature "
                  "FROM Message WHERE Id=?",
                  &stmt) != SQLITE_OK)
    {
      g_warning ("Failed to prepare `get_message' statement with error: %s.",
                 sqlite3_errmsg (dbh));
      goto out;
    }
  if (SQLITE_OK != sqlite3_bind_blob (stmt, 1, id, sizeof (DscussHash),
                      SQLITE_TRANSIENT))
    {
      g_warning ("Failed to bind parameters to `get_message' statement"
                 " with error: %s.", sqlite3_errmsg (dbh));
      goto out;
    }
  if (SQLITE_ROW != sqlite3_step (stmt))
    {
      g_debug ("No such message in the database.");
      goto out;
    }
  g_debug ("Found a message matching the request.");
  if ( (sqlite3_column_bytes (stmt, 3) != sizeof (DscussHash)) )
    {
      g_warning ("Database is corrupted: wrong size of message.author_id.");
      goto out;
    }
  if ( (sqlite3_column_bytes (stmt, 5) !=
        sizeof (struct DscussSignature)) )
    {
      g_warning ("Database is corrupted: wrong size of message.signature.");
      goto out;
    }
  topic = db_get_message_topic (dbh, id);
  if (topic == NULL)
    {
      g_warning ("Database is corrupted: failed to fetch message topic.");
      goto out;
    }
  datetime = g_date_time_new_from_unix_utc (sqlite3_column_int64 (stmt, 2));
  msg = dscuss_message_new_full (
            topic,
            (const gchar*) sqlite3_column_text  (stmt, 0),             /* subject */
            (const gchar*) sqlite3_column_text  (stmt, 1),             /* content */
            (DscussHash *) sqlite3_column_blob  (stmt, 3),             /* author_id */
            datetime,
            (struct DscussSignature *) sqlite3_column_blob  (stmt, 5), /* signature */
            sqlite3_column_int (stmt, 4));                             /* signature_len */

out:
  dscuss_free_non_null (topic, dscuss_topic_free);
  dscuss_free_non_null (datetime, g_date_time_unref);
  if (SQLITE_OK != sqlite3_finalize (stmt))
    {
      g_warning ("Failed to finalize `get_message' statement with error: %s.",
                 sqlite3_errmsg (dbh));
    }

  return msg;
}
