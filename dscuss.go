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

// Dscuss application implements protocol for creating unstructured pure P2P topic-based
// publish-subscribe network.

// Dscuss package implements root API exposed to the user.
package dscuss

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p"
	"vminko.org/dscuss/p2p/peer"
	dstrings "vminko.org/dscuss/strings"
	"vminko.org/dscuss/subs"
	"vminko.org/dscuss/thread"
)

// LoginHandle implements API exposed to a logged user.
type LoginHandle struct {
	owner *owner.Owner
	pp    *p2p.PeerPool
}

// ByNickname implements sort.Interface for []*peer.Info based on
// the Nickname field.
type ByNickname []*peer.Info

func (a ByNickname) Len() int           { return len(a) }
func (a ByNickname) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNickname) Less(i, j int) bool { return a[i].Nickname < a[j].Nickname }

const (
	Name                string = "Dscuss"
	Version             string = "0.1.0"
	DefaultDir          string = "~/.dscuss"
	logFileName         string = "dscuss.log"
	cfgFileName         string = "config.json"
	AddressListFileName string = "addresses.txt"
	debug               bool   = true
)

var (
	logFile *os.File
	dir     string
	cfg     *config
	login   *LoginHandle
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
	if debug {
		log.EnableDebug()
	}

	var cfgPath string = filepath.Join(dir, cfgFileName)
	cfg, err = newConfig(cfgPath)
	if err != nil {
		log.Error("Can't process config file: " + err.Error())
		return err
	}

	log.Info("Dscuss successfully initialized.")
	return nil
}

func Uninit() {
	if IsLoggedIn() {
		login.Logout()
	}
	log.Debug("Dscuss uninitialized.")
	logFile.Close()
}

func FullVersion() string {
	return fmt.Sprintf("%s library version: %s, protocol version: %d, built with %s.\n",
		Name, Version, peer.ProtocolVersion, runtime.Version())
}

func Dir() string {
	return dir
}

func IsLoggedIn() bool {
	return login != nil
}

func Register(nickname, info string, s subs.Subscriptions) error {
	return owner.Register(dir, nickname, info, s)
}

func Login(nickname string) (*LoginHandle, error) {
	if IsLoggedIn() {
		log.Errorf("Login attempt when %s is already logged in", login.owner.User.Nickname)
		return nil, errors.AlreadyLoggedIn
	}
	if !entity.IsNicknameValid(nickname) {
		log.Errorf("Login attempt with invalid nickname '%s", nickname)
		return nil, errors.WrongNickname
	}

	ownr, err := owner.New(dir, nickname)
	if err != nil {
		log.Errorf("Failed to open %s's data: %v", nickname, err)
		return nil, err
	}
	log.Debugf("Trying to login as peer %s", ownr.User.ID().String())

	var aps []p2p.AddressProvider
	for _, name := range strings.Split(cfg.Network.AddressProvider, ",") {
		switch name {
		case "addrlist":
			addrFilePath := filepath.Join(
				dir,
				AddressListFileName,
			)
			ap := p2p.NewAddressList(addrFilePath)
			aps = append(aps, ap)
		case "dht":
			ap := p2p.NewDHTCrawler(
				cfg.Network.Address,
				cfg.Network.DHTPort,
				cfg.Network.DHTBootstrap,
				cfg.Network.Port,
				ownr.Profile.GetSubscriptions(),
			)
			aps = append(aps, ap)
		default:
			log.Error("Unknown address provider is configured: " + name)
		}
	}
	if len(aps) == 0 {
		log.Fatal("Could not found any valid address provider in " + cfg.Network.AddressProvider)
	}

	hp := net.JoinHostPort(cfg.Network.Address, strconv.Itoa(cfg.Network.Port))
	cp := p2p.NewConnectionProvider(aps, hp, cfg.Network.MaxInConnCount, cfg.Network.MaxOutConnCount)

	pp := p2p.NewPeerPool(cp, ownr)
	pp.Start()

	login = &LoginHandle{ownr, pp}
	return login, nil
}

func (lh *LoginHandle) Logout() {
	if lh != login {
		log.Fatal("Unknown LoginHandle")
	}
	login = nil
	lh.pp.Stop()
	lh.owner.Close()
}

func (lh *LoginHandle) GetLoggedUser() *entity.User {
	return lh.owner.User
}

func (lh *LoginHandle) ListPeers() []*peer.Info {
	return lh.pp.ListPeers()
}

func (lh *LoginHandle) NewThread(subj, text string, topic subs.Topic) (*entity.Message, error) {
	return entity.EmergeMessage(
		subj,
		text,
		lh.owner.User.ID(),
		&entity.ZeroID,
		lh.owner.Signer,
		topic,
	)
}

func (lh *LoginHandle) NewReply(subj, text string, parent *entity.ID) (*entity.Message, error) {
	return entity.EmergeMessage(subj, text, lh.owner.User.ID(), parent, lh.owner.Signer, nil)
}

func (lh *LoginHandle) NewOperation(
	typ entity.OperationType,
	reason entity.OperationReason,
	comment string,
	objectID *entity.ID,
) (*entity.Operation, error) {
	return entity.EmergeOperation(
		typ,
		reason,
		comment,
		lh.owner.User.ID(),
		objectID,
		lh.owner.Signer,
	)
}

func (lh *LoginHandle) PostEntity(e entity.Entity) error {
	return lh.owner.Storage.PutEntity(e, nil)
}

func (lh *LoginHandle) GetUser(id *entity.ID) (*entity.User, error) {
	return lh.owner.Storage.GetUser(id)
}

func (lh *LoginHandle) ListOperationsOnUser(id *entity.ID) ([]*entity.Operation, error) {
	return lh.owner.Storage.GetOperationsOnUser(id)
}

func (lh *LoginHandle) ListOperationsOnMessage(id *entity.ID) ([]*entity.Operation, error) {
	return lh.owner.Storage.GetOperationsOnMessage(id)
}

func (lh *LoginHandle) ListBoard(offset, limit int) ([]*entity.Message, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.WrongArguments
	}
	mm, err := lh.owner.Storage.GetRootMessages(offset, limit)
	if err != nil {
		log.Error("Failed to fetch root messages from the storage " + err.Error())
		return nil, err
	}
	return lh.owner.View.ModerateMessages(mm)
}

func (lh *LoginHandle) ListTopic(topic subs.Topic, offset, limit int) ([]*entity.Message, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.WrongArguments
	}
	mm, err := lh.owner.Storage.GetTopicMessages(topic, offset, limit)
	if err != nil {
		log.Error("Failed to get topic from the storage " + err.Error())
		return nil, err
	}
	return lh.owner.View.ModerateMessages(mm)
}

// TBD: add offset and limit
func (lh *LoginHandle) ListThread(id *entity.ID) (*thread.Node, error) {
	t, err := lh.owner.Storage.GetThread(id)
	if err != nil {
		log.Error("Failed to get topic from the storage " + err.Error())
		return nil, err
	}
	return lh.owner.View.ModerateThread(t)
}

func (lh *LoginHandle) Subscribe(topic subs.Topic) error {
	return lh.owner.Profile.PutSubscription(topic)
}

func (lh *LoginHandle) Unsubscribe(topic subs.Topic) error {
	return lh.owner.Profile.RemoveSubscription(topic)
}

func (lh *LoginHandle) ListSubscriptions() string {
	return lh.owner.Profile.GetSubscriptions().String()
}

func (lh *LoginHandle) MakeModerator(id *entity.ID) error {
	has, err := lh.owner.Storage.HasUser(id)
	if err != nil {
		log.Errorf("Failed to check if storage contains %s: %v", id.Shorten(), err)
		return err
	}
	if !has {
		log.Errorf("Attempt to make unknown user %s a moderator", id.Shorten())
		return errors.NoSuchUser
	}
	return lh.owner.Profile.PutModerator(id)
}

func (lh *LoginHandle) RemoveModerator(id *entity.ID) error {
	return lh.owner.Profile.RemoveModerator(id)
}

func (lh *LoginHandle) ListModerators() []*entity.ID {
	return lh.owner.Profile.GetModerators()
}

func (lh *LoginHandle) ListUserHistory() []*entity.UserHistory {
	return lh.owner.Profile.GetFullHistory()
}
