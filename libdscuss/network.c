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

#include <glib.h>
#include <gio/gio.h>
#include "config.h"
#include "connection.h"
#include "util.h"
#include "network.h"

#define DSCUSS_NETWORK_IP_PORT_REGEX   \
  "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):\\d+$"
#define DSCUSS_NETWORK_HOST_PORT_REGEX \
  "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9]):\\d+$"
#define DSCUSS_NETWORK_ADDR_FILE_NAME "addresses"
#define DSCUSS_NETWORK_DEFAULT_PORT 8004
/* How often should we try to establish outgoing connections with other peers?
 * (in seconds) */
#define DSCUSS_NETWORK_DEFAULT_CLIENT_CONNECT_TIMEOUT 1


/* List of known peer addresses. */
static GSList* peer_addresses = NULL;

/* Client for establishing outgoing connections. */
static GSocketClient* client = NULL;

/* Service for establishing incoming connections. */
static GSocketService* service = NULL;

/* Hash of connected peersin the following format
 * [peer -> associated_address]
 * where associated_address is a remote host address from the peer_addresses,
 * which may be NULL in case of incoming connection. */
static GHashTable* peers = NULL;

/* ID of timeout for establishing outgoing connections. */
static guint timeout_id = 0;

/* Function to call when a new peer connects. */
static DscussNewPeerCallback new_peer_callback;

/* User data to pass to new_peer_callback. */
static gpointer new_peer_data;

/* Handler of the incoming connection signal. */
guint incoming_handler;


static void
on_disconnect (DscussPeer* peer,
               DscussPeerDisconnectReason reason,
               gpointer reason_data,
               gpointer user_data)
{
  gchar* address = NULL;
  const gchar* dup_address = NULL;
  DscussPeer* duplicate_peer = NULL;

  g_debug ("Peer '%s' disconnected with reason %d",
           dscuss_peer_get_description (peer), reason);

  switch (reason)
    {
    case DSCUSS_PEER_DISCONNECT_REASON_DUPLICATE:
      duplicate_peer = reason_data;
      gboolean is_found = g_hash_table_lookup_extended (peers,
                                                        peer,
                                                        NULL,
                                                        (gpointer)(&address));
      if (!is_found)
        {
          g_warning ("Peer '%s' was not found in the hash of connected peers",
                     dscuss_peer_get_description (peer));
          break;
        }

      if (address != NULL)
        {
          is_found = g_hash_table_lookup_extended (peers,
                                                   duplicate_peer,
                                                   NULL,
                                                   (gpointer)(&dup_address));
          if (!is_found)
            {
              g_warning ("Duplicated peer connection '%s' was not found"
                         " in the hash of connected peers'",
                         dscuss_peer_get_description (peer));
              break;
            }
          if (dup_address != NULL)
            {
              g_warning ("Addresses '%s' and '%s' are addresses of the same peer",
                         address, dup_address);
            }
          else
            {
              g_hash_table_replace (peers,
                                    duplicate_peer,
                                    address);
            }
        }
      break;

    case DSCUSS_PEER_DISCONNECT_REASON_BROKEN:
    case DSCUSS_PEER_DISCONNECT_REASON_CLOSED:
      /* Nothing to do. */
      break;

    case DSCUSS_PEER_DISCONNECT_REASON_NO_COMMON_INTERESTS:
      /**
       * TBD: remove from peer_addresses
       */
      break;

    case DSCUSS_PEER_DISCONNECT_REASON_BANNED:
      /**
       * TBD: add remote address to ban list
       */
      break;

    default:
      g_assert_not_reached ();
      break;
    }

  if (!g_hash_table_remove (peers, peer))
    {
      g_warning ("Could not remove peer '%s' from the hash of connected peers",
                 dscuss_peer_get_description (peer));
    }
}


void
on_peer_handshaked (DscussPeer* peer,
                    gpointer user_data)
{
  new_peer_callback (peer, new_peer_data);
}


static gboolean
on_incoming_connection (GSocketService*    service,
                        GSocketConnection* socket_connection,
                        GObject*           source_object,
                        gpointer           user_data)
{
  /* TBD: return TRUE if addr is banned */

  g_object_ref (socket_connection);
  DscussPeer* peer = dscuss_peer_new (socket_connection,
                                      on_disconnect,
                                      NULL,
                                      on_peer_handshaked,
                                      NULL);
  g_debug ("New connection from '%s'",
           dscuss_peer_get_connecton_description (peer));
  g_hash_table_insert (peers, peer, NULL);

  /**
   * FIXME: temporary solution, handshake is not implemented yet
   */
  on_peer_handshaked (peer, NULL);

  return FALSE;
}


static gboolean
dscuss_network_start_listening (guint16 port)
{
  GError* error = NULL;

  service = g_socket_service_new ();
  if (!g_socket_listener_add_inet_port (G_SOCKET_LISTENER (service),
					port,
					NULL,
					&error))
    {
      g_warning ("%s", error->message);
      g_error_free (error);
      return FALSE;
    }
  incoming_handler = g_signal_connect (service,
                                       "incoming",
                                       G_CALLBACK (on_incoming_connection),
                                       NULL);
  g_socket_service_start (service);
  g_debug ("Started listening on port %d", port);

  return TRUE;
}


static gboolean
dscuss_network_validate_address (const gchar* addr)
{
  return g_regex_match_simple (DSCUSS_NETWORK_IP_PORT_REGEX, addr, 0, 0) ||
         g_regex_match_simple (DSCUSS_NETWORK_HOST_PORT_REGEX, addr, 0, 0);
}


static gboolean
dscuss_network_read_addresses (const gchar* addr_file_name)
{
  GError* error = NULL;
  GFile* file;
  GFileInputStream* file_in = NULL;
  GDataInputStream* data_in = NULL;
  gchar* line;
  gchar* path;
  gsize length = -1;
  gboolean res = TRUE;

  path = g_build_filename (dscuss_util_get_data_dir (), addr_file_name, NULL);
  file = g_file_new_for_path (path);
  g_free (path);

  file_in = g_file_read (file, NULL, &error);
  if (file_in == NULL)
    {
      g_warning ("%s", error->message);
      g_error_free (error);
      g_object_unref (file);
      return FALSE;
    }
  
  data_in = g_data_input_stream_new ((GInputStream*)file_in);
  error = NULL;
  while (TRUE)
    {
      line = g_data_input_stream_read_line_utf8 (G_DATA_INPUT_STREAM (data_in),
                                                 &length,
                                                 NULL,
                                                 &error);
      g_assert_no_error (error);
      if (error != NULL)
        {
          g_warning ("%s", error->message);
          g_error_free (error);
          res = FALSE;
          break;
        }

      if (line == NULL)
	break;

      if (dscuss_network_validate_address (line))
        {
          if (g_slist_find (peer_addresses, line) != NULL)
            {
              g_warning ("Duplicated peer address: '%s'!", line);
            }
          else
            {
              peer_addresses = g_slist_append (peer_addresses, line);
            }
        }
      else
        {
          g_warning ("'%s' is not a valid peer address, ignoring it.", line);
          g_free (line);
        }
    }

  g_object_unref (data_in);
  g_object_unref (file_in);
  g_object_unref (file);
  return res;
}


static gboolean
dscuss_network_establish_outgoing_connections (gpointer user_data)
{
  GHashTableIter iter;
  gpointer key, value;
  GSList* iterator = NULL;
  gboolean is_found = FALSE;

  for (iterator = peer_addresses; iterator; iterator = iterator->next)
    {
      gchar* address = iterator->data;
      g_debug ("Looking for associated socket connection for '%s'", address);
      is_found = FALSE;
      g_hash_table_iter_init (&iter, peers);
      while (g_hash_table_iter_next (&iter, &key, &value))
      {
        if (value == address)
          {
            DscussPeer* peer = key;
            g_debug ("Address '%s' is already associated with %s",
                     address,
                     dscuss_peer_get_connecton_description (peer));
            is_found = TRUE;
            break;
          }
      }
      if (!is_found)
        {
          g_debug ("Trying to connect to '%s'", address);
          GError* error = NULL;
          GSocketConnection* socket_connection;
          socket_connection = g_socket_client_connect_to_host (client,
                                                               address,
                                                               DSCUSS_NETWORK_DEFAULT_PORT,
                                                               NULL, &error);
          if (socket_connection != NULL)
            {
              g_debug ("Successfully connected to '%s'", address);
              DscussPeer* peer = dscuss_peer_new (socket_connection,
                                                  on_disconnect,
                                                  NULL,
                                                  on_peer_handshaked,
                                                  NULL);
              g_hash_table_insert (peers, peer, address);
              /**
               * FIXME: temporarily solution, handshake is not implemented yet
               */
              on_peer_handshaked (peer, NULL);
            }
          else
            {
              g_debug ("Could not connect to '%s': %s", address, error->message);
              g_error_free (error);
            }
        }
    }

  return TRUE;
}


static gboolean
dscuss_network_start_connecting_to_hosts (void)
{
  gint connect_timeout = dscuss_config_get_integer ("network",
                                                    "connect_timeout",
                                                    DSCUSS_NETWORK_DEFAULT_CLIENT_CONNECT_TIMEOUT);
  if (connect_timeout <= 0)
    {
      g_error ("Invalid value of the 'connect_timeout' parameter from the 'network' group: %d",
               connect_timeout);
      return FALSE;
    }
  client = g_socket_client_new ();
  timeout_id = g_timeout_add_seconds (connect_timeout,
                                      dscuss_network_establish_outgoing_connections,
                                      NULL);
  return TRUE;
}


gboolean
dscuss_network_init (DscussNewPeerCallback new_peer_callback_,
                     gpointer new_peer_data_)
{
  peers = g_hash_table_new_full (g_direct_hash, g_direct_equal,
                                 NULL, NULL);

  gint port = dscuss_config_get_integer ("network", "port",
                                         DSCUSS_NETWORK_DEFAULT_PORT);
  if (port <= 0 || port > 65535)
    {
      g_error ("Invalid value of the 'port' parameter from the 'network' group: %d",
               port);
      goto error;
    }
  if (!dscuss_network_start_listening (port))
    {
      g_error ("Could not start listening incoming connections on port %d",
               DSCUSS_NETWORK_DEFAULT_PORT);
      goto error;
    }
  if (!dscuss_network_read_addresses (DSCUSS_NETWORK_ADDR_FILE_NAME))
    {
      g_error ("Could not read host addresses from '%s'",
               DSCUSS_NETWORK_ADDR_FILE_NAME);
      goto error;
    }
  if (!dscuss_network_start_connecting_to_hosts ())
    {
      g_error ("Could not start connecting to hosts");
      goto error;
    }

  new_peer_callback = new_peer_callback_;
  new_peer_data = new_peer_data_;
  return TRUE;

error:
  dscuss_network_uninit ();
  return FALSE;
}


void
dscuss_network_uninit (void)
{
  g_source_remove (timeout_id);

  if (peers != NULL)
    g_hash_table_destroy (peers);

  if (client != NULL)
    g_object_unref (client);

  if (service != NULL)
    {
      if (incoming_handler != 0)
        g_signal_handler_disconnect (service, incoming_handler);
      g_socket_service_stop (service);
      g_object_unref (service);
    }

  if (peer_addresses != NULL)
    g_slist_free_full (peer_addresses, g_free);

}
