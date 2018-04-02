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

// Dscuss implements protocol for creating unstructured pure P2P topic-based
// publish-subscribe network.
package dscuss

// Main backend file. Implements API exposed to the user.

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
	"vminko.org/dscuss/log"
)

type ctxKey string
type loginContext struct {
	user   *User
	gDB    *globalDB
	signer *Signer
}

const (
	Name                   string = "Dscuss"
	Version                string = "proof-of-concept"
	DefaultDir             string = "~/.dscuss"
	logFileName            string = "dscuss.log"
	cfgFileName            string = "config.json"
	privKeyFileName        string = "privkey.pem"
	globalDatabaseFileName string = "global.db"
	debug                  bool   = true
)

var (
	logFile  *os.File
	dir      string
	cfg      *config
	loginCtx *loginContext
	pp       *peerPool
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
			return ErrInternal
		}
		dir = filepath.Join(usr.HomeDir, dir[2:])
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			log.Error("Can't create directory " + dir + ": " + err.Error())
			return ErrFilesystem
		}
	}

	var logPath string = filepath.Join(dir, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Error("Can't open log file: " + err.Error())
		return ErrFilesystem
	}
	log.SetOutput(logFile)
	log.SetDebug(debug)

	var cfgPath string = filepath.Join(dir, cfgFileName)
	cfg, err = newConfig(cfgPath)
	if err != nil {
		log.Error("Can't process config file: " + err.Error())
		return err
	}

	log.Error("Dscuss successfully iniitialized.")
	return nil
}

func Uninit() {
	if IsLoggedIn() {
		Logout()
	}
	log.Debug("Dscuss successfully uniniitialized.")
	logFile.Close()
}

func FullVersion() string {
	return fmt.Sprintf("%s version: %s, built with %s.\n", Name, Version, runtime.Version())
}

func Dir() string {
	return dir
}

func Register(nickname, info string) error {
	// TBD: validate nickname. It must contain only [\w\d\._]
	log.Debugf("Registering user %s", nickname)
	if nickname == "" {
		return ErrWrongNickname
	}
	userDir := filepath.Join(dir, nickname)
	log.Debugf("Register uses the following user directory: %s", userDir)

	err := os.Mkdir(userDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			log.Infof("Looks like the user %s is already registered", nickname)
			return ErrAlreadyRegistered
		} else {
			log.Errorf("Can't create directory " + userDir + ": " + err.Error())
			return ErrFilesystem
		}
	}

	privKey, err := newPrivateKey()
	if err != nil {
		log.Errorf("Can't generate new private key: %v", err)
		return ErrInternal
	}
	privKeyPEM := privKey.encodeToPEM()
	privKeyPath := filepath.Join(userDir, privKeyFileName)
	err = ioutil.WriteFile(privKeyPath, privKeyPEM, 0600)
	if err != nil {
		log.Errorf("Can't save private key as file %s: %v", privKeyPath, err)
		return ErrFilesystem
	}

	pow := newPowFinder(privKey.public().encodeToDER())
	log.Info(string(privKey.public().encodeToPEM()))
	proof := pow.find()
	user, err := emergeUser(nickname, info, proof, time.Now(), &Signer{privKey})
	if err != nil {
		log.Errorf("Can't create new user %s: %v", nickname, err)
		return err
	}
	if debug {
		log.Debugf("Dumping emerged User %s:", nickname)
		log.Debug(user.String())
	}

	globalDatabasePath := filepath.Join(userDir, globalDatabaseFileName)
	db, err := open(globalDatabasePath)
	if err != nil {
		log.Errorf("Can't open global database file %s: %v", globalDatabasePath, err)
		return ErrDatabase
	}
	err = db.putUser(user)
	if err != nil {
		log.Errorf("Can't add user '%s' to the database: %v", user.Nickname, err)
		return ErrDatabase
	}
	err = db.close()
	if err != nil {
		log.Errorf("Can't close global database: %v", err)
		return ErrDatabase
	}

	return nil
}

func IsLoggedIn() bool {
	return loginCtx != nil
}

func Login(nickname string) error {
	if loginCtx != nil {
		log.Errorf("Login attempt when %s is already logged in", loginCtx.user.Nickname)
		return ErrAlreadyLoggedIn
	}

	userDir := filepath.Join(dir, nickname)
	log.Debugf("Login uses the following user directory: %s", userDir)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return ErrFilesystem
	}

	privKeyPath := filepath.Join(userDir, privKeyFileName)
	privKeyPem, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Errorf("Can't read private key from file %s: %v", privKeyPath, err)
		return ErrFilesystem
	}

	privKey, err := parsePrivateKeyFromPEM(privKeyPem)
	if err != nil {
		log.Errorf("Error parsing private key from file %s: %v", privKeyPath, err)
		return err
	}

	eid := newEntityID(privKey.public().encodeToDER())
	globalDatabasePath := filepath.Join(userDir, globalDatabaseFileName)
	db, err := open(globalDatabasePath)
	if err != nil {
		log.Errorf("Can't open global database file %s: %v", globalDatabasePath, err)
		return ErrDatabase
	}
	u, err := db.getUser(&eid)
	if err != nil {
		log.Errorf("Can't fetch the user with id '%x' from the database: %v", eid, err)
		return err
	}
	if debug {
		log.Debug("Dumping fetched User:")
		log.Debug(u.String())
	}
	/* TBD:
	   read subscriptions
	   initialize network subsystem
	*/

	loginCtx = &loginContext{
		user:   u,
		gDB:    db,
		signer: &Signer{privKey},
	}

	pp = newPeerPool(cfg, dir, loginCtx)
	pp.start()

	return nil
}

func Logout() error {
	if !IsLoggedIn() {
		return nil
	}
	err := loginCtx.gDB.close()
	if err != nil {
		log.Errorf("Can't close global database: %v", err)
		return ErrDatabase
	}
	pp.stop()
	loginCtx = nil
	return nil
}
