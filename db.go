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

package dscuss

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// globalDB stores global network data.
type globalDB sql.DB

func open(fileName string) (*globalDB, error) {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		Logf(ERROR, "Unable to open SQLite connection: %s", err.Error())
		return nil, ErrDatabase
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
		"  Timestamp       INTEGER NOT NULL," +
		"  Signature       BLOB NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS  Message (" +
		"  Id              BLOB PRIMARY KEY," +
		"  Subject         TEXT," +
		"  Content         TEXT," +
		"  Timestamp       UNSIGNED BIG INT NOT NULL," +
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
		"  Timestamp       UNSIGNED BIG INT NOT NULL," +
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
		Logf(ERROR, "Unable to initialize the database: %s", execErr.Error())
		return nil, ErrDatabase
	}

	return (*globalDB)(db), nil
}

func (gdb *globalDB) close(fileName string) error {
	db := (*sql.DB)(gdb)
	return db.Close()
}

func (gdb *globalDB) putUser(user *User) error {
	Logf(DEBUG, "Adding user `%s' to the database.", user.Nickname)

	db := (*sql.DB)(gdb)
	stmt, err := db.Prepare("INSERT INTO User " +
		"( Id," +
		"  Public_key," +
		"  Proof," +
		"  Nickname," +
		"  Info," +
		"  Timestamp," +
		"  Signature ) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic("error preparing 'putUser' statement: %s" + err.Error())
	}

	pkpem, err := user.PubKey.encode()
	if err != nil {
		Logf(ERROR, "Can't encode %s's public key: %s", user.Nickname, err.Error())
		return ErrInternal
	}

	_, err = stmt.Exec(
		user.ID[:],
		pkpem,
		user.Proof,
		user.Nickname,
		user.Info,
		user.RegDate,
		user.Sig.encode(),
	)
	if err != nil {
		Logf(ERROR, "Can't execute 'putUser' statement: %s", err.Error())
		return ErrDatabase
	}

	return nil
}
