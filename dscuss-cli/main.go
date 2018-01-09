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
	"vminko.org/dscuss"
)

const (
	cliVersion string = "1.0"
)

var (
	argConfig  = flag.String("config", dscuss.DefaultCfgDir, "Directory with config files to use")
	argVersion = flag.Bool("version", false, "Display version of the program and exit")
	argHelp    = flag.Bool("help", false, "Print help message and exit")
)

var commandList = []*ishell.Cmd{
	{
		Name: "register",
		Help: "<nickname> [additional_info], register new user.",
		Func: doRegister,
	},
	{
		Name: "login",
		Help: "<nickname>, login as user <nickname>.",
		Func: doLogin,
	},
	{
		Name: "logout",
		Help: "logout from the network.",
		Func: doLogout,
	},
	{
		Name: "list peers",
		Help: "list connected peers.",
		Func: doListPeers,
	},
	{
		Name: "make thread",
		Help: "start a new thread.",
		Func: doMakeThread,
	},
	{
		Name: "make reply",
		Help: "<id>, publish a new reply to message <id>.",
		Func: doMakeReply,
	},
	{
		// TBD: add optional topic parameter
		Name: "list board",
		Help: "list threads on the board.",
		Func: doListBoard,
	},
	{
		Name: "subscribe",
		Help: "<topic>. subscribe to <topic>.",
		Func: doSubscribe,
	},
	{
		Name: "unsubscribe",
		Help: "<topic>, unsubscribe from <topic>.",
		Func: doUnsubscribe,
	},
	{
		Name: "list subscriptions",
		Help: "list the current user's subscriptions.",
		Func: doListSubscriptions,
	},
	{
		Name: "version",
		Help: "display versions of " + dscuss.Name + " and the CLI.",
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

func doLogin(c *ishell.Context)             { c.Println("Not implemented yet.") }
func doLogout(c *ishell.Context)            { c.Println("Not implemented yet.") }
func doListPeers(c *ishell.Context)         { c.Println("Not implemented yet.") }
func doMakeThread(c *ishell.Context)        { c.Println("Not implemented yet.") }
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
	dscuss.Logf(dscuss.DEBUG, "Using CLI version %s.", cliVersion)

	runShell()

	dscuss.Logf(dscuss.DEBUG, "Stopping Dscuss...")
	dscuss.Uninit()
	fmt.Println("Dscuss stopped.")
}
