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
	"encoding/binary"
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
	cfg *config
}

const (
	Name                   string = "Dscuss"
	Version                string = "proof-of-concept"
	DefaultCfgDir          string = "~/.dscuss"
	debug                  bool   = true
	logFileName            string = "dscuss.log"
	cfgFileName            string = "config"
	privKeyFileName        string = "privkey.pem"
	powFileName            string = "proof_of_work"
	globalDatabaseFileName string = "global.db"
)

var (
	li      *loginInfo
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
	Logf(DEBUG, "Using the following user directory: %s", userDir)

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

	privkey, err := newPrivateKey()
	if err != nil {
		Logf(ERROR, "Can't generate new private key: %v", err)
		return ErrInternal
	}
	privkeyPEM, err := privkey.encode()
	if err != nil {
		Logf(ERROR, "Can't encode the private key to PEM format: %v", err)
		return ErrInternal
	}
	privkeyPath := filepath.Join(userDir, privKeyFileName)
	err = ioutil.WriteFile(privkeyPath, privkeyPEM, 0600)
	if err != nil {
		Logf(ERROR, "Can't save private key as file %s: %v", privkeyPath, err)
		return ErrFilesystem
	}

	pubPem, err := privkey.public().encode()
	if err != nil {
		Logf(ERROR, "Can't encode public key %v", err)
		return err
	}
	pow := newPowFinder(pubPem)
	proof := pow.find()
	pbuf := make([]byte, 8)
	binary.BigEndian.PutUint64(pbuf, uint64(proof))
	powPath := filepath.Join(userDir, powFileName)
	err = ioutil.WriteFile(privkeyPath, pbuf, 0644)
	if err != nil {
		Logf(ERROR, "Can't save proof-of-work as file %s: %v", powPath, err)
		return ErrFilesystem
	}

	user, err := EmergeUser(nickname, info, proof, time.Now(), &Signer{privkey})
	if err != nil {
		Logf(ERROR, "Can't create new user %s: %v", nickname, err)
		return err
	}

	globalDatabasePath := filepath.Join(cfgDir, globalDatabaseFileName)
	db, err := open(globalDatabasePath)
	if err != nil {
		Logf(ERROR, "Can't open the database file %s: %v", powPath, err)
		return ErrDatabase
	}
	err = db.putUser(user)
	if err != nil {
		Logf(ERROR, "Can't add user '%s' to the database: %v", user.Nickname, err)
		return ErrDatabase
	}

	return nil
}
