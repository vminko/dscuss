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
#include <glib.h>
#include <glib-unix.h>
#include <glib/gprintf.h>
#include "libdscuss/dscuss.h"

#define PROG_NAME "Dscuss"
#define PROG_VERSION "proof-of-concept"


static gboolean
stdio_callback (GIOChannel * io, GIOCondition condition, gpointer data)
{
  gchar *line = NULL;
  DscussMessage* msg = NULL;
  GError *error = NULL;

  switch (g_io_channel_read_line (io, &line, NULL, NULL, &error))
    {
    case G_IO_STATUS_NORMAL:
      line[strlen (line) - 1] = '\0';
      g_debug ("Sending message '%s'", line);
      msg = dscuss_message_new (line);
      dscuss_send_message (msg);
      dscuss_entity_unref ((DscussEntity*) msg);
      g_free (line);
      return TRUE;

    case G_IO_STATUS_ERROR:
      g_printerr ("IO error: %s\n", error->message);
      g_error_free (error);

      return FALSE;

    case G_IO_STATUS_EOF:
      g_warning ("No input data available");
      return TRUE;

    case G_IO_STATUS_AGAIN:
      return TRUE;

    default:
      g_return_val_if_reached (FALSE);
      break;
    }

  return FALSE;
}


static void
start_handling_input ()
{
  GIOChannel *stdio = NULL;
  stdio = g_io_channel_unix_new (fileno (stdin));
  g_io_add_watch (stdio, G_IO_IN, stdio_callback, NULL);
  g_io_channel_unref (stdio);
}


static void
on_init_finished (gboolean result, gpointer user_data)
{
  gboolean* stop_flag = user_data;

  if (result)
    {
      g_printf ("Initialization is finished successfully.\n"
                "Your input will be handled from now on.\n");
      start_handling_input ();
    }
  else
    {
      g_printf ("Initialization failed. Quitting...\n");
      *stop_flag = TRUE;
    }
}


static void
on_new_message (DscussMessage* msg, gpointer user_data)
{
  g_printf ("New message received: '%s'.\n",
            dscuss_message_get_description (msg));
  dscuss_entity_unref ((DscussEntity*) msg);
}


static void
on_new_user (DscussUser* user, gpointer user_data)
{
  g_printf ("New user received.\n");
}


static void
on_new_operation (DscussOperation* oper, gpointer user_data)
{
  g_printf ("New operation received.\n");
}


static gboolean
on_stop (gpointer data)
{
  /* Do NOT call any glib function from the signal handler.
   * Let glib finish its current iteration instead and
   * free all the resources after that. */
  gboolean* stop_flag = data;
  *stop_flag = TRUE;
  return TRUE;
}


int
main (int argc, char* argv[])
{
  gboolean opt_version = FALSE;
  gchar* opt_config_dir_arg = NULL;
  GError* error = NULL;
  GOptionContext* opt_context;
  GOptionEntry opt_entries[] =
    {
      { "version", 'v', 0, G_OPTION_ARG_NONE,
        &opt_version, "Display version of the program and exit", NULL },
      { "config",  'c', 0, G_OPTION_ARG_STRING,
        &opt_config_dir_arg, "Directory with config files to use", NULL },
      { NULL}
    };
  gboolean stop_requested = FALSE;

  g_set_prgname (argv[0]);
  g_set_application_name (PROG_NAME);

  opt_context = g_option_context_new ("");
  g_option_context_set_summary (opt_context, PROG_NAME " - decentralized forum.");
  g_option_context_set_description (opt_context, "Please report bugs to <vitaly.minko@gmail.com>.");
  g_option_context_add_main_entries (opt_context, opt_entries, NULL);
  if (!g_option_context_parse (opt_context, &argc, &argv, &error))
    {
      g_printerr ("Error parsing command line options: %s\n", error->message);
      g_error_free (error);
      return 1;
    }
  g_option_context_free (opt_context);

  if (opt_version) {
    g_printf ("%s %s.\n", PROG_NAME, PROG_VERSION);
    return 0;
  }

  g_unix_signal_add (SIGTERM, on_stop, &stop_requested);
  g_unix_signal_add (SIGINT, on_stop, &stop_requested);
  g_unix_signal_add (SIGHUP, on_stop, &stop_requested);

  g_printf ("Initializing the system, this can take a while.\n");
  if (!dscuss_init (opt_config_dir_arg,
                    on_init_finished, &stop_requested,
                    on_new_message, NULL,
                    on_new_user, NULL,
                    on_new_operation, NULL))
    {
      g_printerr ("Failed to initialize the Dscuss system.\n");
      return 1;
    }

  while (!stop_requested)
      dscuss_iterate ();

  dscuss_uninit ();

  return 0;
}
