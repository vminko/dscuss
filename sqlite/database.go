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
	"fmt"
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
	db, err := sql.Open("sqlite3", fileName+"?_mutex=no")
	if err != nil {
		log.Errorf("Unable to open SQLite connection: %s", err.Error())
		return nil, errors.CantOpenDB
	}
	db.SetMaxOpenConns(2)

	var execErr error
	exec := func(req string) {
		if execErr != nil {
			return
		}
		_, execErr = db.Exec(req)
	}
	exec("PRAGMA temp_store=MEMORY")
	exec("PRAGMA synchronous=FULL")
	exec("PRAGMA locking_mode=EXCLUSIVE")
	exec("PRAGMA page_size=4092")
	//exec("PRAGMA journal_mode=WAL")
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
	log.Debugf("Adding user `%s' to the database", user.Nickname())

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
	pkpem := user.PubKey.EncodeToDER()
	_, err := db.Exec(
		query,
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
	log.Debugf("Fetching user with id '%s' from the database", eid.String())

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
	log.Debugf("Adding message `%s' to the database", msg.ShortID())
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
	_, err := db.Exec(
		query,
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
	if d.putMesageTopic(msg) != nil {
		log.Errorf("The DB is corrupted. Message %s is saved without topic", msg.Desc())
		return errors.DBOperFailed
	}
	return nil
}

func (d *Database) GetMessage(eid *entity.ID) (*entity.Message, error) {
	log.Debugf("Fetching message with id '%s' from the database", eid.String())

	var subj string
	var text string
	var wrdate time.Time
	var encodedSig []byte
	var authID entity.ID
	var parID entity.ID
	var topicStr string
	query := `
	SELECT Message.Subject,
	       Message.Content,
	       Message.Timestamp,
	       Message.Author_id,
	       Message.Parent_id,
	       Message.Signature,
	       GROUP_CONCAT(Tag.Name)
	FROM Message
	INNER JOIN Message_Tag on Message.Id=Message_Tag.Message_Id
	INNER JOIN Tag on Tag.Id=Message_Tag.Tag_Id
	WHERE Message.Id=?
	GROUP BY Message.Id
	`

	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(
		&subj,
		&text,
		&wrdate,
		&authID,
		&parID,
		&encodedSig,
		&topicStr)
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
	topic, err := subs.NewTopic(topicStr)
	if err != nil {
		log.Fatalf("The topic '%s' fetched from DB is invalid", topicStr)
	}
	m := entity.NewMessage(eid, subj, text, &authID, &parID, wrdate, sig, topic)
	return m, nil
}

func scanMessageRows(rows *sql.Rows) ([]*entity.Message, error) {
	var res []*entity.Message
	for rows.Next() {
		var rawID []byte
		var subj string
		var text string
		var wrdate time.Time
		var encodedSig []byte
		var rawAuthID []byte
		var rawParID []byte
		var topicStr string
		err := rows.Scan(
			&rawID,
			&subj,
			&text,
			&wrdate,
			&rawAuthID,
			&rawParID,
			&encodedSig,
			&topicStr)
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

		topic, err := subs.NewTopic(topicStr)
		if err != nil {
			log.Fatalf("The topic '%s' fetched from DB is invalid", topicStr)
		}
		m := entity.NewMessage(&id, subj, text, &authID, &parID, wrdate, sig, topic)
		res = append(res, m)
	}
	return res, nil
}

func (d *Database) GetRootMessages(offset, limit int) ([]*entity.Message, error) {
	log.Debugf("Fetching root messages from the database")
	query := `
	SELECT Message.Id,
	       Message.Subject,
	       Message.Content,
	       Message.Timestamp,
	       Message.Author_id,
	       Message.Parent_id,
	       Message.Signature,
	       GROUP_CONCAT(Tag.Name)
	FROM Message
	INNER JOIN Message_Tag on Message.Id=Message_Tag.Message_Id
	INNER JOIN Tag on Tag.Id=Message_Tag.Tag_Id
	WHERE Message.Parent_id=?
	GROUP BY Message.Id
	ORDER BY Message.Timestamp DESC
	LIMIT ? OFFSET ?
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, entity.ZeroID[:], limit, offset)
	if err != nil {
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	res, err := scanMessageRows(rows)
	if err != nil {
		log.Errorf("Error scanning message rows: %v", err)
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting next message row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func (d *Database) GetTopicMessages(topic subs.Topic, offset, limit int) ([]*entity.Message, error) {
	log.Debugf("Fetching topic messages from the database")
	query := `
	SELECT Message.Id,
	       Message.Subject,
	       Message.Content,
	       Message.Timestamp,
	       Message.Author_id,
	       Message.Parent_id,
	       Message.Signature,
	       GROUP_CONCAT(Tag.Name)
	FROM Message
	INNER JOIN Message_Tag on Message.Id=Message_Tag.Message_Id
	INNER JOIN Tag on Tag.Id=Message_Tag.Tag_Id
	WHERE Message.Id IN (
		SELECT Message_Tag.Message_Id
		FROM Message_Tag
		JOIN Tag on Message_Tag.Tag_Id = Tag.Id
		WHERE Tag.Name IN (%s)
		GROUP BY Message_Tag.Message_Id
		HAVING COUNT(DISTINCT Tag.Name) = %d
	)
	GROUP BY Message.Id
	ORDER BY Message.Timestamp DESC
	LIMIT %d OFFSET %d
	`
	var params []interface{}
	inCondition := ""
	for _, t := range topic {
		params = append(params, t)
		if inCondition != "" {
			inCondition += ", "
		}
		inCondition += "?"
	}
	db := (*sql.DB)(d)
	query = fmt.Sprintf(query, inCondition, len(topic), limit, offset)
	rows, err := db.Query(query, params...)
	if err != nil {
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	res, err := scanMessageRows(rows)
	if err != nil {
		log.Errorf("Error scanning message rows: %v", err)
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting next message row: %v", err)
		return nil, errors.DBOperFailed
	}

	return res, nil
}

func (d *Database) HasMessage(eid *entity.ID) (bool, error) {
	log.Debugf("Fetching message with id '%s' from the database", eid.String())

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
	log.Debugf("Fetching entity with id '%s' from the database", eid.String())

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
	log.Debugf("Adding tag `%s' to the database", tag)

	query := "INSERT INTO Tag (Name) VALUES (?)"
	db := (*sql.DB)(d)
	_, err := db.Exec(query, tag)
	if err != nil {
		log.Errorf("Can't execute 'putTag' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *Database) putMessageTag(tag string, id *entity.ID) error {
	log.Debugf("Adding tag '%s' for the message '%s' to the database", tag, id.Shorten())
	query := `
	INSERT INTO Message_Tag
        ( Message_id, Tag_id )
        VALUES (?, (SELECT Id FROM Tag WHERE Name=?))
	`
	db := (*sql.DB)(d)
	_, err := db.Exec(query, id[:], tag)
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
			log.Errorf("Failed to store tag %s in the DB", tag)
			return errors.DBOperFailed
		}
		if d.putMessageTag(tag, m.ID()) != nil {
			log.Errorf("Failed to store tag %s for message %s in the DB",
				tag, m.ID().Shorten())
			return errors.DBOperFailed
		}
	}
	return nil
}
