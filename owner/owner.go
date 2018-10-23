/*
This file is part of Dscuss.
Copyright (C) 2018  Vitaly Minko

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

// Owner is the user which owns the current network node. Unlike other users, it
// has access to the Storage of this node and the private key of the user.
package owner

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/sqlite"
	"vminko.org/dscuss/storage"
	"vminko.org/dscuss/subs"
)

type Owner struct {
	User    *entity.User
	Subs    subs.Subscriptions
	Profile *Profile
	Signer  *crypto.Signer
	View    *View
	storage *storage.Storage
}

const (
	privKeyFileName         string = "privkey.pem"
	subscriptionsFileName   string = "subscriptions.txt"
	profileDatabaseFileName string = "profile.db"
)

func Register(dir, nickname, info string, subs subs.Subscriptions, s *storage.Storage) error {
	log.Debugf("Registering user %s", nickname)
	// Nickname will be validated via regexp later during EmergeUser
	if nickname == "" {
		return errors.WrongNickname
	}
	userDir := filepath.Join(dir, nickname)
	log.Debugf("Register uses the following user directory: %s", userDir)

	err := os.Mkdir(userDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			log.Infof("Looks like the user %s is already registered", nickname)
			return errors.AlreadyRegistered
		} else {
			log.Errorf("Can't create directory " + userDir + ": " + err.Error())
			return errors.Filesystem
		}
	}

	subsPath := filepath.Join(userDir, subscriptionsFileName)
	err = ioutil.WriteFile(subsPath, []byte(subs.String()), 0640)
	if err != nil {
		log.Errorf("Can't save user subscriptions as file %s: %v", subsPath, err)
		return errors.Filesystem
	}

	privKey, err := crypto.NewPrivateKey()
	if err != nil {
		log.Errorf("Can't generate new private key: %v", err)
		return errors.Internal
	}
	privKeyPEM := privKey.EncodeToPEM()
	privKeyPath := filepath.Join(userDir, privKeyFileName)
	err = ioutil.WriteFile(privKeyPath, privKeyPEM, 0600)
	if err != nil {
		log.Errorf("Can't save private key as file %s: %v", privKeyPath, err)
		return errors.Filesystem
	}

	pow := crypto.NewPowFinder(privKey.Public().EncodeToDER())
	proof := pow.Find()

	user, err := entity.EmergeUser(nickname, info, proof, crypto.NewSigner(privKey))
	if err != nil {
		log.Errorf("Can't create user '%s': %v", nickname, err)
		return err
	}
	log.Debugf("Dumping emerged User %s:", nickname)
	log.Debug(user.String())

	err = s.PutEntity((entity.Entity)(user), nil)
	if err != nil {
		log.Errorf("Can't add user '%s' to the storage: %v", user.Nickname(), err)
		return errors.Database
	}

	return nil
}

func New(dir, nickname string, stor *storage.Storage) (*Owner, error) {
	userDir := filepath.Join(dir, nickname)
	log.Debugf("Owner uses the following user directory: %s", userDir)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		log.Warningf("User directory '%s' does not exist", userDir)
		return nil, errors.NoSuchUser
	}

	privKeyPath := filepath.Join(userDir, privKeyFileName)
	privKeyPem, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Errorf("Can't read private key from file %s: %v", privKeyPath, err)
		return nil, errors.Filesystem
	}

	privKey, err := crypto.ParsePrivateKeyFromPEM(privKeyPem)
	if err != nil {
		log.Errorf("Error parsing private key from file %s: %v", privKeyPath, err)
		return nil, err
	}

	eid := entity.NewID(privKey.Public().EncodeToDER())
	u, err := stor.GetUser(&eid)
	if err != nil {
		log.Errorf("Can't fetch the user with id '%x' from the storage: %v", eid, err)
		return nil, err
	}
	log.Debug("Dumping fetched User:")
	log.Debug(u.String())

	subsPath := filepath.Join(userDir, subscriptionsFileName)
	sub, err := subs.ReadFile(subsPath)
	if err != nil {
		log.Errorf("Error reading subscriptions of the user '%s': %v", nickname, err)
		return nil, err
	}

	profileDatabasePath := filepath.Join(userDir, profileDatabaseFileName)
	db, err := sqlite.OpenProfileDatabase(profileDatabasePath)
	if err != nil {
		log.Errorf("Can't open profile database file %s: %v", profileDatabasePath, err)
		return nil, errors.Database
	}

	prf := NewProfile(db, u.ID())
	return &Owner{
		User:    u,
		Subs:    sub,
		Profile: prf,
		Signer:  crypto.NewSigner(privKey),
		View:    NewView(prf, stor),
		storage: stor,
	}, nil
}

func (o *Owner) Close() {
	err := o.Profile.Close()
	if err != nil {
		log.Errorf("Error closing entity storage: %v", err)
	}
}
