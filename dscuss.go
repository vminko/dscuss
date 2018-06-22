/*
This file is part of Dscuss.
Copyright (C) 2017-2018  Vitaly Minko

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

// Dscuss implements protocol for creating unstructured pure P2P topic-based
// publish-subscribe network.
package dscuss

// Main backend file. Implements API exposed to the user.

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p"
)

const (
	Name                   string = "Dscuss"
	Version                string = "proof-of-concept"
	DefaultDir             string = "~/.dscuss"
	logFileName            string = "dscuss.log"
	cfgFileName            string = "config.json"
	privKeyFileName        string = "privkey.pem"
	globalDatabaseFileName string = "global.db"
	addressListFileName    string = "addresses"
	debug                  bool   = true
)

var (
	logFile *os.File
	dir     string
	cfg     *config
	ownr    *owner.Owner
	pp      *p2p.PeerPool
)

func Init(initDir string) error {
	if initDir == "" {
		dir = DefaultDir
	} else {
		dir = initDir
	}

	if dir[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			log.Error("Can't get get current OS user: " + err.Error())
			return errors.Internal
		}
		dir = filepath.Join(usr.HomeDir, dir[2:])
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			log.Error("Can't create directory " + dir + ": " + err.Error())
			return errors.Filesystem
		}
	}

	var logPath string = filepath.Join(dir, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Error("Can't open log file: " + err.Error())
		return errors.Filesystem
	}
	log.SetOutput(logFile)
	log.SetDebug(debug)

	var cfgPath string = filepath.Join(dir, cfgFileName)
	cfg, err = newConfig(cfgPath)
	if err != nil {
		log.Error("Can't process config file: " + err.Error())
		return err
	}

	log.Error("Dscuss successfully initialized.")
	return nil
}

func Uninit() {
	if IsLoggedIn() {
		Logout()
	}
	log.Debug("Dscuss successfully uninitialized.")
	logFile.Close()
}

func Register(nickname, info string) error {
	return owner.Register(dir, nickname, info)
}

func Login(nickname string) error {
	if IsLoggedIn() {
		log.Errorf("Login attempt when %s is already logged in", ownr.User.Nickname)
		return errors.AlreadyLoggedIn
	}

	var err error
	ownr, err = owner.New(dir, nickname)
	if err != nil {
		log.Errorf("Failed to open %s's data: %v", nickname, err)
		return err
	}
	log.Debugf("Trying to login as peer %s", ownr.User.ID.String())

	var ap p2p.AddressProvider
	switch cfg.Network.AddressProvider {
	case "addrlist":
		addrFilePath := filepath.Join(
			dir,
			addressListFileName,
		)
		ap = p2p.NewAddressList(addrFilePath)
	/* TBD:
	case "dht":
	case "dns":
	*/
	default:
		log.Fatal("Unknown address provider is configured: " + cfg.Network.AddressProvider)
	}

	cp := p2p.NewConnectionProvider(
		ap, cfg.Network.HostPort,
		cfg.Network.MaxInConnCount, cfg.Network.MaxOutConnCount,
	)

	pp = p2p.NewPeerPool(cp, ownr)
	pp.Start()
	return nil
}

func Logout() error {
	if !IsLoggedIn() {
		return nil
	}
	ownr.Close()
	pp.Stop()
	ownr = nil
	return nil
}

func IsLoggedIn() bool {
	return ownr != nil
}

func ListPeers() []*p2p.PeerInfo {
	return pp.ListPeers()
}

func FullVersion() string {
	return fmt.Sprintf("%s version: %s, built with %s.\n", Name, Version, runtime.Version())
}

func Dir() string {
	return dir
}
