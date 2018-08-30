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

package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

// Database stores Entities. Implements EntityStorage interface.
type Database sql.DB

func Open(fileName string) (*Database, error) {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		log.Errorf("Unable to open SQLite connection: %s", err.Error())
		return nil, errors.CantOpenDB
	}

	var execErr error
	exec := func(req string) {
		if execErr != nil {
			return
		}
		_, execErr = db.Exec(req)
	}
	exec("PRAGMA temp_store=MEMORY")
	exec("PRAGMA synchronous=OFF")
	exec("PRAGMA locking_mode=EXCLUSIVE")
	exec("PRAGMA page_size=4092")
	exec("CREATE TABLE IF NOT EXISTS  User (" +
		"  Id              BLOB PRIMARY KEY," +
		"  Public_key      BLOB NOT NULL," +
		"  Proof           UNSIGNED BIG INT NOT NULL," +
		"  Nickname        TEXT NOT NULL," +
		"  Info            TEXT," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Signature       BLOB NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS  Message (" +
		"  Id              BLOB PRIMARY KEY," +
		"  Subject         TEXT," +
		"  Content         TEXT," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Author_id       BLOB NOT NULL," +
		"  Parent_id       BLOB NOT NULL," +
		"  Signature       BLOB NOT NULL," +
		"  FOREIGN KEY (Author_id) REFERENCES User(Id))")
	exec("CREATE TABLE IF NOT EXISTS  Operation (" +
		"  Id              BLOB PRIMARY KEY," +
		"  Type            INTEGER NOT NULL," +
		"  Reason          INTEGER NOT NULL," +
		"  Comment         TEXT," +
		"  Author_id       BLOB NOT NULL," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Signature       BLOB NOT NULL," +
		"  FOREIGN KEY (Author_id) REFERENCES User(Id))")
	exec("CREATE TABLE IF NOT EXISTS  Operation_on_User (" +
		"  Operation_id    BLOB NOT NULL," +
		"  User_id         BLOB NOT NULL," +
		"  FOREIGN KEY (Operation_id) REFERENCES Operation(Id)," +
		"  FOREIGN KEY (User_id) REFERENCES User(Id))")
	exec("CREATE TABLE IF NOT EXISTS  Operation_on_Message (" +
		"  Operation_id    BLOB NOT NULL," +
		"  Message_id      BLOB NOT NULL," +
		"  FOREIGN KEY (Operation_id) REFERENCES Operation(Id)," +
		"  FOREIGN KEY (Message_id) REFERENCES Message(Id))")
	exec("CREATE TABLE IF NOT EXISTS  Tag (" +
		"  Id              INTEGER PRIMARY KEY AUTOINCREMENT," +
		"  Name            TEXT NOT NULL UNIQUE ON CONFLICT IGNORE)")
	exec("CREATE TABLE IF NOT EXISTS  Message_Tag (" +
		"  Tag_id          INTEGER NOT NULL," +
		"  Message_id      BLOB NOT NULL," +
		"  FOREIGN KEY (Tag_id) REFERENCES Tag(Id)," +
		"  FOREIGN KEY (Message_id) REFERENCES Message(Id)," +
		"  UNIQUE (Tag_id, Message_id))")
	// TBD: create indexes?
	if execErr != nil {
		log.Errorf("Unable to initialize the database: %s", execErr.Error())
		return nil, errors.DBOperFailed
	}

	return (*Database)(db), nil
}

func (d *Database) Close() error {
	db := (*sql.DB)(d)
	err := db.Close()
	if err != nil {
		log.Errorf("Unable to close the database: %v", err)
		return errors.DBOperFailed
	}
	return nil
}

func (d *Database) PutUser(user *entity.User) error {
	log.Debugf("Adding user `%s' to the database.", user.Nickname())

	query := `
	INSERT INTO User
	( Id,
	  Public_key,
	  Proof,
	  Nickname,
	  Info,
	  Timestamp,
	  Signature )
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	db := (*sql.DB)(d)
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatalf("Error preparing 'putUser' statement: %v", err)
	}

	pkpem := user.PubKey.EncodeToDER()
	_, err = stmt.Exec(
		user.ID()[:],
		pkpem,
		user.Proof,
		user.Nickname(),
		user.Info,
		user.RegDate,
		user.Sig.Encode(),
	)
	if err != nil {
		log.Errorf("Can't execute 'putUser' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *Database) GetUser(eid *entity.ID) (*entity.User, error) {
	log.Debugf("Fetching user with id '%s' from the database.", eid.String())

	var nickname string
	var info string
	var proof crypto.ProofOfWork
	var regdate time.Time
	var encodedSig []byte
	var encodedKey []byte
	query := `
	SELECT Public_key,
	       Proof,
	       Nickname,
	       Info,
	       Timestamp,
	       Signature
	FROM User WHERE Id=?
	`

	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(
		&encodedKey,
		&proof,
		&nickname,
		&info,
		&regdate,
		&encodedSig)
	switch {
	case err == sql.ErrNoRows:
		log.Warning("No user with that ID.")
		return nil, errors.NoSuchEntity
	case err != nil:
		log.Errorf("Error fetching user from the database: %v", err)
		return nil, errors.DBOperFailed
	default:
		log.Debug("The user found successfully")
	}

	sig, err := crypto.ParseSignature(encodedSig)
	if err != nil {
		log.Errorf("Can't parse signature fetched from DB: %v", err)
		return nil, errors.Parsing
	}
	pubkey, err := crypto.ParsePublicKeyFromDER(encodedKey)
	if err != nil {
		log.Errorf("Can't parse public key fetched from DB: %v", err)
		return nil, errors.Parsing
	}

	u := entity.NewUser(nickname, info, pubkey, proof, regdate, sig)
	return u, nil
}
