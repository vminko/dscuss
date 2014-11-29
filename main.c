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
#include <glib/gstdio.h>
#include "libdscuss/dscuss.h"

#define PROG_NAME        "Dscuss"
#define PROG_VERSION     "proof-of-concept"
#define DEFAULT_DATA_DIR ".dscuss"
#define DEFAULT_LOGFILE_NAME "dscuss.log"

static FILE* log_file = 0;
static gchar* log_file_name = NULL;


/* TBD: move logging to a separate file. */

static void
log_handler (const gchar* log_domain,
             GLogLevelFlags log_level,
             const gchar* message,
             gpointer user_data)
{
  GDateTime* datetime = NULL;
  gchar* log_level_str = NULL;
  gchar* datetime_str = NULL;

  g_assert (log_file != NULL);

  datetime = g_date_time_new_now_local ();
  datetime_str = g_date_time_format (datetime, "%F %T");
  g_date_time_unref (datetime);

  switch (log_level)
    {
    case G_LOG_LEVEL_DEBUG:
      log_level_str = "DEBUG";
      break;

    case G_LOG_LEVEL_INFO:
      log_level_str = "INFO";
      break;

    case G_LOG_LEVEL_MESSAGE:
      log_level_str = "INFO";
      break;

    case G_LOG_LEVEL_WARNING:
      log_level_str = "WARNING";
      break;

    case G_LOG_LEVEL_CRITICAL:
      log_level_str = "CRITICAL";
      break;

    case G_LOG_LEVEL_ERROR:
      log_level_str = "ERROR";
      break;

    default:
      log_level_str = "UNKNOWN";
      break;
    }

  g_fprintf (log_file, "<%s> %s: %s\n", datetime_str, log_level_str, message);
  g_free (datetime_str);
}


static gboolean
logger_init (const gchar* log_file_name)
{
  log_file = g_fopen (log_file_name, "a");
  if (log_file == NULL)
    return FALSE;

  g_log_set_handler (NULL,
                     G_LOG_LEVEL_ERROR | G_LOG_LEVEL_CRITICAL
                       | G_LOG_LEVEL_WARNING | G_LOG_LEVEL_MESSAGE | G_LOG_LEVEL_DEBUG,
                     log_handler,
                     NULL);

  return TRUE;
}


static void
logger_uninit (void)
{
  if (log_file != NULL)
    fclose (log_file);
}


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
  g_option_context_set_summary (opt_context,
                                PROG_NAME " - decentralized forum.");
  g_option_context_set_description (opt_context,
                                    "Please report bugs to <vitaly.minko@gmail.com>.");
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

  if (!opt_config_dir_arg)
    opt_config_dir_arg = g_build_filename (g_get_home_dir (),
                                           DEFAULT_DATA_DIR,
                                           NULL);
  if (g_mkdir_with_parents (opt_config_dir_arg, 0700) != 0)
    {
      g_printerr ("Failed to create data directory '%s'.\n",
                  opt_config_dir_arg);
      return 1;
    }
  log_file_name = g_build_filename (opt_config_dir_arg,
                                    DEFAULT_LOGFILE_NAME, NULL);
  if (!logger_init (log_file_name))
    {
      g_printerr ("Failed to initialize to logging subsystem.\n");
      goto uninit;
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

uninit:
  dscuss_uninit ();
  logger_uninit ();
  if (log_file_name != NULL)
    {
      g_free (log_file_name);
      log_file_name = NULL;
    }
  if (opt_config_dir_arg != NULL)
    {
      g_free (opt_config_dir_arg);
      opt_config_dir_arg = NULL;
    }

  return 0;
}
