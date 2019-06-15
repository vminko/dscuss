/*
This file is part of Dscuss.
Copyright (C) 2019  Vitaly Minko

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

// Web interface for Dscuss. It's supposed to the primary UI for end used.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/controller"
	"vminko.org/dscuss/log"
)

const (
	webVersion         string = "0.1"
	webDefaultPassword string = "qwerty"
	webPort            int    = 8080
)

var (
	argConfig   = flag.String("config", dscuss.DefaultDir, "Directory with config files to use")
	argVersion  = flag.Bool("version", false, "Display version of the program and exit")
	argHelp     = flag.Bool("help", false, "Print help message and exit")
	argUser     = flag.String("user", "adam", "Name of the user to log in as")
	argPassword = flag.String("password", webDefaultPassword, "Password to protect access to the Web UI")
	// Looks like there is no way to pass LoginHandle via ishell.Context.
	loginHandle *dscuss.LoginHandle
)

func getVersion() string {
	v := dscuss.FullVersion()
	v += fmt.Sprintf("Web UI version: %s.", webVersion)
	return v
}

func setupQuitSignalHandler() {
	sigChan := make(chan os.Signal)
	go func() {
		stacktrace := make([]byte, 8192)
		for _ = range sigChan {
			length := runtime.Stack(stacktrace, true)
			fmt.Println(string(stacktrace[:length]))
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}

func setupTermSignalHandler() {
	sigChan := make(chan os.Signal)
	go func() {
		for _ = range sigChan {
			log.Debug("Stopping Dscuss...")
			dscuss.Uninit()
			fmt.Println("Dscuss stopped.")
		}
	}()
	signal.Notify(sigChan, syscall.SIGTERM)
}

func setupSignalHandlers() {
	setupQuitSignalHandler()
	setupTermSignalHandler()
}

func main() {
	setupSignalHandlers()
	rand.Seed(time.Now().UnixNano())

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
	if *argPassword == webDefaultPassword {
		fmt.Println("You have to specify a custom password.")
		return
	}
	if len(*argPassword) > controller.MaxPasswordLen {
		fmt.Print("The specified password is too long. ")
		fmt.Printf("Max password length = %d.\n", controller.MaxPasswordLen)
		return
	}
	controller.SetPassword(*argPassword)

	fmt.Println("Starting Dscuss...")
	err := dscuss.Init(*argConfig)
	if err != nil {
		panic("Could not initialize Dscuss: " + err.Error())
	}
	log.Debugf("Using Web UI version %s.", webVersion)

	loginHandle, err = dscuss.Login(*argUser)
	if err != nil {
		log.Errorf("Failed to log in as %s: %v\n", *argUser, err)
		return
	}

	http.HandleFunc("/", controller.MakeRootHandler(loginHandle))
	http.HandleFunc("/static/dscuss.css", controller.CSSHandler)
	http.HandleFunc("/static/dscuss.js", controller.JavaScriptHandler)
	http.HandleFunc("/login", controller.MakeLoginHandler(loginHandle))
	http.HandleFunc("/logout", controller.LogoutHandler)
	http.HandleFunc("/profile", controller.MakeProfileHandler(loginHandle))
	http.HandleFunc("/board", controller.MakeBoardHandler(loginHandle))
	http.HandleFunc("/thread", controller.MakeThreadHandler(loginHandle))
	http.HandleFunc("/reply", controller.MakeReplyHandler(loginHandle))
	http.HandleFunc("/start", controller.MakeStartHandler(loginHandle))
	http.HandleFunc("/addmdr", controller.MakeAddModeratorHandler(loginHandle))
	http.HandleFunc("/rmmdr", controller.MakeRemoveModeratorHandler(loginHandle))
	http.HandleFunc("/sub", controller.MakeSubscribeHandler(loginHandle))
	http.HandleFunc("/unsub", controller.MakeUnsubscribeHandler(loginHandle))
	http.HandleFunc("/rmmsg", controller.MakeRemoveMessageHandler(loginHandle))
	http.HandleFunc("/ban", controller.MakeBanHandler(loginHandle))

	log.Debugf("Starting HTTP server on port %d\n", webPort)
	http.ListenAndServe(":"+strconv.Itoa(webPort), nil)
}
