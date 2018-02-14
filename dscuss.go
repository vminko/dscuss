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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
)

type loginInfo struct {
	user    *User
	gDB     *globalDB
	privKey *privateKey
}

const (
	Name                   string = "Dscuss"
	Version                string = "proof-of-concept"
	DefaultCfgDir          string = "~/.dscuss"
	debug                  bool   = false
	logFileName            string = "dscuss.log"
	cfgFileName            string = "config"
	privKeyFileName        string = "privkey.pem"
	globalDatabaseFileName string = "global.db"
)

var (
	logInf  *loginInfo
	logFile *os.File
	cfgDir  string
)

func Init(dir string) error {
	if dir == "" {
		cfgDir = DefaultCfgDir
	} else {
		cfgDir = dir
	}

	if cfgDir[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			Log(ERROR, "Can't get get current OS user: "+err.Error())
			return ErrInternal
		}
		cfgDir = filepath.Join(usr.HomeDir, cfgDir[2:])
	}

	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		err = os.MkdirAll(cfgDir, 0700)
		if err != nil {
			Log(ERROR, "Can't create directory "+cfgDir+": "+err.Error())
			return ErrFilesystem
		}
	}

	var logPath string = filepath.Join(cfgDir, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Log(ERROR, "Can't open log file: "+err.Error())
		return ErrFilesystem
	}
	log.SetOutput(logFile)
	Log(DEBUG, "Dscuss successfully iniitialized.")
	return nil
}

func Uninit() {
	logFile.Close()
	Log(DEBUG, "Dscuss successfully uniniitialized.")
}

func FullVersion() string {
	return fmt.Sprintf("%s version: %s, built with %s.\n", Name, Version, runtime.Version())
}

func CfgDir() string {
	return cfgDir
}

func Register(nickname, info string) error {
	// TBD: validate nickname. It must contain only [\w\d\._]
	Logf(DEBUG, "Registering user %s", nickname)
	if nickname == "" {
		return ErrWrongNickname
	}
	userDir := filepath.Join(cfgDir, nickname)
	Logf(DEBUG, "Register uses the following user directory: %s", userDir)

	err := os.Mkdir(userDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			Logf(INFO, "Looks like the user %s is already registered", nickname)
			return ErrAlreadyRegistered
		} else {
			Log(ERROR, "Can't create directory "+userDir+": "+err.Error())
			return ErrFilesystem
		}
	}

	privKey, err := newPrivateKey()
	if err != nil {
		Logf(ERROR, "Can't generate new private key: %v", err)
		return ErrInternal
	}
	privKeyPEM := privKey.encodeToPEM()
	privKeyPath := filepath.Join(userDir, privKeyFileName)
	err = ioutil.WriteFile(privKeyPath, privKeyPEM, 0600)
	if err != nil {
		Logf(ERROR, "Can't save private key as file %s: %v", privKeyPath, err)
		return ErrFilesystem
	}

	pow := newPowFinder(privKey.public().encodeToDER())
	Log(INFO, string(privKey.public().encodeToPEM()))
	proof := pow.find()
	user, err := emergeUser(nickname, info, proof, time.Now(), &Signer{privKey})
	if err != nil {
		Logf(ERROR, "Can't create new user %s: %v", nickname, err)
		return err
	}
	if debug {
		juser, err := json.Marshal(user)
		if err != nil {
			Logf(ERROR, "Can't marshall %v", err)
		}
		Logf(DEBUG, "Dumping emerged User %s:", nickname)
		Log(DEBUG, string(juser))
	}

	globalDatabasePath := filepath.Join(userDir, globalDatabaseFileName)
	db, err := open(globalDatabasePath)
	if err != nil {
		Logf(ERROR, "Can't open global database file %s: %v", globalDatabasePath, err)
		return ErrDatabase
	}
	err = db.putUser(user)
	if err != nil {
		Logf(ERROR, "Can't add user '%s' to the database: %v", user.Nickname, err)
		return ErrDatabase
	}
	err = db.close()
	if err != nil {
		Logf(ERROR, "Can't close global database: %v", err)
		return ErrDatabase
	}

	return nil
}

func IsLoggedIn() bool {
	return logInf != nil
}

func Login(nickname string) error {
	if logInf != nil {
		Logf(ERROR, "Login attempt when %s is already logged in", logInf.user.Nickname)
		return ErrAlreadyLoggedIn
	}

	userDir := filepath.Join(cfgDir, nickname)
	Logf(DEBUG, "Login uses the following user directory: %s", userDir)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return ErrFilesystem
	}

	privKeyPath := filepath.Join(userDir, privKeyFileName)
	privKeyPem, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		Logf(ERROR, "Can't read private key from file %s: %v", privKeyPath, err)
		return ErrFilesystem
	}

	privKey, err := parsePrivateKeyFromPEM(privKeyPem)
	if err != nil {
		Logf(ERROR, "Error parsing private key from file %s: %v", privKeyPath, err)
		return err
	}

	eid := newEntityID(privKey.public().encodeToDER())
	globalDatabasePath := filepath.Join(userDir, globalDatabaseFileName)
	db, err := open(globalDatabasePath)
	if err != nil {
		Logf(ERROR, "Can't open global database file %s: %v", globalDatabasePath, err)
		return ErrDatabase
	}
	u, err := db.getUser(&eid)
	if err != nil {
		Logf(ERROR, "Can't fetch the user with id '%x' from the database: %v", eid, err)
		return err
	}
	if debug {
		juser, _ := json.Marshal(u)
		Log(DEBUG, "Dumping fetched User:")
		Log(DEBUG, string(juser))
	}
	/* TBD:
	   read subscriptions
	   initialize network subsystem
	*/

	logInf = &loginInfo{
		user:    u,
		gDB:     db,
		privKey: privKey,
	}

	return nil
}

func Logout() error {
	err := logInf.gDB.close()
	if err != nil {
		Logf(ERROR, "Can't close global database: %v", err)
		return ErrDatabase
	}
	return nil
}
