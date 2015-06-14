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
#include "network.h"
#include "crypto.h"
#include "crypto_pow.h"
#include "crypto_hash.h"
#include "db.h"
#include "util.h"
#include "topic.h"
#include "subscriptions.h"
#include "core.h"


struct LoggedUser
{
  /* The user's private key. */
  DscussPrivateKey* privkey;

  /* Handle to the connection with the user's database. */
  DscussDb* dbh;

  /* The user entity. */
  DscussUser* user;

  /* List of the topics the user is subscribed to. */
  GSList* subscriptions;

  /* List of connected peers. */
  GSList* peers;

  /* Callbacks for notification of the UI */
  DscussNewMessageCallback msg_callback;
  gpointer msg_data;
  DscussNewUserCallback user_callback;
  gpointer user_data;
  DscussNewOperationCallback oper_callback;
  gpointer oper_data;
};

static struct LoggedUser* self = NULL;


/*** FreeDuplicatePeerContext ************************************************/

typedef struct FreeDuplicatePeerContext
{
  DscussPeer* peer_to_free;
  const DscussPeer* duplicate_peer;
} FreeDuplicatePeerContext;


static FreeDuplicatePeerContext*
free_duplicate_peer_context_new (DscussPeer* peer_to_free,
                                 const DscussPeer* duplicate_peer)
{
  FreeDuplicatePeerContext* ctx = g_new0 (FreeDuplicatePeerContext, 1);
  ctx->peer_to_free = peer_to_free;
  ctx->duplicate_peer = duplicate_peer;
  return ctx;
}


static void
free_duplicate_peer_context_free (FreeDuplicatePeerContext* ctx)
{
  g_assert (ctx != NULL);
  g_free (ctx);
}

/*** End of FreeDuplicatePeerContext *****************************************/


static gboolean
free_peer (gpointer user_data)
{
  DscussPeer* peer = user_data;
  dscuss_peer_free (peer);
  return FALSE;
}


static gboolean
free_duplicate_peer (gpointer user_data)
{
  FreeDuplicatePeerContext* ctx = user_data;
  g_assert (ctx != NULL);
  dscuss_peer_free_full (ctx->peer_to_free,
                         DSCUSS_PEER_DISCONNECT_REASON_DUPLICATE,
                         (gpointer) ctx->duplicate_peer);
  free_duplicate_peer_context_free (ctx);
  return FALSE;
}



static gboolean
is_message_relevant (const GSList* subscriptions,
                     const DscussMessage* msg)
{
  g_assert (subscriptions != NULL);
  g_assert (msg != NULL);

  GSList* iterator = (GSList*) subscriptions;
  for (; iterator; iterator = iterator->next)
    {
      DscussTopic* topic = iterator->data;
      if (dscuss_topic_contains_topic (topic,
                                       dscuss_message_get_topic (msg)))
        {
          return TRUE;
        }
    }
  return FALSE;
}


static void
on_new_entity (DscussPeer* peer,
               DscussEntity* entity,
               gboolean result,
               gpointer user_data)
{
  DscussMessage* msg = NULL;

  g_assert (self != NULL);

  if (!result)
    {
      g_warning ("Failed to read from peer '%s'",
                 dscuss_peer_get_description (peer));
      self->peers = g_slist_remove (self->peers, peer);
      g_idle_add (free_peer, peer);
      return;
    }

  g_debug ("New entity from '%s' received: %s",
           dscuss_peer_get_description (peer),
           dscuss_entity_get_description (entity));
  switch (dscuss_entity_get_type (entity))
    {
    case DSCUSS_ENTITY_TYPE_USER:
      g_assert_not_reached ();
      /* TBD */
      break;

    case DSCUSS_ENTITY_TYPE_MSG:
      msg = (DscussMessage*) entity;
      if (!is_message_relevant (self->subscriptions, msg))
        {
          gchar* topic_str = dscuss_topic_to_string (dscuss_message_get_topic (msg));
          g_warning ("Peer '%s' sent an uninteresting message from the topic '%s'.",
                     dscuss_peer_get_description (peer), topic_str);
          g_free (topic_str);
          dscuss_message_free (msg);
          self->peers = g_slist_remove (self->peers, peer);
          g_idle_add (free_peer, peer);
          return;
        }
      if (!dscuss_db_put_message (self->dbh, msg))
        {
          g_critical ("Failed to store message '%s' in the database!",
                      dscuss_message_get_description (msg));
        }
      self->msg_callback (msg, self->msg_data);
      break;

    case DSCUSS_ENTITY_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }
}


static gboolean
start_receiving_entities (gpointer user_data)
{
  DscussPeer* peer = user_data;
  g_assert (peer != NULL);
  dscuss_peer_set_receive_callback (peer,
                                    on_new_entity,
                                    NULL);
  return FALSE;
}


static void
on_peer_handshaked (DscussPeer* peer,
                    gboolean result,
                    gpointer user_data)
{
  g_assert (self != NULL);
  FreeDuplicatePeerContext* free_dup_ctx = NULL;

  if (result)
    {
      g_debug ("Successfully handshaked with peer '%s'.",
               dscuss_peer_get_description (peer));

      /* Drop connection if we have already connected to this peer
       * from another address. */
      GSList* iterator = NULL;
      for (iterator = self->peers; iterator; iterator = iterator->next)
        {
          DscussPeer* ipeer = iterator->data;
          if (peer != ipeer &&
              dscuss_peer_is_handshaked (ipeer) &&
              !memcmp (dscuss_user_get_id (dscuss_peer_get_user (ipeer)),
                       dscuss_user_get_id (dscuss_peer_get_user (peer)),
                       sizeof (DscussHash)))
            {
              g_debug ("Already connected with this peer from '%s'.",
                       dscuss_peer_get_connecton_description (ipeer));
              self->peers = g_slist_remove (self->peers, peer);
              free_dup_ctx = free_duplicate_peer_context_new (peer, ipeer);
              g_idle_add (free_duplicate_peer, free_dup_ctx);
              return;
            }
        }

      /* TBD: synchronize with peer */
      g_idle_add (start_receiving_entities, peer);
    }
  else
    {
      g_warning ("Error handshaking with peer '%s'.",
                 dscuss_peer_get_description (peer));
      self->peers = g_slist_remove (self->peers, peer);
      g_idle_add (free_peer, peer);
    }
}


static void
peer_connected_cb (DscussPeer* peer,
                   gpointer user_data)
{
  g_assert (self != NULL);
  g_assert (peer != NULL);
  g_debug ("Connection with a new peer is established '%s'.",
           dscuss_peer_get_connecton_description (peer));
  self->peers = g_slist_append (self->peers, peer);

  dscuss_peer_handshake (peer,
                         self->user,
                         self->privkey,
                         self->subscriptions,
                         self->dbh,
                         on_peer_handshaked,
                         NULL);
}


gboolean
dscuss_init (const gchar* data_dir)
{
  dscuss_util_init (data_dir);
  dscuss_topic_cache_init ();

  if (!dscuss_config_init ())
    {
      g_critical ("Error initializing the configuration subsystem!");
      goto error;
    }

  return TRUE;

error:
  dscuss_uninit ();
  return FALSE;
}


void
dscuss_uninit ()
{
  g_debug ("Uninitializing Dscuss");

  if (dscuss_is_logged_in ())
    dscuss_logout ();

  dscuss_config_uninit ();
  dscuss_topic_cache_uninit ();
  dscuss_util_uninit ();
  while (g_main_context_pending (NULL))
    g_main_context_iteration (NULL, TRUE);
}

/*** RegisterContext *********************************************************/

typedef struct RegisterContext
{
  gchar* nickname;
  gchar* info;
  gchar* db_filename;
  DscussPrivateKey* privkey;
  DscussRegisterCallback callback;
  gpointer user_data;
} RegisterContext;


static RegisterContext*
register_context_new (const gchar* nickname,
                      const gchar* info,
                      const gchar* db_filename,
                      DscussPrivateKey* privkey,
                      DscussRegisterCallback callback,
                      gpointer user_data)
{
  RegisterContext* ctx = g_new0 (RegisterContext, 1);
  ctx->nickname = g_strdup (nickname);
  ctx->info = g_strdup (info);
  ctx->db_filename = g_strdup (db_filename);
  ctx->privkey = privkey;
  ctx->callback = callback;
  ctx->user_data = user_data;
  return ctx;
}


static void
register_context_free (RegisterContext* ctx)
{
  g_assert (ctx != NULL);
  g_free (ctx->nickname);
  g_free (ctx->info);
  g_free (ctx->db_filename);
  dscuss_crypto_private_key_free (ctx->privkey);
  g_free (ctx);
}

/*** End of RegisterContext **************************************************/


static void
on_pow_search_finished (gboolean result_,
                        guint64 proof_,
                        gpointer user_data)
{
  RegisterContext* ctx = (RegisterContext*) user_data;
  guint64 proof = 0;
  GDateTime* datetime = NULL;
  gboolean result = result_;
  DscussDb* dbh = NULL;
  gchar* db_filename = NULL;
  DscussUser* user = NULL;

  g_assert (ctx != NULL);

  g_debug ("The search of proof-of-work is finished with the result %d.",
           result_);
  if (result)
    proof = proof_;
  else
    {
      result = FALSE;
      goto out;
    }

  datetime = g_date_time_new_now_utc ();
  user = dscuss_user_emerge (ctx->privkey,
                             proof,
                             ctx->nickname,
                             ctx->info,
                             datetime);
  g_date_time_unref (datetime);

  db_filename = g_build_filename (dscuss_util_get_data_dir (),
                                  ctx->nickname, "db", NULL);
  dbh = dscuss_db_open (db_filename);
  if (dbh == NULL)
    {
      g_warning ("Failed to open database connection with '%s'.", db_filename);
      result = FALSE;
      goto out;
    }
  if (!dscuss_db_put_user (dbh, user))
    {
      g_warning ("Failed to store new user in the database.");
      result = FALSE;
      goto out;
    }

out:
  dscuss_free_non_null (user, dscuss_user_free);
  if (dbh != NULL)
    dscuss_db_close (dbh);
  dscuss_free_non_null (db_filename, g_free);
  ctx->callback (result, ctx->user_data);
  register_context_free (ctx);
}


gboolean
dscuss_register (const gchar* nickname,
                 const gchar* info,
                 DscussRegisterCallback callback,
                 gpointer user_data)
{
  gchar* user_directory = NULL;
  gchar* privkey_filename = NULL;
  gchar* pow_filename = NULL;
  gchar* db_filename = NULL;
  DscussPrivateKey* privkey = NULL;
  const DscussPublicKey* pubkey = NULL;
  RegisterContext* ctx = NULL;
  gboolean result = FALSE;

  g_assert (nickname);
  /* TBD: validate nickname */


  user_directory = g_build_filename (dscuss_util_get_data_dir (),
                                     nickname, NULL);
  if (g_mkdir_with_parents(user_directory, 0755) == -1)
    {
      g_critical ("Failed to create directory '%s'.", user_directory);
      goto out;
    }

  privkey_filename = g_build_filename (user_directory, "privkey.pem", NULL);
  privkey = dscuss_crypto_private_key_init (privkey_filename);
  if (privkey == NULL)
    {
      g_critical ("Failed to initialize private key.");
      goto out;
    }

  db_filename = g_build_filename (user_directory, "db", NULL);
  if (g_file_test (db_filename, G_FILE_TEST_EXISTS))
    {
      g_critical ("Database file '%s' already exists."
                  "Looks like the user is already registered.",
                  db_filename);
      goto out;
    }

  pubkey = dscuss_crypto_private_key_get_public (privkey);
  pow_filename = g_build_filename (user_directory, "proof_of_work.tmp", NULL);
  ctx = register_context_new (nickname,
                              info,
                              db_filename,
                              privkey,
                              callback,
                              user_data);
  privkey = NULL;
  if (!dscuss_crypto_pow_find (pubkey,
                               pow_filename,
                               on_pow_search_finished,
                               ctx))
    {
      g_critical ("Failed to start finding PoW.");
      register_context_free (ctx);
      goto out;
    }

  result = TRUE;

out:
  dscuss_free_non_null (user_directory, g_free);
  dscuss_free_non_null (privkey_filename, g_free);
  dscuss_free_non_null (pow_filename, g_free);
  dscuss_free_non_null (db_filename, g_free);
  if (privkey != NULL)
    dscuss_crypto_private_key_free (privkey);
  return result;
}


gboolean
dscuss_login (const gchar* nickname,
              DscussNewMessageCallback msg_callback_,
              gpointer msg_data_,
              DscussNewUserCallback user_callback_,
              gpointer user_data_,
              DscussNewOperationCallback oper_callback_,
              gpointer oper_data_)
{
  gchar* privkey_filename = NULL;
  gchar* db_filename = NULL;
  gchar* subs_filename = NULL;
  gchar* addr_filename = NULL;
  const DscussPublicKey* pubkey = NULL;
  gchar* pubkey_digest = NULL;
  gsize pubkey_digest_len = 0;
  DscussHash id;
  gboolean result = FALSE;

  g_assert (nickname != NULL);

  if (self != NULL)
    {
      g_warning ("You are already logged in as '%s'.",
                 dscuss_user_get_nickname (self->user));
      return FALSE;
    }

  self = g_new0 (struct LoggedUser, 1);

  privkey_filename = g_build_filename (dscuss_util_get_data_dir (),
                                       nickname, "privkey.pem", NULL);
  self->privkey = dscuss_crypto_private_key_init (privkey_filename);
  if (self->privkey == NULL)
    {
      g_critical ("Failed to initialize private key from '%s'.",
                  privkey_filename);
      goto out;
    }

  pubkey = dscuss_crypto_private_key_get_public (self->privkey);
  if (!dscuss_crypto_public_key_to_der (pubkey,
                                        &pubkey_digest,
                                        &pubkey_digest_len))
    {
      g_critical ("Failed to serialize public key.");
      goto out;
    }
  dscuss_crypto_hash_sha512 (pubkey_digest,
                             pubkey_digest_len,
                             &id);

  db_filename = g_build_filename (dscuss_util_get_data_dir (),
                                  nickname, "db", NULL);
  if (!(self->dbh = dscuss_db_open (db_filename)))
    {
      g_critical ("Failed to open database connection with '%s'.", db_filename);
      goto out;
    }
  if (!(self->user = dscuss_db_get_user (self->dbh, &id)))
    {
      g_critical ("Failed to fetch the user with id '%s' from the database.",
                  dscuss_crypto_hash_to_string (&id));
      goto out;
    }
  subs_filename = g_build_filename (dscuss_util_get_data_dir (),
                                    nickname, "subscriptions", NULL);
  if (!(self->subscriptions = dscuss_subscriptions_read (subs_filename)))
    {
      g_critical ("Error initializing the user's subscriptions.");
      goto out;
    }

  self->peers         = NULL;
  self->msg_callback  = msg_callback_;
  self->msg_data      = msg_data_;
  self->user_callback = user_callback_;
  self->user_data     = user_data_;
  self->oper_callback = oper_callback_;
  self->oper_data     = oper_data_;

  addr_filename = g_build_filename (dscuss_util_get_data_dir (),
                                    nickname, "addresses", NULL);
  if (!dscuss_network_init (addr_filename,
                            peer_connected_cb, NULL))
    {
      g_critical ("Error initializing the network subsystem!");
      goto out;
    }

  result = TRUE;

out:
  dscuss_free_non_null (privkey_filename, g_free);
  dscuss_free_non_null (db_filename, g_free);
  dscuss_free_non_null (subs_filename, g_free);
  dscuss_free_non_null (addr_filename, g_free);
  dscuss_free_non_null (pubkey_digest, g_free);
  if (!result)
    dscuss_logout ();
  return result;
}


void
dscuss_logout (void)
{
  g_debug ("Logging out...");

  if (self == NULL)
    {
      g_warning ("Failed to log out: you are not logged in.");
      return;
   }

  if (self->peers != NULL)
    {
      g_debug ("Freeing peers...");
      g_slist_free_full (self->peers, (GDestroyNotify) dscuss_peer_free);
    }

  dscuss_network_uninit ();

  if (self->subscriptions != NULL)
    {
      g_debug ("Freeing the user's subscription.");
      dscuss_subscriptions_free (self->subscriptions);
    }

  if (self->dbh != NULL)
    {
      g_debug ("Closing database connection.");
      dscuss_db_close (self->dbh);
    }

  if (self->user != NULL)
    {
      g_debug ("Freeing the user entity.");
      dscuss_user_free (self->user);
    }

  if (self->privkey != NULL)
    {
      g_debug ("Freeing the user's private key.");
      dscuss_crypto_private_key_free (self->privkey);
    }

  g_free (self);
  self = NULL;
}


gboolean
dscuss_is_logged_in (void)
{
  return (self != NULL);
}


const gchar*
dscuss_get_data_dir ()
{
  return dscuss_util_get_data_dir ();
}


const GSList*
dscuss_get_peers (void)
{
  if (!dscuss_is_logged_in ())
    {
      g_warning ("Can't list peers: not logged in.");
      return NULL;
    }

  g_assert (self != NULL);
  return self->peers;
}


void
dscuss_iterate ()
{
  g_main_context_iteration (NULL, TRUE);
}


static void
on_send_finished (DscussPeer* peer,
                  const DscussEntity* entity,
                  gboolean result,
                  gpointer user_data)
{
  if (result)
    g_debug ("Entity '%s' has been successfully sent to '%s'",
             dscuss_entity_get_description (entity),
             dscuss_peer_get_description (peer));
  else
    g_debug ("Failed to send the entity '%s' to '%s'",
             dscuss_entity_get_description (entity),
             dscuss_peer_get_description (peer));
}


void
dscuss_send_message (DscussMessage* msg)
{
  GSList* iterator = NULL;

  g_assert (self != NULL);

  if (!dscuss_db_put_message (self->dbh, msg))
    {
      g_critical ("Failed to store message '%s' in the database!",
                  dscuss_message_get_description (msg));
    }

  for (iterator = self->peers; iterator; iterator = iterator->next)
    {
      DscussPeer* peer = iterator->data;
      if (is_message_relevant (dscuss_peer_get_subscriptions (peer), msg))
        {
          if (!dscuss_peer_send (peer,
                                 (DscussEntity*) msg,
                                 self->privkey,
                                 on_send_finished,
                                 NULL))
            g_warning ("Failed to queue message '%s' for delivery"
                       " to the peer '%s'",
                       dscuss_message_get_description (msg),
                       dscuss_peer_get_description (peer));
        }
    }
}


/*** IterateMessageContext ***************************************************/

typedef struct IterateMessageContext
{
  DscussIterateMessageCallback callback;
  gpointer user_data;
} IterateMessageContext;


static IterateMessageContext*
iterate_message_context_new (DscussIterateMessageCallback callback,
                             gpointer user_data)
{
  IterateMessageContext* ctx = g_new0 (IterateMessageContext, 1);
  ctx->callback = callback;
  ctx->user_data = user_data;
  return ctx;
}


static void
iterate_message_context_free (IterateMessageContext* ctx)
{
  g_assert (ctx != NULL);
  g_free (ctx);
}

/*** End of IterateMessageContext ********************************************/


static void
iterate_message_callback (gboolean success,
                          DscussMessage* msg,
                          gpointer user_data)
{
  IterateMessageContext* ctx = user_data;
  ctx->callback (success, msg, ctx->user_data);
}


void
dscuss_get_messages (DscussIterateMessageCallback callback,
                     gpointer user_data)
{
  IterateMessageContext* ctx = NULL;
  ctx = iterate_message_context_new (callback,
                                     user_data);
  dscuss_db_get_recent_messages (self->dbh,
                                 iterate_message_callback,
                                 ctx);
  iterate_message_context_free (ctx);
}


DscussMessage*
dscuss_get_message (const DscussHash* msg_id)
{
  return dscuss_db_get_message (self->dbh, msg_id);
}


/*** Internal API ************************************************************/

const DscussUser*
dscuss_get_logged_user ()
{
  if (!dscuss_is_logged_in ())
    return NULL;

  g_assert (self != NULL);
  return self->user;
}


const DscussPrivateKey*
dscuss_get_logged_user_private_key ()
{
  if (!dscuss_is_logged_in ())
    return NULL;

  g_assert (self != NULL);
  return self->privkey;
}

/*** End of Internal API *****************************************************/
