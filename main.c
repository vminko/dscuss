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
#include <glib.h>
#include <glib-unix.h>
#include <gio/gio.h>
#include <glib/gprintf.h>
#include <glib/gstdio.h>
#include "libdscuss/include/dscuss.h"

#define PROG_NAME                "Dscuss"
#define PROG_VERSION             "proof-of-concept"
#define DEFAULT_DATA_DIR         ".dscuss"
#define DEFAULT_LOGFILE_NAME     "dscuss.log"
#define DEFAULT_TMPFILE_NAME     "dscuss.tmp"
#define EDITOR_CMD_MAX_LEN       1024


/**** Global variables *******************************************************/


/* Flag indicates that user wants to stop the program. */
gboolean stop_requested = FALSE;

static void
request_stop ()
{
  stop_requested = TRUE;
}

static gboolean
is_stop_requested ()
{
  return stop_requested;
}


/**** End of global variables ************************************************/


struct DscussCommand
{
  const gchar* command;
  gboolean (*action) (const gchar* arguments);
  const gchar* helptext;
};

static gboolean do_help (const gchar* args);
static void start_handling_input (void);



/**** Start of logging *******************************************************/

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
  fflush (log_file);
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

/**** End of logging *********************************************************/



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



/**** Command handlers *******************************************************/

static void
print_prompt (void)
{
  g_printf (">");
  fflush (stdout);
}


static void
on_register_finished (gboolean result,
                      gpointer user_data)
{
  if (result)
    g_printf ("New user successfully registered.\n");
  else
    g_printf ("Failed to register new user!\n");
  start_handling_input ();
}


static gboolean
do_register (const gchar* args)
{
  gchar* nickname = NULL;
  gchar* space = NULL;
  const gchar* info = NULL;
  gboolean enable_input = TRUE;

  if (args == NULL || strlen (args) == 0)
    {
      g_printf ("You must specify a nickname.\n");
      return TRUE;
    }

  nickname = strdup (args);
  space = strstr (nickname, " ");
  if (space != NULL)
    {
      *space = '\0';
      info = args + strlen (nickname) + 1;
    }

  /* TBD: validate input */

  if (dscuss_register (nickname,
                       info,
                       on_register_finished,
                       NULL))
    {
      g_printf ("Registering new user '%s', this will take about 4 hours...\n",
                nickname);
      enable_input = FALSE;
    }
  else
    {
      g_printf ("Failed to register new user '%s'. See log file for details.\n",
                 nickname);
    }
  g_free (nickname);

  return enable_input;
}


static gboolean
do_login (const gchar* nickname)
{
  if (dscuss_is_logged_in ())
    {
      g_printf ("You are already logged into the network."
                "You need to `logout' before logging in as another user.\n");
      return TRUE;
    }

  if (nickname == NULL || strlen (nickname) == 0)
    {
      g_printf ("You must to specify a nickname.");
      return TRUE;
    }

  /* TBD: validate nickname */
  if (!dscuss_login (nickname,
                     on_new_message, NULL,
                     on_new_user, NULL,
                     on_new_operation, NULL))
    {
      g_printf ("Failed to log in as '%s'\n", nickname);
    }

  return TRUE;
}


static gboolean
do_logout (const gchar* args)
{
  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged int.\n");
    }
  else
    {
      g_printf ("Logging out...\n");
      dscuss_logout ();
    }

  return TRUE;
}


static gboolean
do_list_peers (const gchar* args)
{
  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged in.\n");
    }
  else
    {
      const GSList* peers = dscuss_get_peers ();
      const GSList* iterator = NULL;
      for (iterator = peers; iterator; iterator = iterator->next)
        {
          DscussPeer* peer = iterator->data;
          g_printf ("%s\n", dscuss_peer_get_description (peer));
        }
    }

  return TRUE;
}


/**** EnteredMsg *************************************************************/

typedef struct
{
  DscussTopic* topic;
  gchar* subject;
  gchar* text;
} EnteredMsg;


static EnteredMsg*
entered_msg_new (DscussTopic* topic,
                 gchar* subject,
                 gchar* text)
{
  EnteredMsg* em = g_new0 (EnteredMsg, 1);
  em->topic = topic != NULL ? dscuss_topic_copy (topic) : NULL;
  em->subject = strdup (subject);
  em->text = strdup (text);
  return em;
}


static void
entered_msg_free (EnteredMsg* ctx)
{
  dscuss_topic_free (ctx->topic);
  g_free (ctx->subject);
  g_free (ctx->text);
  g_free (ctx);
}

/**** End of EnteredMsg ******************************************************/


EnteredMsg*
entered_msg_read_from_file (const gchar* tmp_file_name,
                            gboolean read_topic)
{
  GError* error = NULL;
  GFile* file;
  GFileInputStream* file_in = NULL;
  GDataInputStream* data_in = NULL;
  gchar* line = NULL;
  DscussTopic* topic = NULL;
  gchar* subject = NULL;
  gchar* tmp = NULL;
  gchar* text = NULL;
  EnteredMsg* entered_msg = NULL;

  file = g_file_new_for_path (tmp_file_name);
  file_in = g_file_read (file, NULL, &error);
  if (error != NULL)
    {
      g_printf ("Failed to open temporary input file '%s': %s",
                tmp_file_name, error->message);
      g_error_free (error);
      g_object_unref (file);
      return FALSE;
    }

  data_in = g_data_input_stream_new ((GInputStream*) file_in);
  error = NULL;
  while (TRUE)
    {
      line = g_data_input_stream_read_line_utf8 (G_DATA_INPUT_STREAM (data_in),
                                                 NULL,
                                                 NULL,
                                                 &error);
      if (error != NULL)
        {
          g_printf ("Failed to read from temporary input file '%s': %s",
                    tmp_file_name, error->message);
          g_error_free (error);
          break;
        }

      if (line == NULL)
	break;

      if (read_topic && topic == NULL)
        {
          topic = dscuss_topic_new (line);
          g_free (line);
          if (topic == NULL)
            {
              g_printf ("Failed to parse topic.");
              break;
            }
          else
            continue;
        }

      if (subject == NULL)
        {
          subject = line;
          continue;
        }

      if (text == NULL)
        {
          text = line;
        }
      else
        {
          tmp = g_strjoin (" ", text, line, NULL);
          g_free (text);
          g_free (line);
          text = tmp;
        }
    }

  if (text != NULL)
    {
      entered_msg = entered_msg_new (topic, subject, text);
    }

  if (topic != NULL)
    dscuss_topic_free (topic);
  if (subject != NULL)
    g_free (subject);
  if (text != NULL)
    g_free (text);
  g_object_unref (data_in);
  g_object_unref (file_in);
  g_object_unref (file);
  return entered_msg;
}

/**** EnteredMsg *************************************************************/


/*** EnterMsgContext ***********************************************/

typedef struct EnterMsgContext
{
  gchar* tmp_file_name;
  DscussHash* parent_id;
} EnterMsgContext;


static EnterMsgContext*
enter_message_context_new (gchar* tmp_file_name,
                           const DscussHash* parent_id)
{
  EnterMsgContext* ctx = g_new0 (EnterMsgContext, 1);
  ctx->tmp_file_name = tmp_file_name;
  if (parent_id != NULL)
    {
      ctx->parent_id = g_new0 (DscussHash, 1);
      memcpy (ctx->parent_id,
              parent_id,
              sizeof (DscussHash));
    }
  else
    ctx->parent_id = NULL;
  return ctx;
}


static void
enter_message_context_free (EnterMsgContext* ctx)
{
  g_assert (ctx != NULL);
  g_free (ctx->parent_id);
  g_free (ctx);
}

/*** End of IterateMessageTagsContext ****************************************/


void
on_msg_entered (GPid pid, int status, gpointer user_data)
{
  EnterMsgContext* ctx = user_data;
  EnteredMsg* em = NULL;
  DscussMessage* msg = NULL;

  g_spawn_close_pid (pid);

  em = entered_msg_read_from_file (ctx->tmp_file_name,
                                   ctx->parent_id == NULL);
  if (em != NULL)
    {
      if (ctx->parent_id)
        {
          msg = dscuss_message_new_reply (ctx->parent_id,
                                          em->subject,
                                          em->text);
        }
      else
        {
          msg = dscuss_message_new_thread (em->topic,
                                           em->subject,
                                           em->text);
        }
      dscuss_send_message (msg);
      dscuss_entity_unref ((DscussEntity*) msg);
      entered_msg_free (em);
    }
  else
    {
      g_printf ("Failed to parse entered message.\n");
    }


  if (g_unlink (ctx->tmp_file_name) != 0)
    {
      g_printf ("Failed to remove temporary input file '%s'", ctx->tmp_file_name);
    }
  g_free (ctx->tmp_file_name);
  enter_message_context_free (ctx);
  start_handling_input ();
}


static gboolean
publish_message (const DscussHash* parent_id)
{
  gint          rc             = 0;
  gboolean      boolrc         = FALSE;
  gchar**       argv           = NULL;
  gint          argp           = 0;
  const gchar*  editor         = NULL;
  GError*       error          = NULL;
  GPid          pid;
  gboolean      enable_input   = TRUE;
  gchar*        tmp_file_name  = NULL;
  gchar         editor_cmd[EDITOR_CMD_MAX_LEN];
  GFile*        file           = NULL;
  const gchar*  message_draft  = NULL;
  const gchar*  thread_draft   = "test,devel,dscuss\n"
                                 "This is a test thread\n"
                                 "This thread starts a new test dscussion.";
  const gchar*  reply_draft    = "This is a test subject\n"
                                 "This is a test reply.";

  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged in.\n");
      return TRUE;
    }

  editor = g_getenv ("EDITOR");
  if (editor == NULL)
    {
      g_printf ("The environment variable `EDITOR' is not set.\n");
      goto out;
    }

  tmp_file_name = g_build_filename (dscuss_get_data_dir(),
                                    DEFAULT_TMPFILE_NAME, NULL);

  file = g_file_new_for_path (tmp_file_name);
  message_draft = (parent_id == NULL) ? thread_draft : reply_draft;
  boolrc = g_file_replace_contents (file,
                                    message_draft,
                                    strlen (message_draft),
                                    NULL,                     /* old etag */
                                    FALSE,                    /* make baclup */
                                    G_FILE_CREATE_NONE,
                                    NULL,                     /* old etag */
                                    NULL,                     /* cancellable */
                                    &error);
  g_object_unref (file);
  if (!boolrc)
    {
      g_printf ("Failed to write draft the temporary input file: %s.\n", error->message);
      g_error_free (error);
      goto out;
    }

  g_snprintf (editor_cmd,
              EDITOR_CMD_MAX_LEN,
              "%s %s",
              editor, tmp_file_name);

  rc = g_shell_parse_argv (editor_cmd, &argp, &argv, NULL);
  if (!rc)
    {
      g_printf ("Failed to parse the environment variable `EDITOR': %d.\n", rc);
      goto out;
    }

  rc = g_spawn_async (NULL, argv, NULL,
		      G_SPAWN_CHILD_INHERITS_STDIN | G_SPAWN_DO_NOT_REAP_CHILD,
		      NULL, NULL, &pid, &error);
  if (!rc)
    {
      g_printf ("Failed to start the `EDITOR': %s.\n", error->message);
      g_error_free (error);
      goto out;
    }

  enable_input = FALSE;
  EnterMsgContext* ctx = enter_message_context_new (tmp_file_name,
                                                    parent_id);
  g_child_watch_add (pid, (GChildWatchFunc)on_msg_entered, ctx);

out:
  if (argv != NULL)
    g_strfreev (argv);
  if (enable_input && tmp_file_name != NULL)
    g_free (tmp_file_name);
  return enable_input;
}


static gboolean
do_publish_thread (const gchar* args)
{
  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged in.\n");
      return TRUE;
    }

  return publish_message (NULL);
}


static gboolean
do_publish_reply (const gchar* msg_id_str)
{
  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged in.\n");
      return TRUE;
    }

  if (msg_id_str == NULL)
    {
      g_printf ("You must specify message ID.\n");
      return TRUE;
    }
  DscussHash* msg_id = dscuss_crypto_hash_from_string (msg_id_str);
  if (msg_id == NULL)
    {
      g_printf ("Malformed hash string.\n");
      return TRUE;
    }
  /* TBD: check that message <msg_id> is stored in the DB. */

  gboolean result = publish_message (msg_id);
  g_free (msg_id);

  return result;
}


void
list_board_callback (gboolean success,
                     GList*   board_listing,
                     gpointer user_data)
{
  if (!success)
    {
      g_printf ("Failed to fetch messages from the database\n");
      start_handling_input ();
      return;
    }

  const GList* iterator = NULL;
  for (iterator = board_listing; iterator; iterator = iterator->next)
    {
      DscussMessage* msg = iterator->data;
      gchar* topic_str = dscuss_topic_to_string (dscuss_message_get_topic (msg));
      g_printf ("Topic: %s\n"
                "Subject: %s\n"
                "ID: %s\n\n",
                topic_str,
                dscuss_message_get_subject (msg),
                dscuss_crypto_hash_to_string (dscuss_message_get_id (msg)));
      g_free (topic_str);
    }
  g_printf ("done\n");

  g_list_free_full (board_listing, (GDestroyNotify)dscuss_message_free);
  start_handling_input ();
}


static gboolean
do_list_board (const gchar* args)
{
  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged in.\n");
      return TRUE;
    }

  dscuss_list_board (list_board_callback, NULL);
  return FALSE;
}


/* TBD: move this to dscuss_thread_free? */
static gboolean
free_message_node_callback (GNode *node,
                            gpointer data)
{
  dscuss_message_free (node->data);
  return FALSE;
}


static void
thread_free (GNode* root)
{
  g_node_traverse (root,
                   G_POST_ORDER,
                   G_TRAVERSE_ALL,
                   -1,   /* max_depth */
                   free_message_node_callback,
                   NULL);
  g_node_destroy (root);
}


static void
print_indent (GNode *node)
{
  guint i = 0;
  for (i = 0; i < g_node_depth (node); i++)
    {
      g_printf ("    ");
    }
}


static gboolean
print_message_node_callback (GNode* node,
                             gpointer data)
{
  DscussMessage* msg = node->data;
  gchar* topic_str = NULL;


  if (dscuss_message_get_topic (msg) != NULL)
    topic_str = dscuss_topic_to_string (dscuss_message_get_topic (msg));

  gchar* datetime_str = g_date_time_format (dscuss_message_get_datetime (msg), "%F %T");
  print_indent (node);
  g_printf ("Dumping Message entity:\n");
  print_indent (node);
  g_printf ("  id:  %s\n", dscuss_crypto_hash_to_string (dscuss_message_get_id (msg)));
  if (topic_str != NULL)
    {
      print_indent (node);
      g_printf ("  topic:     '%s'\n", topic_str);
    }
  print_indent (node);
  g_printf ("  subject:   '%s'\n", dscuss_message_get_subject (msg));
  print_indent (node);
  g_printf ("  text:      '%s'\n", dscuss_message_get_content (msg));
  print_indent (node);
  g_printf ("  author_id:  %s\n", dscuss_crypto_hash_to_string (dscuss_message_get_author_id (msg)));
  print_indent (node);
  g_printf ("  parent_id:  %s\n", dscuss_crypto_hash_to_string (dscuss_message_get_parent_id (msg)));
  print_indent (node);
  g_printf ("  datetime:   %s\n", datetime_str);

  g_free (datetime_str);
  if (topic_str != NULL)
     g_free (topic_str);

  return FALSE;
}


static void
list_thread_callback (gboolean success,
                      GNode* message_tree,
                      gpointer user_data)
{
  if (!success)
    {
      g_printf ("Failed to list thread\n");
      start_handling_input ();
      return;
    }

  g_assert (message_tree != NULL);
  g_node_traverse (message_tree,
                   G_PRE_ORDER,
                   G_TRAVERSE_ALL,
                   -1,   /* max_depth */
                   print_message_node_callback,
                   NULL);
  thread_free (message_tree);
  start_handling_input ();
}


static gboolean
do_list_thread (const gchar* msg_id_str)
{
  if (!dscuss_is_logged_in ())
    {
      g_printf ("You are not logged in.\n");
      return TRUE;
    }

  if (msg_id_str == NULL)
    {
      g_printf ("You must specify message ID.\n");
      return TRUE;
    }
  DscussHash* msg_id = dscuss_crypto_hash_from_string (msg_id_str);
  if (msg_id == NULL)
    {
      g_printf ("Malformed hash string.\n");
      return TRUE;
    }

  dscuss_list_thread (msg_id,
                      list_thread_callback,
                      NULL);
  g_free (msg_id);
  return FALSE;
}


static gboolean
do_unknown (const gchar* args)
{
  g_printf ("Unknown command `%s'\n", args);
  return TRUE;
}


static gboolean
do_quit (const gchar* args)
{
  request_stop ();
  return FALSE;
}


/**
 * List of supported commands. The order matters!
 */
static struct DscussCommand commands[] = {
 /**
  * TBD
  {"subscribe", &do_subscribe,
   gettext_noop ("Use `subscribe category' to subscribe to a category")},
  {"list_subscriptions", &do_list_subscriptions,
   gettext_noop ("Use `list_subscriptions' to list all categories you are"
		 " subscribed to")},
  {"unsubscribe", &do_unsubscribe,
   gettext_noop ("Use `unsubscribe category' to unsubscribe from a category")},
  */
  {"register", &do_register,
   "Use `register <nickname> [additional_info]' to register new user."
    " with nickname <nickname> and optional additional info."},
  {"login",    &do_login,
   "Use `login <nickname>' to login as user <nickname>."},
  {"logout",   &do_logout,
   "Use `logout' to logout from the network."},
  {"lspeers",   &do_list_peers,
   "Use `peers to list connected peers."},
  {"thread",      &do_publish_thread,
   "Use `thread' to start a new thread"},
  {"reply",      &do_publish_reply,
   "Use `reply <id>' to publish a new reply to message <id>"},
  /* TBD: add optional topic parameter */
  {"lsboard",   &do_list_board,
   "Use `board' to list threads on the board."},
  {"lsthread",   &do_list_thread,
   "Use `thread <id>' to print all messages in the thread <id>."},
  {"quit",     &do_quit,
   "Use `quit' to terminate " PROG_NAME "."},
  {"help",     &do_help,
   "Use `help <command>' to get help for a specific command."},
  /* The following two commands must be last! */
  {"",         &do_unknown, NULL},
  {NULL, NULL, NULL},
};


static gboolean
do_help (const gchar* args)
{
  int i = 0;

  while ((NULL != args) &&
	 (0 != strlen (args)) && (commands[i].action != &do_help))
    {
      if (0 ==
	  g_ascii_strncasecmp (&args[1],
                               &commands[i].command[1],
                               strlen (args) - 1))
	{
	  g_printf ("%s\n", commands[i].helptext);
	  return TRUE;
	}
      i++;
    }

  i = 0;
  g_printf ("Available commands:");
  while (commands[i].action != &do_help)
    {
      g_printf (" %s", commands[i].command);
      i++;
    }
  g_printf ("\nMandatory arguments are enclosed in angle brackets."
            " Optional arguments are enclosed in square brackets.");
  g_printf ("\n%s\n", commands[i].helptext);
  return TRUE;
}

/**** End of command handlers ************************************************/


static gboolean
stdio_callback (GIOChannel* io, GIOCondition condition, gpointer data)
{
  gchar* line = NULL;
  gchar* args = NULL;
  GError* error = NULL;
  guint i = 0;
  gboolean res = FALSE;

  switch (g_io_channel_read_line (io, &line, NULL, NULL, &error))
    {
    case G_IO_STATUS_NORMAL:
      line[strlen (line) - 1] = '\0';
      i = 0;
      while ((NULL != commands[i].command) &&
             (0 != g_ascii_strncasecmp (commands[i].command,
                                        line,
                                        strlen (commands[i].command))))
        i++;

      /* Remove leading spaces from the argument list */
      args = line + strlen (commands[i].command);
      while (g_ascii_isspace (*args))
        args++;

      res = commands[i].action (args);
      g_free (line);
      if (res)
        {
          print_prompt ();
        }
      return res;

    case G_IO_STATUS_ERROR:
      g_printerr ("IO error: %s\n", error->message);
      g_error_free (error);
      return FALSE;

    case G_IO_STATUS_EOF:
      g_printerr ("No input data available");
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
start_handling_input (void)
{
  GIOChannel* stdio = NULL;
  stdio = g_io_channel_unix_new (fileno (stdin));
  g_io_add_watch (stdio, G_IO_IN, stdio_callback, NULL);
  g_io_channel_unref (stdio);
  print_prompt ();
}


static gboolean
on_stop_signal (gpointer data)
{
  /* Do NOT call any glib function from the signal handler.
   * Let glib finish its current iteration instead and
   * free all the resources after that. */
  request_stop ();
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

  g_unix_signal_add (SIGTERM, on_stop_signal, NULL);
  g_unix_signal_add (SIGINT,  on_stop_signal, NULL);
  g_unix_signal_add (SIGHUP,  on_stop_signal, NULL);

  if (!dscuss_init (opt_config_dir_arg))
    {
      g_printerr ("Failed to initialize the Dscuss system.\n");
      return 1;
    }

  start_handling_input ();

  while (!is_stop_requested ())
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
