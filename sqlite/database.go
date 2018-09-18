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
	"vminko.org/dscuss/subs"
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
		"  Id              BLOB PRIMARY KEY ON CONFLICT IGNORE," +
		"  Public_key      BLOB NOT NULL," +
		"  Proof           UNSIGNED BIG INT NOT NULL," +
		"  Nickname        TEXT NOT NULL," +
		"  Info            TEXT," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Signature       BLOB NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS  Message (" +
		"  Id              BLOB PRIMARY KEY ON CONFLICT IGNORE," +
		"  Subject         TEXT," +
		"  Content         TEXT," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Author_id       BLOB NOT NULL," +
		"  Parent_id       BLOB NOT NULL," +
		"  Signature       BLOB NOT NULL," +
		"  FOREIGN KEY (Author_id) REFERENCES User(Id))")
	exec("CREATE TABLE IF NOT EXISTS  Operation (" +
		"  Id              BLOB PRIMARY KEY ON CONFLICT IGNORE," +
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
		user.Info(),
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

func (d *Database) PutMessage(msg *entity.Message) error {
	log.Debugf("Adding message `%s' to the database.", msg.ShortID())

	query := `
	INSERT INTO Message
	( Id,
	  Subject,
	  Content,
	  Timestamp,
	  Author_id,
	  Parent_id,
	  Signature )
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	db := (*sql.DB)(d)
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatalf("Error preparing 'putMessage' statement: %v", err)
	}

	_, err = stmt.Exec(
		msg.ID()[:],
		msg.Subject,
		msg.Text,
		msg.DateWritten,
		msg.AuthorID[:],
		msg.ParentID[:],
		msg.Sig.Encode(),
	)
	if err != nil {
		log.Errorf("Can't execute 'putMessage' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *Database) GetMessage(eid *entity.ID) (*entity.Message, error) {
	log.Debugf("Fetching message with id '%s' from the database.", eid.String())

	var subj string
	var text string
	var wrdate time.Time
	var encodedSig []byte
	var authID entity.ID
	var parID entity.ID
	query := `
	SELECT Subject,
	       Content,
	       Timestamp,
	       Author_id,
	       Parent_id,
	       Signature
	FROM Message WHERE Id=?
	`

	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(
		&subj,
		&text,
		&wrdate,
		&authID,
		&parID,
		&encodedSig)
	switch {
	case err == sql.ErrNoRows:
		log.Warning("No message with that ID.")
		return nil, errors.NoSuchEntity
	case err != nil:
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	default:
		log.Debug("The message found successfully")
	}
	sig, err := crypto.ParseSignature(encodedSig)
	if err != nil {
		log.Errorf("Can't parse signature fetched from DB: %v", err)
		return nil, errors.Parsing
	}
	topic, err := d.getMessageTopic(eid)
	if err != nil {
		log.Fatalf("Error fetching topic of message '%s': %v. DB is corrupted",
			eid.String(), err)
	}
	u := entity.NewMessage(eid, subj, text, &authID, &parID, wrdate, sig, topic)
	return u, nil
}

func (d *Database) GetRootMessages(offset, limit int) ([]*entity.Message, error) {
	log.Debugf("Fetching root messages from the database.")

	var res []*entity.Message

	query := `
	SELECT Id,
	       Subject,
	       Content,
	       Timestamp,
	       Author_id,
	       Parent_id,
	       Signature
	FROM Message WHERE Parent_id=?
	ORDER BY Timestamp DESC
	LIMIT ? OFFSET ?
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, entity.ZeroID[:], limit, offset)
	if err != nil {
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	} else {
		log.Debug("The message found successfully")
	}
	defer rows.Close()

	for rows.Next() {
		var rawID []byte
		var subj string
		var text string
		var wrdate time.Time
		var encodedSig []byte
		var rawAuthID []byte
		var rawParID []byte

		err = rows.Scan(
			&rawID,
			&subj,
			&text,
			&wrdate,
			&rawAuthID,
			&rawParID,
			&encodedSig)
		if err != nil {
			log.Errorf("Error scanning message row: %v", err)
			return nil, errors.DBOperFailed
		}

		sig, err := crypto.ParseSignature(encodedSig)
		if err != nil {
			log.Errorf("Can't parse signature fetched from DB: %v", err)
			return nil, errors.Parsing
		}

		var id, authID, parID entity.ID
		parsOK := id.ParseSlice(rawID) == nil &&
			authID.ParseSlice(rawAuthID) == nil &&
			parID.ParseSlice(rawParID) == nil
		if !parsOK {
			log.Error("Can't parse ID fetched from DB")
			return nil, errors.Parsing
		}

		topic, err := d.getMessageTopic(&id)
		if err != nil {
			log.Fatalf("Error fetching topic of message '%s': %v. DB is corrupted",
				id.String(), err)
		}
		m := entity.NewMessage(&id, subj, text, &authID, &parID, wrdate, sig, topic)
		res = append(res, m)
	}

	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting next message row: %v", err)
		return nil, errors.DBOperFailed
	}

	return res, nil
}

func (d *Database) HasMessage(eid *entity.ID) (bool, error) {
	log.Debugf("Fetching message with id '%s' from the database.", eid.String())

	// FIXME: This is definitely not the most efficient implementation.
	var subj string
	query := `SELECT Subject FROM Message WHERE Id=?`
	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(&subj)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		log.Errorf("Error fetching message from the database: %v", err)
		return false, errors.DBOperFailed
	default:
		return true, nil
	}
}

func (d *Database) GetEntity(eid *entity.ID) (entity.Entity, error) {
	log.Debugf("Fetching entity with id '%s' from the database.", eid.String())

	m, err := d.GetMessage(eid)
	if err == errors.NoSuchEntity {
		u, err := d.GetUser(eid)
		if err != nil {
			return nil, err
		} else {
			return (entity.Entity)(u), nil
		}
	} else {
		return nil, err
	}
	return (entity.Entity)(m), nil
}

func (d *Database) putTag(tag string) error {
	log.Debugf("Adding tag `%s' to the database.", tag)

	query := "INSERT INTO Tag (Name) VALUES (?)"
	db := (*sql.DB)(d)
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatalf("Error preparing 'putTag' statement: %v", err)
	}

	_, err = stmt.Exec(tag)
	if err != nil {
		log.Errorf("Can't execute 'putTag' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *Database) putMessageTag(tag string, id *entity.ID) error {
	log.Debugf("Adding tag '%s' for the message '%s' to the database.", tag, id.Shorten())

	query := `
	INSERT INTO Message_Tag
        ( Message_id, Tag_id )
        VALUES (?, (SELECT Id FROM Tag WHERE Name=?))
	`
	db := (*sql.DB)(d)
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatalf("Error preparing 'putMessageTag' statement: %v", err)
	}

	_, err = stmt.Exec(tag, id[:])
	if err != nil {
		log.Errorf("Can't execute 'putMessageTag' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *Database) putMesageTopic(m *entity.Message) error {
	t := m.Topic
	if t == nil {
		log.Fatalf("BUG: messages with empty topics are prohibited")
	}
	for _, tag := range t {
		if d.putTag(tag) != nil {
			log.Errorf("Failed to store tag %s in the DB.", tag)
			return errors.DBOperFailed
		}
		if d.putMessageTag(tag, m.ID()) != nil {
			log.Errorf("Failed to store tag %s for message %s in the DB.",
				tag, m.ID().Shorten())
			return errors.DBOperFailed
		}
	}
	return nil
}

func (d *Database) getMessageTopic(eid *entity.ID) (subs.Topic, error) {
	log.Debugf("Fetching topic of the message '%s' from the database.", eid.Shorten())

	query := `
	SELECT Name
	FROM Tag
	INNER JOIN Message_Tag
	ON Tag.Id = Message_Tag.Tag_id AND Message_Tag.Message_id = ?
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, eid[:])
	if err != nil {
		log.Errorf("Error fetching topic from the database: %v", err)
		return nil, errors.DBOperFailed
	} else {
		log.Debug("The topic found successfully")
	}
	defer rows.Close()

	var topic subs.Topic
	for rows.Next() {
		var tag string
		err = rows.Scan(tag)
		if err != nil {
			log.Errorf("Error scanning message-tag row: %v", err)
			return nil, errors.DBOperFailed
		}
		topic, err = topic.AddTag(tag)
		if err != nil {
			log.Fatalf("Invalid tag '%s' occurred. DB is corrupted.", tag)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting next message-tag row: %v", err)
		return nil, errors.DBOperFailed
	}
	return topic, nil
}
