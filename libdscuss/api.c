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

#include "config.h"
#include "network.h"
#include "crypto.h"
#include "util.h"
#include "api.h"


/* List of connected peers. */
static GSList *peers = NULL;

/* Callbacks for notification of the UI */
static DscussInitCallback init_callback;
static gpointer init_data;
static DscussNewMessageCallback msg_callback;
static gpointer msg_data;
static DscussNewUserCallback user_callback;
static gpointer user_data;
static DscussNewOperationCallback oper_callback;
static gpointer oper_data;


static void
on_new_entity (DscussPeer* peer,
               DscussEntity* entity,
               gboolean result,
               gpointer user_data)
{
  if (!result)
    {
      g_warning ("Failed to read from peer '%s'",
                 dscuss_peer_get_description (peer));
      peers = g_slist_remove (peers, peer);
      dscuss_peer_free (peer);
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
      msg_callback ((DscussMessage*) entity, msg_data);
      break;;

    case DSCUSS_ENTITY_TYPE_OPER:
      /* TBD */
      g_assert_not_reached ();
      break;

    default:
      g_assert_not_reached ();
    }
}


static void
peer_connected_cb (DscussPeer* peer,
                   gpointer user_data)
{
  g_debug ("Connection with a new peer is established '%s'.",
           dscuss_peer_get_connecton_description (peer));
  peers = g_slist_append (peers, peer);
  dscuss_peer_set_receive_callback (peer,
                                    on_new_entity,
                                    NULL);
}

static void
on_crypto_init_finished (gboolean result,
                         gpointer user_data)
{
  g_debug ("Crypto initialization finished with the result %d",
           result);

  if (!result)
    goto error;

  if (!dscuss_network_init (peer_connected_cb , NULL))
    {
      g_error ("Error initializing the network subsystem!");
      goto error;
    }

  init_callback (result, init_data);
  return;

error:
  dscuss_uninit ();
}


gboolean
dscuss_init (const gchar* data_dir,
             DscussInitCallback init_callback_,
             gpointer init_data_,
             DscussNewMessageCallback msg_callback_,
             gpointer msg_data_,
             DscussNewUserCallback user_callback_,
             gpointer user_data_,
             DscussNewOperationCallback oper_callback_,
             gpointer oper_data_)
{
  init_callback = init_callback_;
  init_data = init_data_;
  msg_callback = msg_callback_;
  msg_data = msg_data_;
  user_callback = user_callback_;
  user_data = user_data_;
  oper_callback = oper_callback_;
  oper_data = oper_data_;

  dscuss_util_init (data_dir);

  if (!dscuss_config_init ())
    {
      g_error ("Error initializing the configuration subsystem!");
      goto error;
    }

  if (!dscuss_crypto_init (on_crypto_init_finished, NULL))
    {
      g_error ("Error initializing the crypto subsystem!");
      goto error;
    }

  /* TBD: establish database connection */

  return TRUE;

error:
  dscuss_uninit ();
  return FALSE;
}


void
dscuss_uninit ()
{
  g_debug ("Uninitializing Dscuss");

  if (peers != NULL)
    {
      g_slist_free_full (peers, (GDestroyNotify) dscuss_peer_free);
      peers = NULL;
    }

  dscuss_crypto_uninit ();
  dscuss_network_uninit ();
  dscuss_config_uninit ();
  dscuss_util_uninit ();
  while (g_main_context_pending (NULL))
    g_main_context_iteration (NULL, TRUE);

  init_callback = NULL;
  init_data = NULL;
  msg_callback = NULL;
  msg_data = NULL;
  user_callback = NULL;
  user_data = NULL;
  oper_callback = NULL;
  oper_data = NULL;
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

  for (iterator = peers; iterator; iterator = iterator->next)
    {
      /* TBD: logic deciding whether we need to send this message to this peer
       * gboolean
       * dscuss_peer_is_message_relevant (const DscussPeer* peer,
       *                                  const DscussMessage* msg);
       **/
      DscussPeer* peer = iterator->data;
      dscuss_peer_send (peer,
                        (DscussEntity*) msg,
                        on_send_finished,
                        NULL);
    }
}


