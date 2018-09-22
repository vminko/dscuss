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
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p"
	"vminko.org/dscuss/p2p/peer"
	"vminko.org/dscuss/sqlite"
	"vminko.org/dscuss/storage"
	dstrings "vminko.org/dscuss/strings"
	"vminko.org/dscuss/subs"
)

const (
	Name                   string = "Dscuss"
	Version                string = "proof-of-concept"
	DefaultDir             string = "~/.dscuss"
	logFileName            string = "dscuss.log"
	cfgFileName            string = "config.json"
	entityDatabaseFileName string = "entity.db"
	addressListFileName    string = "addresses"
	debug                  bool   = true
)

var (
	logFile *os.File
	dir     string
	cfg     *config
	db      *sqlite.Database
	stor    *storage.Storage
	ownr    *owner.Owner
	pp      *p2p.PeerPool
)

func Init(initDir string) error {
	if initDir == "" {
		dir = DefaultDir
	} else {
		dir = initDir
	}

	if dstrings.Truncate(dir, 2) == "~/" {
		usr, err := user.Current()
		if err != nil {
			log.Error("Can't get current OS user: " + err.Error())
			return errors.Internal
		}
		if len(dir) > 2 {
			dir = filepath.Join(usr.HomeDir, dir[2:])
		} else {
			dir = usr.HomeDir
		}
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

	entityDatabasePath := filepath.Join(dir, entityDatabaseFileName)
	db, err = sqlite.Open(entityDatabasePath)
	if err != nil {
		log.Errorf("Can't open entity database file %s: %v", entityDatabasePath, err)
		return errors.Database
	}

	stor = storage.New(db)

	log.Error("Dscuss successfully initialized.")
	return nil
}

func Uninit() {
	if IsLoggedIn() {
		Logout()
	}

	err := db.Close()
	if err != nil {
		log.Errorf("Can't close entity database: %v", err)
	}

	log.Debug("Dscuss uninitialized.")
	logFile.Close()
}

func Register(nickname, info, subStr string) error {
	s, err := subs.ReadString(subStr)
	if err != nil {
		log.Errorf("Attempt to register a user with unacceptable subscriptions: %v", err)
		return errors.WrongArguments

	}
	return owner.Register(dir, nickname, info, s, stor)
}

func Login(nickname string) error {
	if IsLoggedIn() {
		log.Errorf("Login attempt when %s is already logged in", ownr.User.Nickname)
		return errors.AlreadyLoggedIn
	}
	if !entity.IsNicknameValid(nickname) {
		log.Errorf("Login attempt with invalid nickname '%s", nickname)
		return errors.WrongNickname
	}

	var err error
	ownr, err = owner.New(dir, nickname, stor)
	if err != nil {
		log.Errorf("Failed to open %s's data: %v", nickname, err)
		return err
	}
	log.Debugf("Trying to login as peer %s", ownr.User.ID().String())

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

	pp = p2p.NewPeerPool(cp, ownr, stor)
	pp.Start()
	return nil
}

func Logout() {
	if !IsLoggedIn() {
		return
	}
	pp.Stop()
	ownr.Close()
	ownr = nil
}

func IsLoggedIn() bool {
	return ownr != nil
}

func ListPeers() []*peer.Info {
	if !IsLoggedIn() {
		log.Fatal("Attempt to list peers when no user is logged in")
	}
	return pp.ListPeers()
}

func FullVersion() string {
	return fmt.Sprintf("%s version: %s, built with %s.\n", Name, Version, runtime.Version())
}

func Dir() string {
	return dir
}

func NewThread(subj string, text string, topic string) (*entity.Message, error) {
	if !IsLoggedIn() {
		log.Fatal("Attempt to create a new thread when no user is logged in")
	}
	t, err := subs.NewTopic(topic)
	if err != nil {
		return nil, errors.WrongTopic
	}
	return entity.EmergeMessage(subj, text, ownr.User.ID(), &entity.ZeroID, ownr.Signer, t)
}

/*func NewReply(subject string, body string) *entity.Message {
	return
}*/

func PostMessage(m *entity.Message) error {
	if !IsLoggedIn() {
		log.Fatal("Attempt to post message when no user is logged in")
	}
	err := stor.PutMessage(m, nil)
	if err != nil {
		log.Errorf("Failed to post message '%s': %v", m.Desc(), err)
		return err
	}
	return nil
}

func ListBoard(offset, limit int) ([]*entity.Message, error) {
	if !IsLoggedIn() {
		log.Fatal("Attempt to list board when no user is logged in")
	}
	if offset < 0 || limit < 0 {
		return nil, errors.WrongArguments
	}
	return stor.GetRootMessages(offset, limit)
}

func ListTopic(topic string, offset, limit int) ([]*entity.Message, error) {
	if !IsLoggedIn() {
		log.Fatal("Attempt to list board when no user is logged in")
	}
	if offset < 0 || limit < 0 {
		return nil, errors.WrongArguments
	}
	t, err := subs.NewTopic(topic)
	if err != nil {
		return nil, errors.WrongTopic
	}
	return stor.GetTopicMessages(t, offset, limit)
}

func ListSubscriptions() string {
	if !IsLoggedIn() {
		log.Fatal("Attempt to list subscriptions when no user is logged in")
	}
	return ownr.Subs.String()
}
