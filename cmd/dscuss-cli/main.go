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
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"vminko.org/dscuss"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/p2p/peer"
	"vminko.org/dscuss/subs"
	"vminko.org/dscuss/thread"
)

const (
	cliVersion string = "1.0"
)

var (
	argConfig  = flag.String("config", dscuss.DefaultDir, "Directory with config files to use")
	argVersion = flag.Bool("version", false, "Display version of the program and exit")
	argHelp    = flag.Bool("help", false, "Print help message and exit")
	// Looks like there is no way to pass LoginHandle via ishell.Context.
	loginHandle *dscuss.LoginHandle
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
		Help: "[topic], list a particular topic or all threads on the board",
		Func: doListBoard,
	},
	{
		Name: "lsthread",
		Help: "<id>, display a particular thread",
		Func: doListThread,
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
		Name: "mkmdr",
		Help: "<id>, make user <id> a moderator",
		Func: doMakeModerator,
	},
	{
		Name: "rmmdr",
		Help: "<id>, remove user <id> from the list of moderators",
		Func: doRemoveModerator,
	},
	{
		Name: "lsmdr",
		Help: "list the current user's moderators",
		Func: doListModerators,
	},
	{
		Name: "rmmsg",
		Help: "<id> <reason>, remove message <id> because of <reason>",
		Func: doRemoveMessage,
	},
	{
		Name: "ban",
		Help: "<id> <reason>, ban user <id> because of <reason>",
		Func: doBanUser,
	},
	{
		Name: "lsop",
		Help: "(user|msg) <id>, list operations on user or message <id>",
		Func: doListOperations,
	},
	{
		Name: "whoami",
		Help: "display nickname of the current user",
		Func: doWhoAmI,
	},
	{
		Name: "ver",
		Help: "display versions of " + dscuss.Name + " and the CLI",
		Func: doVersion,
	},
}

func userSummary(id *entity.ID) string {
	u, err := loginHandle.GetUser(id)
	var nick string
	switch {
	case err == errors.NoSuchEntity:
		nick = "[unknown user]"
	case err != nil:
		nick = "[error fetching user from db]"
	default:
		nick = u.Nickname
	}
	return fmt.Sprintf("%s (%s)", nick, id)
}

func doRegister(c *ishell.Context) {
	if loginHandle != nil {
		c.Println("You need to 'logout' before registering new user.")
		return
	}
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
	subsStr := readMultiLines(c, prompt)
	if subsStr == "" {
		c.Println("Error: user subscriptions can not be nil.")
		return
	}
	s, err := subs.ReadString(subsStr)
	if err != nil {
		c.Printf("Specified subscriptions are unacceptable: %v", err)
		return
	}

	c.Println("Registering new user. Do not interrupt the process.")
	c.Println("Otherwise you'll have to remove the user directory manually.")
	err = dscuss.Register(username, info, s)
	if err != nil {
		c.Println("Could not register new user: " + err.Error() + ".")
		return
	}

	c.Println("User registered successfully,")
	addrPath := filepath.Join(dscuss.Dir(), dscuss.AddressListFileName)
	c.Printf("Edit %s in you favorite editor if you want to customize peer addresses.\n",
		addrPath)
}

func doLogin(c *ishell.Context) {
	if loginHandle != nil {
		c.Println("You are already logged into the network." +
			" You need to 'logout' before logging in as another user.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	nickname := c.Args[0]
	var err error
	loginHandle, err = dscuss.Login(nickname)
	if err != nil {
		c.Printf("Failed to log in as %s: %v\n", nickname, err)
		return
	}
}

func doLogout(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.Help)
		return
	}
	c.Println("Logging out...")
	loginHandle.Logout()
	loginHandle = nil
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
		c.Printf("%s-%s (%s) is %s\n", p.Nickname, p.ShortID, p.RemoteAddr, p.StateName)
	}
}

func doListPeers(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) > 1 {
		c.Println(c.Cmd.Help)
		return
	}
	verb := false
	if len(c.Args) > 0 {
		if c.Args[0] == "full" {
			verb = true
		} else {
			c.Printf("Unknown parameter '%s'\n", c.Args[0])
			return
		}
	}
	peers := loginHandle.ListPeers()
	if len(peers) > 0 {
		if len(peers) > 1 {
			c.Printf("There are %d connected peers:\n", len(peers))
		} else {
			c.Println("There is one connected peer:")
		}
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
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.Help)
		return
	}

	c.Print("Enter thread topic: ")
	topicStr := c.ReadLine()
	topic, err := subs.NewTopic(topicStr)
	if err != nil {
		c.Println("Unacceptable topic: " + err.Error() + ".")
		return
	}

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

	t, err := loginHandle.NewThread(subj, text, topic)
	if err != nil {
		c.Println("Error making new thread: " + err.Error() + ".")
		return
	}
	err = loginHandle.PostEntity((entity.Entity)(t))
	if err != nil {
		c.Println("Error posting new thread: " + err.Error() + ".")
	} else {
		c.Println("Thread '" + t.String() + "' created successfully.")
	}
}

func doListBoard(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) > 1 {
		c.Println(c.Cmd.Help)
		return
	}
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	var err error
	var topic subs.Topic
	if len(c.Args) == 1 {
		topic, err = subs.NewTopic(c.Args[0])
		if err != nil {
			c.Println("Unacceptable topic: " + err.Error() + ".")
			return
		}
	}

	const boardSize = 10
	var messages []*entity.Message
	if topic != nil {
		messages, err = loginHandle.ListTopic(topic, 0, boardSize)
	} else {
		messages, err = loginHandle.ListBoard(0, boardSize)
	}
	if err != nil {
		c.Println("Can't list board: " + err.Error() + ".")
		return
	}

	for i, msg := range messages {
		if i != 0 {
			c.Println()
		}
		c.Printf("#%d by %s, %s\n", i, userSummary(&msg.AuthorID),
			msg.DateWritten.Format(time.RFC3339))
		c.Printf("ID: %s\n", msg.ID())
		if topic == nil {
			c.Printf("Topic: %s\n", msg.Topic.String())
		}
		c.Printf("Subject: %s\n", msg.Subject)
		c.Println(msg.Text)
	}
}

type ThreadPrinter struct {
	c *ishell.Context
}

func (tp *ThreadPrinter) composeIndentation(n *thread.Node) string {
	return strings.Repeat(" ", 4*n.Depth())
}

func (tp *ThreadPrinter) Handle(n *thread.Node) bool {
	m := n.Msg
	if m == nil {
		return true
	}
	if n.IsRoot() {
		tp.c.Printf("Topic: %s\n", m.Topic.String())
	} else {
		tp.c.Println()
	}
	tp.c.Printf("%sSubject: %s\n", tp.composeIndentation(n), m.Subject)
	lines := strings.Split(m.Text, "\n")
	indentedText := strings.Join(lines, "\n"+tp.composeIndentation(n))
	tp.c.Printf("%s%s\n", tp.composeIndentation(n), indentedText)
	tp.c.Printf("%sby %s, %s\n", tp.composeIndentation(n), userSummary(&m.AuthorID),
		m.DateWritten.Format(time.RFC3339))
	tp.c.Printf("%sID: %s\n", tp.composeIndentation(n), m.ID().String())
	return true
}

func doListThread(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	idStr := c.Args[0]
	var tid entity.ID
	err := tid.ParseString(idStr)
	if err != nil {
		c.Println(idStr + " is not a valid entity ID.")
		return
	}
	t, err := loginHandle.ListThread(&tid)
	if err != nil {
		c.Println("Can't list thread: " + err.Error() + ".")
		return
	}
	tp := ThreadPrinter{c}
	tvis := thread.NewViewingVisitor(&tp)
	t.View(tvis)
}

func doMakeReply(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	idStr := c.Args[0]
	var pid entity.ID
	err := pid.ParseString(idStr)
	if err != nil {
		c.Println(idStr + " is not a valid entity ID.")
		return
	}

	c.Print("Enter reply subject: ")
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

	r, err := loginHandle.NewReply(subj, text, &pid)
	if err != nil {
		c.Println("Error making new reply: " + err.Error() + ".")
		return
	}
	err = loginHandle.PostEntity((entity.Entity)(r))
	if err != nil {
		c.Println("Error posting new reply: " + err.Error() + ".")
	} else {
		c.Println("Reply '" + r.String() + "' created successfully.")
	}
}

func doSubscribe(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	topic, err := subs.NewTopic(c.Args[0])
	if err != nil {
		c.Println("Unacceptable topic: " + err.Error() + ".")
		return
	}
	err = loginHandle.Subscribe(topic)
	if err != nil {
		c.Printf("Error subscribing to %s: %v\n", topic, err.Error())
		return
	}
	c.Println("In order to apply changes you need to logout and login back again.")
}

func doUnsubscribe(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	topic, err := subs.NewTopic(c.Args[0])
	if err != nil {
		c.Println("Unacceptable topic: " + err.Error() + ".")
		return
	}
	err = loginHandle.Unsubscribe(topic)
	if err != nil {
		c.Printf("Failed to unsubscribe from %s: %v\n", topic, err.Error())
		return
	}
	c.Println("In order to apply changes you need to logout and login back again.")
}

func doListSubscriptions(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.Help)
		return
	}
	subs := loginHandle.ListSubscriptions()
	c.Print(subs)
}

func doMakeModerator(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	idStr := c.Args[0]
	var id entity.ID
	err := id.ParseString(idStr)
	if err != nil {
		c.Println(idStr + " is not a valid entity ID.")
		return
	}
	err = loginHandle.MakeModerator(&id)
	if err != nil {
		c.Println("Error making new moderator: " + err.Error() + ".")
	}
}

func doRemoveModerator(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 1 {
		c.Println(c.Cmd.Help)
		return
	}
	idStr := c.Args[0]
	var id entity.ID
	err := id.ParseString(idStr)
	if err != nil {
		c.Println(idStr + " is not a valid entity ID.")
		return
	}
	err = loginHandle.RemoveModerator(&id)
	if err != nil {
		c.Println("Error removing moderator: " + err.Error() + ".")
	}
}

func doListModerators(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 0 {
		c.Println(c.Cmd.Help)
		return
	}
	mm := loginHandle.ListModerators()
	for i, mdr := range mm {
		c.Printf("#%d %s\n", i, userSummary(mdr))
	}
}

func makeOperation(c *ishell.Context, typ entity.OperationType) {
	if loginHandle == nil {
		c.Println("You are not logged in.")
		return
	}
	if len(c.Args) != 2 {
		c.Println(c.Cmd.Help)
		return
	}
	idStr := c.Args[0]
	var id entity.ID
	err := id.ParseString(idStr)
	if err != nil {
		c.Println(idStr + " is not a valid entity ID.")
		return
	}
	reasonStr := c.Args[1]
	var reason entity.OperationReason
	err = reason.ParseString(reasonStr)
	if err != nil {
		c.Println(idStr + " is not a valid operation reason.")
		return
	}
	c.Print("Enter optional comment: ")
	comment := c.ReadLine()
	op, err := loginHandle.NewOperation(typ, reason, comment, &id)
	if err != nil {
		c.Println("Error making new operation: " + err.Error() + ".")
	}
	err = loginHandle.PostEntity((entity.Entity)(op))
	if err != nil {
		c.Println("Error posting new operation: " + err.Error() + ".")
	} else {
		c.Println("Operation performed successfully.")
	}
}

func doBanUser(c *ishell.Context) {
	makeOperation(c, entity.OperationTypeBanUser)
}

func doRemoveMessage(c *ishell.Context) {
	makeOperation(c, entity.OperationTypeRemoveMessage)
}

func doListOperations(c *ishell.Context) {
	if loginHandle == nil {
		c.Println("No user is logged in.")
		return
	}
	if len(c.Args) != 2 {
		c.Println(c.Cmd.Help)
		return
	}
	idStr := c.Args[1]
	var id entity.ID
	err := id.ParseString(idStr)
	if err != nil {
		c.Println(idStr + " is not a valid entity ID.")
		return
	}
	entType := c.Args[0]
	var ops []*entity.Operation
	switch entType {
	case "user":
		ops, err = loginHandle.ListOperationsOnUser(&id)
	case "msg":
		ops, err = loginHandle.ListOperationsOnMessage(&id)
	default:
		c.Println(idStr + " is not a valid entity type.")
		c.Println("Expected: 'user' or 'msg'")
		return
	}
	if err != nil {
		c.Println("Error fetching operations: " + err.Error() + ".")
		return
	}
	for i, op := range ops {
		if i != 0 {
			c.Println()
		}
		c.Printf("#%d by %s, %s\n", i, userSummary(&op.AuthorID),
			op.DatePerformed.Format(time.RFC3339))
		c.Printf("ID: %s\n", op.ID())
		c.Printf("Type: %s\n", op.OperationType().String())
		c.Printf("Reason: %s\n", op.Reason.String())
		if op.Comment != "" {
			c.Printf("Comment: %s\n", op.Comment)
		}
	}
}

func doWhoAmI(c *ishell.Context) {
	if len(c.Args) != 0 {
		c.Println(c.Cmd.Help)
		return
	}
	if loginHandle == nil {
		c.Println("No user is logged in.")
		return
	}
	u := loginHandle.GetLoggedUser()
	c.Printf("%s (%s)\n", u.Nickname, u.ID())
}

func doVersion(c *ishell.Context) {
	if len(c.Args) != 0 {
		c.Println(c.Cmd.Help)
		return
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
		fmt.Println(dscuss.Name + " - P2P network for public discussions.")
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
