/*
This file is part of Dscuss.
Copyright (C) 2017  Vitaly Minko

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <http://www.gnu.org/licenses/>.
*/

// Command line interface for Dscuss. It's not supposed to be final UI for end user.
// Used mostly for developing the library.
package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"strings"
	"time"
	"vminko.org/dscuss"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/p2p/peer"
)

const (
	cliVersion string = "1.0"
)

var (
	argConfig  = flag.String("config", dscuss.DefaultDir, "Directory with config files to use")
	argVersion = flag.Bool("version", false, "Display version of the program and exit")
	argHelp    = flag.Bool("help", false, "Print help message and exit")
)

var commandList = []*ishell.Cmd{
	{
		Name: "reg",
		Help: "<nickname> [additional_info], register new user",
		Func: doRegister,
	},
	{
		Name: "login",
		Help: "<nickname>, login as user <nickname>",
		Func: doLogin,
	},
	{
		Name: "logout",
		Help: "logout from the network",
		Func: doLogout,
	},
	{
		Name: "lspeers",
		Help: "list connected peers",
		Func: doListPeers,
	},
	{
		Name: "mkthread",
		Help: "start a new thread",
		Func: doMakeThread,
	},
	{
		Name: "mkreply",
		Help: "<id>, publish a new reply to message <id>",
		Func: doMakeReply,
	},
	{
		Name: "lsboard",
		Help: "[topic], list particular topic or all threads on the board",
		Func: doListBoard,
	},
	{
		Name: "sub",
		Help: "<topic>. subscribe to <topic>",
		Func: doSubscribe,
	},
	{
		Name: "unsub",
		Help: "<topic>, unsubscribe from <topic>",
		Func: doUnsubscribe,
	},
	{
		Name: "lssubs",
		Help: "list the current user's subscriptions",
		Func: doListSubscriptions,
	},
	{
		Name: "ver",
		Help: "display versions of " + dscuss.Name + " and the CLI",
		Func: doVersion,
	},
}

func doRegister(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Print("Nickname: ")
	username := c.ReadLine()
	if username == "" {
		c.Println("You must specify a nickname.")
		return
	}

	c.Print("Enter some additional info: ")
	info := c.ReadLine()

	prompt := "Enter list of topics you are interested in. " +
		"Each topic is a set of comma separated tags."
	subs := readMultiLines(c, prompt)
	if subs == "" {
		c.Println("Error: user subscriptions can not be nil.")
		return
	}

	c.Println("Registering new user. Do not interrupt the process.")
	c.Println("Otherwise you'll have to remove the user directory manually.")
	err := dscuss.Register(username, info, subs)
	if err != nil {
		c.Println("Could not register new user: " + err.Error() + ".")
	} else {
		c.Println("User registered successfully,")
	}
}

func doLogin(c *ishell.Context) {
	if dscuss.IsLoggedIn() {
		c.Println("You are already logged into the network." +
			" You need to 'logout' before logging in as another user.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.HelpText())
	}
	nickname := c.Args[0]
	err := dscuss.Login(nickname)
	if err != nil {
		c.Printf("Failed to log in as %s: %v\n", nickname, err)
	}
}

func doLogout(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.HelpText())
	}
	c.Println("Logging out...")
	dscuss.Logout()
}

func printPeerInfo(c *ishell.Context, i int, p *peer.Info, verbose bool) {
	if verbose {
		if i != 0 {
			c.Println("")
		}
		c.Printf("PEER #%d\n", i)
		c.Printf("Nickname:		%s\n", p.Nickname)
		c.Printf("ID:			%s\n", p.ID)
		c.Printf("LocalAddr:		%s\n", p.LocalAddr)
		c.Printf("RemoteAddr:		%s\n", p.RemoteAddr)
		c.Printf("AssociatedAddrs:	%s\n", strings.Join(p.AssociatedAddrs, ","))
		c.Print("Subscriptions:		")
		for j, t := range p.Subscriptions {
			if j == 0 {
				c.Printf("%s\n", t)
			} else {
				c.Printf("			%s\n", t)
			}
		}
		c.Printf("State:			%s\n", p.StateName)
	} else {
		c.Printf("%s-%s (%s) is %s\n", p.Nickname, p.ID, p.RemoteAddr, p.StateName)
	}
}

func doListPeers(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.HelpText())
	}
	peers := dscuss.ListPeers()
	if len(peers) > 0 {
		if len(peers) > 1 {
			c.Printf("There are %d connected peers:\n", len(peers))
		} else {
			c.Println("There is one connected peer:")
		}
		verb := len(c.Args) > 0 && c.Args[0] == "full"
		for i, p := range peers {
			printPeerInfo(c, i, p, verb)
		}
	} else {
		c.Printf("There are no peers connected\n")
	}

}

func readMultiLines(c *ishell.Context, prompt string) string {
	var term string = "DSC"
	c.Printf("%s and end with '%s': \n", prompt, term)
	text := c.ReadMultiLines(term)
	text = strings.TrimSuffix(text, term)
	text = strings.TrimRight(text, "\r\n")
	return text
}

func doMakeThread(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.HelpText())
	}
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Print("Enter thread topic: ")
	topic := c.ReadLine()

	c.Print("Enter thread subject: ")
	subj := c.ReadLine()
	if subj == "" {
		c.Println("Error: subject can not be empty.")
		return
	}

	text := readMultiLines(c, "Enter message text")
	if text == "" {
		c.Println("Error: message text can not be empty.")
		return
	}

	t, err := dscuss.NewThread(subj, text, topic)
	if err != nil {
		c.Println("Error making new thread: " + err.Error() + ".")
		return
	}
	err = dscuss.PostMessage(t)
	if err != nil {
		c.Println("Error posting new thread: " + err.Error() + ".")
	} else {
		c.Println("Thread '" + t.Desc() + "' created successfully.")
	}
}

func doListBoard(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) > 1 {
		c.Println(c.Cmd.HelpText())
	}
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	topic := ""
	if len(c.Args) == 1 {
		topic = c.Args[0]
	}

	const boardSize = 10
	var messages []*entity.Message
	var err error
	if topic != "" {
		messages, err = dscuss.ListTopic(topic, 0, boardSize)
	} else {
		messages, err = dscuss.ListBoard(0, boardSize)
	}
	if err != nil {
		c.Println("Can't list board: " + err.Error() + ".")
		return
	}

	for i, msg := range messages {
		if i != 0 {
			c.Println()
		}
		c.Printf("#%d by %s, %s\n", i, msg.AuthorID.Shorten(),
			msg.DateWritten.Format(time.RFC3339))
		if topic == "" {
			c.Printf("Topic: %s\n", msg.Topic.String())
		}
		c.Printf("Subject: %s\n", msg.Subject)
		c.Println(msg.Text)
	}
}

func doMakeReply(c *ishell.Context) { c.Println("Not implemented yet.") }
func doSubscribe(c *ishell.Context) {
	msg := `Not implemented yet.
To edit you subscriptions:
1. Logout;
2. Edit %s/<nickname>/subscriptions.txt using your favorite editor;
3/ Login.
`
	c.Printf(msg, dscuss.Dir())
}
func doUnsubscribe(c *ishell.Context) {
	msg := `Not implemented yet.
To edit you subscriptions:
1. Logout;
2. Edit %s/<nickname>/subscriptions.txt using your favorite editor;
3/ Login.
`
	c.Printf(msg, dscuss.Dir())
}

func doListSubscriptions(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.HelpText())
	}
	c.Print(dscuss.ListSubscriptions())
}

func doVersion(c *ishell.Context) {
	if len(c.Args) != 0 {
		c.Println(c.Cmd.HelpText())
	}
	c.Println(getVersion())
}

func runShell() {
	var shell = ishell.New()
	shell.Println("Welcome to Dscuss.")
	shell.SetPrompt("> ")
	for _, c := range commandList {
		shell.AddCmd(c)
	}
	defer shell.Close()
	shell.Run()
}

func getVersion() string {
	v := dscuss.FullVersion()
	v += fmt.Sprintf("CLI version: %s.", cliVersion)
	return v
}

func main() {

	sigChan := make(chan os.Signal)
	go func() {
		stacktrace := make([]byte, 8192)
		for _ = range sigChan {
			length := runtime.Stack(stacktrace, true)
			fmt.Println(string(stacktrace[:length]))
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)

	flag.Parse()

	if *argHelp {
		fmt.Println(dscuss.Name + " - P2P network for public discussion.")
		flag.Usage()
		return
	}

	if *argVersion {
		fmt.Println(getVersion())
		return
	}

	fmt.Println("Starting Dscuss...")
	err := dscuss.Init(*argConfig)
	if err != nil {
		panic("Could not initialize Dscuss: " + err.Error())
	}
	log.Debugf("Using CLI version %s.", cliVersion)

	runShell()

	log.Debug("Stopping Dscuss...")
	dscuss.Uninit()
	fmt.Println("Dscuss stopped.")
}
