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
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"strings"
	"vminko.org/dscuss"
	"vminko.org/dscuss/log"
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
		// TBD: add optional topic parameter
		Name: "lsboard",
		Help: "list threads on the board",
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
	if username == "" { // TBD: validate nickname via regexp
		c.Println("You must specify a nickname.")
		return
	}

	c.Print("Enter some additional info: ")
	info := c.ReadLine()

	c.Println("Registering new user...")
	err := dscuss.Register(username, info)
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
	/* TBD: validate nickname */

	err := dscuss.Login(nickname)
	if err != nil {
		c.Printf("Failed to log in as %s: %v\n", nickname, err)
	}
}

func doLogout(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
	} else {
		c.Println("Logging out...")
		dscuss.Logout()
	}
}

func doListPeers(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
		return
	}
	peers := dscuss.ListPeers()
	if len(peers) > 0 {
		if len(peers) > 1 {
			c.Printf("There are %d connected peers:\n", len(peers))
		} else {
			c.Println("There is one connected peer:")
		}
		for _, p := range peers {
			c.Printf("%s-%s (%s) is %s\n", p.Nickname, p.ID, p.RemoteAddr, p.StateName)
		}
	} else {
		c.Printf("There are no peers connected\n")
	}

}

func multiLineStopper(s string) bool {
	var terminator string = "DSC"
	if strings.HasSuffix(s, terminator) {
		s = strings.TrimSuffix(s, terminator)
		return false
	}
	return true
}

func doMakeThread(c *ishell.Context) {
	if !dscuss.IsLoggedIn() {
		c.Println("You are not logged in.")
		return
	}
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Print("Enter thread subject: ")
	subj := c.ReadLine()

	var term string = "DSC"
	c.Printf("Enter message text and end with '%s': ", term)
	text := c.ReadMultiLines(term)
	text = strings.TrimSuffix(text, term)
	text = strings.TrimRight(text, "\r\n")
	c.Println(subj, text)

	t := dscuss.NewThread(subj, text)
	dscuss.SendMessage(t)
}

func doMakeReply(c *ishell.Context)         { c.Println("Not implemented yet.") }
func doListBoard(c *ishell.Context)         { c.Println("Not implemented yet.") }
func doSubscribe(c *ishell.Context)         { c.Println("Not implemented yet.") }
func doUnsubscribe(c *ishell.Context)       { c.Println("Not implemented yet.") }
func doListSubscriptions(c *ishell.Context) { c.Println("Not implemented yet.") }

func doVersion(c *ishell.Context) {
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
