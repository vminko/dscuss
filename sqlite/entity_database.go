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
	"vminko.org/dscuss/thread"
)

// EntityDatabase stores Entities.
type EntityDatabase sql.DB

func OpenEntityDatabase(fileName string) (*EntityDatabase, error) {
	db, err := sql.Open("sqlite3", fileName+"?_mutex=no&_timeout=60")
	if err != nil {
		log.Errorf("Unable to open SQLite connection: %s", err.Error())
		return nil, errors.CantOpenDB
	}
	db.SetMaxOpenConns(1)

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

	return (*EntityDatabase)(db), nil
}

func (d *EntityDatabase) Close() error {
	db := (*sql.DB)(d)
	err := db.Close()
	if err != nil {
		log.Errorf("Unable to close the database: %v", err)
		return errors.DBOperFailed
	}
	return nil
}

func (d *EntityDatabase) PutUser(user *entity.User) error {
	log.Debugf("Adding user `%s' to the database", user.Nickname)

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
		user.Nickname,
		user.Info,
		user.RegDate,
		user.Sig.Encode(),
	)
	if err != nil {
		log.Errorf("Can't execute 'PutUser' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *EntityDatabase) GetUser(eid *entity.ID) (*entity.User, error) {
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
		log.Debug("No user with that ID.")
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

func (d *EntityDatabase) HasUser(eid *entity.ID) (bool, error) {
	log.Debugf("Checking whether DB contains user with id '%s'", eid.String())

	// FIXME: This is definitely not the most efficient implementation.
	var nick string
	query := `SELECT Nickname FROM User WHERE Id=?`
	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(&nick)
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

func (d *EntityDatabase) PutMessage(msg *entity.Message) error {
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
		log.Errorf("Can't execute 'PutMessage' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	if d.putMesageTopic(msg) != nil {
		log.Errorf("The DB is corrupted. Message %s is saved without topic", msg.Desc())
		return errors.DBOperFailed
	}
	return nil
}

func (d *EntityDatabase) GetMessage(eid *entity.ID) (*entity.Message, error) {
	log.Debugf("Fetching message with id '%s' from the database", eid.String())

	var subj string
	var text string
	var wrdate time.Time
	var encodedSig []byte
	var rawAuthID []byte
	var rawParID []byte
	var topicStr sql.NullString
	query := `
	SELECT Message.Subject,
	       Message.Content,
	       Message.Timestamp,
	       Message.Author_id,
	       Message.Parent_id,
	       Message.Signature,
	       GROUP_CONCAT(Tag.Name)
	FROM Message
	LEFT JOIN Message_Tag on Message.Id=Message_Tag.Message_Id
	LEFT JOIN Tag on Tag.Id=Message_Tag.Tag_Id
	WHERE Message.Id=?
	GROUP BY Message.Id
	`

	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(
		&subj,
		&text,
		&wrdate,
		&rawAuthID,
		&rawParID,
		&encodedSig,
		&topicStr)
	switch {
	case err == sql.ErrNoRows:
		log.Debug("No message with that ID.")
		return nil, errors.NoSuchEntity
	case err != nil:
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	default:
		log.Debug("The message found successfully")
	}

	var authID, parID entity.ID
	parsOK := authID.ParseSlice(rawAuthID) == nil && parID.ParseSlice(rawParID) == nil
	if !parsOK {
		log.Error("Can't parse ID fetched from DB")
		return nil, errors.Parsing
	}
	sig, err := crypto.ParseSignature(encodedSig)
	if err != nil {
		log.Errorf("Can't parse signature fetched from DB: %v", err)
		return nil, errors.Parsing
	}
	var topic subs.Topic
	if topicStr.Valid {
		topic, err = subs.NewTopic(topicStr.String)
		if err != nil {
			log.Errorf("The topic '%s' fetched from DB is invalid", topicStr.String)
			return nil, errors.InconsistentDB
		}
	}
	m, err := entity.NewMessage(subj, text, &authID, &parID, wrdate, sig, topic)
	if err != nil {
		log.Errorf("The message '%s' fetched from DB is invalid", m.Desc())
		return nil, errors.InconsistentDB
	}
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
			log.Error("Can't parse an ID fetched from DB")
			return nil, errors.Parsing
		}
		log.Debugf("Found message id %s", id.String())

		topic, err := subs.NewTopic(topicStr)
		if err != nil {
			log.Errorf("The topic '%s' fetched from DB is invalid", topicStr)
			return nil, errors.InconsistentDB
		}
		m, err := entity.NewMessage(subj, text, &authID, &parID, wrdate, sig, topic)
		if err != nil {
			log.Errorf("The message '%s' fetched from DB is invalid", m.Desc())
			return nil, errors.InconsistentDB
		}
		res = append(res, m)
	}
	err := rows.Err()
	if err != nil {
		log.Errorf("Error getting next message row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func (d *EntityDatabase) GetRootMessages(offset, limit int) ([]*entity.Message, error) {
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
	return res, nil
}

func (d *EntityDatabase) GetTopicMessages(topic subs.Topic, offset, limit int) ([]*entity.Message, error) {
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
	return res, nil
}

func (d *EntityDatabase) GetReplies(eid *entity.ID) ([]*entity.Message, error) {
	log.Debugf("Fetching replies for '%s' from the database", eid.Shorten())
	query := `
	SELECT Message.Id,
	       Message.Subject,
	       Message.Content,
	       Message.Timestamp,
	       Message.Author_id,
	       Message.Parent_id,
	       Message.Signature,
	       ''
	FROM Message
	WHERE Message.Parent_id=?
	ORDER BY Message.Timestamp DESC
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, eid[:])
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
	return res, nil
}

func (d *EntityDatabase) fillSubthreads(t *thread.Node) error {
	eid := t.Msg.ID()
	replies, err := d.GetReplies(eid)
	if err != nil {
		log.Errorf("Error fetching replies for %s: %v", eid.Shorten())
		return err
	}
	for _, r := range replies {
		node := t.AddReply(r)
		err = d.fillSubthreads(node)
		if err != nil {
			log.Errorf("Error fetching replies for %s: %v", eid.Shorten())
			return err
		}
	}
	return nil
}

func (d *EntityDatabase) GetThread(eid *entity.ID) (*thread.Node, error) {
	log.Debugf("Fetching thread %s from the database", eid.Shorten())
	root, err := d.GetMessage(eid)
	if err != nil {
		log.Errorf("Error fetching root if the thread %s: %v", eid.Shorten())
		return nil, err
	}
	t := thread.New(root)
	err = d.fillSubthreads(t)
	if err != nil {
		log.Errorf("Error fetching replies for %s: %v", eid.Shorten())
		return nil, err
	}

	return t, nil
}

func (d *EntityDatabase) HasMessage(eid *entity.ID) (bool, error) {
	log.Debugf("Checking whether DB contains message with id '%s'", eid.String())

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

func (d *EntityDatabase) putTag(tag string) error {
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

func (d *EntityDatabase) putMessageTag(tag string, id *entity.ID) error {
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

func (d *EntityDatabase) putMesageTopic(m *entity.Message) error {
	t := m.Topic
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

func (d *EntityDatabase) putMessageOperation(operID, msgID *entity.ID) error {
	log.Debugf("Adding association between operation '%s' and message '%s' to the database",
		operID.Shorten(), msgID.Shorten())
	query := `
	INSERT INTO Operation_on_Message
        ( Operation_id, Message_id )
        VALUES (?, ?)
	`
	db := (*sql.DB)(d)
	_, err := db.Exec(query, operID[:], msgID[:])
	if err != nil {
		log.Errorf("Can't execute 'putMessageOperation' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	return nil
}

func (d *EntityDatabase) putUserOperation(operID, userID *entity.ID) error {
	log.Debugf("Adding association between operation '%s' and user '%s' to the database",
		operID.Shorten(), userID.Shorten())
	query := `
	INSERT INTO Operation_on_User
        ( Operation_id, User_id )
        VALUES (?, ?)
	`
	db := (*sql.DB)(d)
	_, err := db.Exec(query, operID[:], userID[:])
	if err != nil {
		log.Errorf("Can't execute 'putUserOperation' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	return nil
}

func (d *EntityDatabase) putOperationObject(o *entity.Operation) error {
	var hasFunc func(*entity.ID) (bool, error)
	var putFunc func(op, obj *entity.ID) error
	switch o.OperationType() {
	case entity.OperationTypeRemoveMessage:
		hasFunc = d.HasMessage
		putFunc = d.putMessageOperation
	case entity.OperationTypeBanUser:
		hasFunc = d.HasUser
		putFunc = d.putUserOperation
	default:
		log.Fatalf("BUG: unexpected operation type %d.", o.Type)
	}

	has, err := hasFunc(&o.ObjectID)
	if err != nil {
		log.Errorf("Failed to check if the DB contains %s: %v", o.ID().Shorten(), err)
		return err
	}
	if !has {
		log.Errorf("Attempt to store operation %s on non-stored object %s",
			o.ID().Shorten(), o.ObjectID.Shorten())
		return errors.NoSuchEntity
	}
	if putFunc(o.ID(), &o.ObjectID) != nil {
		log.Errorf("Failed to store association between operation %s and object %s",
			o.ID().Shorten(), o.ObjectID.Shorten())
		return errors.DBOperFailed
	}
	return nil
}

func (d *EntityDatabase) PutOperation(oper *entity.Operation) error {
	log.Debugf("Adding operation '%s' to the database", oper.ShortID())
	query := `
	INSERT INTO Operation
	( Id,
	  Type,
	  Reason,
	  Comment,
	  Author_id,
	  Timestamp,
	  Signature )
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	db := (*sql.DB)(d)
	_, err := db.Exec(
		query,
		oper.ID()[:],
		oper.OperationType(),
		oper.Reason,
		oper.Comment,
		oper.AuthorID[:],
		oper.DatePerformed,
		oper.Sig.Encode(),
	)
	if err != nil {
		log.Errorf("Can't execute 'PutOperation' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	err = d.putOperationObject(oper)
	if err == errors.NoSuchEntity {
		return err
	} else if err != nil {
		log.Errorf("The DB is corrupted. Operation %s is saved,"+
			" but association with object %s is not",
			oper.Desc(), oper.ObjectID.Shorten())
		return errors.DBOperFailed
	}
	return nil
}

func scanOperationRows(rows *sql.Rows, objID *entity.ID) ([]*entity.Operation, error) {
	var res []*entity.Operation
	for rows.Next() {
		var rawID []byte
		var typ int
		var reason int
		var comment string
		var rawAuthID []byte
		var perfDate time.Time
		var encodedSig []byte
		err := rows.Scan(
			&rawID,
			&typ,
			&reason,
			&comment,
			&rawAuthID,
			&perfDate,
			&encodedSig)
		if err != nil {
			log.Errorf("Error scanning opertion row: %v", err)
			return nil, errors.DBOperFailed
		}

		sig, err := crypto.ParseSignature(encodedSig)
		if err != nil {
			log.Errorf("Can't parse signature fetched from DB: %v", err)
			return nil, errors.Parsing
		}

		var id, authID entity.ID
		parsOK := id.ParseSlice(rawID) == nil && authID.ParseSlice(rawAuthID) == nil
		if !parsOK {
			log.Error("Can't parse an ID fetched from DB")
			return nil, errors.Parsing
		}
		log.Debugf("Found operation id %s", id.String())

		o, err := entity.NewOperation(
			(entity.OperationType)(typ),
			(entity.OperationReason)(reason),
			comment,
			&authID,
			objID,
			perfDate,
			sig)
		if err != nil {
			log.Errorf("The message '%s' fetched from DB is invalid", o.Desc())
			return nil, errors.InconsistentDB
		}
		res = append(res, o)
	}
	err := rows.Err()
	if err != nil {
		log.Errorf("Error getting next operation row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func (d *EntityDatabase) GetOperationsOnUser(uid *entity.ID) ([]*entity.Operation, error) {
	log.Debugf("Fetching operations on user %s from the database", uid.Shorten())
	query := `
	SELECT Operation.Id,
	       Operation.Type,
	       Operation.Reason,
	       Operation.Comment,
	       Operation.Author_id,
	       Operation.Timestamp,
	       Operation.Signature
	FROM Operation
	INNER JOIN Operation_on_User on Operation.Id=Operation_on_User.Operation_Id
	WHERE Operation_on_User.User_id=?
	GROUP BY Operation.Id
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, uid[:])
	if err != nil {
		log.Errorf("Error fetching operations from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	res, err := scanOperationRows(rows, uid)
	if err != nil {
		log.Errorf("Error scanning operation rows: %v", err)
		return nil, err
	}
	return res, nil
}

func (d *EntityDatabase) GetOperationsOnMessage(mid *entity.ID) ([]*entity.Operation, error) {
	log.Debugf("Fetching operations on message %s from the database", mid.Shorten())
	query := `
	SELECT Operation.Id,
	       Operation.Type,
	       Operation.Reason,
	       Operation.Comment,
	       Operation.Author_id,
	       Operation.Timestamp,
	       Operation.Signature
	FROM Operation
	INNER JOIN Operation_on_Message on Operation.Id=Operation_on_Message.Operation_Id
	WHERE Operation_on_Message.Message_id=?
	GROUP BY Operation.Id
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, mid[:])
	if err != nil {
		log.Errorf("Error fetching operations from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	res, err := scanOperationRows(rows, mid)
	if err != nil {
		log.Errorf("Error scanning operation rows: %v", err)
		return nil, err
	}
	return res, nil
}

func (d *EntityDatabase) GetOperation(oid *entity.ID) (*entity.Operation, error) {
	log.Debugf("Fetching operation with id '%s' from the database", oid.String())

	var typ int
	var reason int
	var comment string
	var rawAuthID []byte
	var perfDate time.Time
	var encodedSig []byte
	var rawMsgID []byte
	var rawUserID []byte
	query := `
	SELECT Operation.Type,
	       Operation.Reason,
	       Operation.Comment,
	       Operation.Author_id,
	       Operation.Timestamp,
	       Operation.Signature
	       Operation_on_Message.Message_Id
	       Operation_on_User.User_Id
	FROM Operation
	LEFT JOIN Operation_on_Message on Operation.Id=Operation_on_Message.Operation_Id
	LEFT JOIN Operation_on_User on Operation.Id=Operation_on_User.Operation_Id
	WHERE Operation.Id=?
	GROUP BY Operation.Id
	`
	db := (*sql.DB)(d)
	err := db.QueryRow(query, oid[:]).Scan(
		&typ,
		&reason,
		&comment,
		&rawAuthID,
		&perfDate,
		&encodedSig,
		&rawMsgID,
		&rawUserID)
	switch {
	case err == sql.ErrNoRows:
		log.Debug("No message with that ID.")
		return nil, errors.NoSuchEntity
	case err != nil:
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	default:
		log.Debug("The message found successfully")
	}

	var authID, objID entity.ID
	var rawObjID []byte
	switch {
	case len(rawMsgID) > 0:
		rawObjID = rawMsgID
	case err != nil:
		rawObjID = rawUserID
	default:
		log.Errorf("Failed to fetch object ID of the operation %s", oid.Shorten())
		return nil, errors.InconsistentDB
	}

	parsOK := authID.ParseSlice(rawAuthID) == nil && objID.ParseSlice(rawObjID) == nil
	if !parsOK {
		log.Error("Can't parse ID fetched from DB")
		return nil, errors.Parsing
	}
	sig, err := crypto.ParseSignature(encodedSig)
	if err != nil {
		log.Errorf("Can't parse signature fetched from DB: %v", err)
		return nil, errors.Parsing
	}
	o, err := entity.NewOperation(
		(entity.OperationType)(typ),
		(entity.OperationReason)(reason),
		comment,
		&authID,
		&objID,
		perfDate,
		sig)
	if err != nil {
		log.Errorf("The operation '%s' fetched from DB is invalid", o.Desc())
		return nil, errors.InconsistentDB
	}
	return o, nil
}

func (d *EntityDatabase) HasOperation(eid *entity.ID) (bool, error) {
	log.Debugf("Checking whether DB contains operation with id '%s'", eid.String())

	// FIXME: This is definitely not the most efficient implementation.
	var typ int
	query := `SELECT Type FROM Operation WHERE Id=?`
	db := (*sql.DB)(d)
	err := db.QueryRow(query, eid[:]).Scan(&typ)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		log.Errorf("Error fetching operation from the database: %v", err)
		return false, errors.DBOperFailed
	default:
		return true, nil
	}
}
