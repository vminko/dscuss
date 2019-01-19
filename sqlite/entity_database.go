/*
This file is part of Dscuss.
Copyright (C) 2017-2019  Vitaly Minko

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
	exec("CREATE TABLE IF NOT EXISTS  Users (" +
		"  Id              BLOB PRIMARY KEY ON CONFLICT IGNORE," +
		"  Public_key      BLOB NOT NULL," +
		"  Proof           UNSIGNED BIG INT NOT NULL," +
		"  Nickname        TEXT NOT NULL," +
		"  Info            TEXT," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Signature       BLOB NOT NULL," +
		"  TimeStored      TIMESTAMP NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS  Messages (" +
		"  Id              BLOB PRIMARY KEY ON CONFLICT IGNORE," +
		"  Subject         TEXT," +
		"  Content         TEXT," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Author_id       BLOB NOT NULL REFERENCES Users," +
		"  Parent_id       BLOB NOT NULL," +
		"  Signature       BLOB NOT NULL," +
		"  TimeStored      TIMESTAMP NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS  Operations (" +
		"  Id              BLOB PRIMARY KEY ON CONFLICT IGNORE," +
		"  Type            INTEGER NOT NULL," +
		"  Reason          INTEGER NOT NULL," +
		"  Comment         TEXT," +
		"  Author_id       BLOB NOT NULL REFERENCES Users," +
		"  Timestamp       TIMESTAMP NOT NULL," +
		"  Signature       BLOB NOT NULL," +
		"  TimeStored      TIMESTAMP NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS  Operations_on_Users (" +
		"  Operation_id    BLOB NOT NULL REFERENCES Operations," +
		"  User_id         BLOB NOT NULL REFERENCES Users)")
	exec("CREATE TABLE IF NOT EXISTS  Operations_on_Messages (" +
		"  Operation_id    BLOB NOT NULL REFERENCES Operations," +
		"  Message_id      BLOB NOT NULL REFERENCES Messages)")
	exec("CREATE TABLE IF NOT EXISTS  Tags (" +
		"  Id              INTEGER PRIMARY KEY AUTOINCREMENT," +
		"  Name            TEXT NOT NULL UNIQUE ON CONFLICT IGNORE)")
	exec("CREATE TABLE IF NOT EXISTS  Message_Tags (" +
		"  Tag_id          INTEGER NOT NULL REFERENCES Tags," +
		"  Message_id      BLOB NOT NULL REFERENCES Messages," +
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

func (d *EntityDatabase) PutUser(user *entity.User, ts time.Time) error {
	log.Debugf("Adding user `%s' to the database", user.Nickname)

	query := `
	INSERT INTO Users
	( Id,
	  Public_key,
	  Proof,
	  Nickname,
	  Info,
	  Timestamp,
	  Signature,
	  TimeStored )
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		ts,
	)
	if err != nil {
		log.Errorf("Can't execute 'PutUser' statement: %s", err.Error())
		return errors.DBOperFailed
	}

	return nil
}

func (d *EntityDatabase) GetUser(eid *entity.ID) (*entity.User, error) {
	log.Debugf("Fetching user with id '%s' from the database", eid)

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
	FROM Users WHERE Id=?
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
	log.Debugf("Checking whether DB contains user with id '%s'", eid)

	// FIXME: This is definitely not the most efficient implementation.
	var nick string
	query := `SELECT Nickname FROM Users WHERE Id=?`
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

func (d *EntityDatabase) PutMessage(msg *entity.Message, ts time.Time) error {
	log.Debugf("Adding message `%s' to the database", msg.ShortID())
	query := `
	INSERT INTO Messages
	( Id,
	  Subject,
	  Content,
	  Timestamp,
	  Author_id,
	  Parent_id,
	  Signature,
	  TimeStored )
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		ts,
	)
	if err != nil {
		log.Errorf("Can't execute 'PutMessage' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	if d.putMesageTopic(msg) != nil {
		log.Errorf("The DB is corrupted. Message %s is saved without topic", msg)
		return errors.DBOperFailed
	}
	return nil
}

func (d *EntityDatabase) GetMessage(eid *entity.ID) (*entity.Message, error) {
	log.Debugf("Fetching message with id '%s' from the database", eid)

	var subj string
	var text string
	var wrdate time.Time
	var encodedSig []byte
	var rawAuthID []byte
	var rawParID []byte
	var topicStr sql.NullString
	query := `
	SELECT Messages.Subject,
	       Messages.Content,
	       Messages.Timestamp,
	       Messages.Author_id,
	       Messages.Parent_id,
	       Messages.Signature,
	       GROUP_CONCAT(Tags.Name)
	FROM Messages
	LEFT JOIN Message_Tags on Messages.Id=Message_Tags.Message_id
	LEFT JOIN Tags on Tags.Id=Message_Tags.Tag_id
	WHERE Messages.Id=?
	GROUP BY Messages.Id
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
		log.Error("Can't parse ID fetched from the entity DB")
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
		log.Errorf("The message '%s' fetched from DB is invalid", m)
		return nil, errors.InconsistentDB
	}
	return m, nil
}

func scanMessageRows(rows *sql.Rows) ([]*entity.Message, error) {
	var res []*entity.Message
	for rows.Next() {
		m, _, err := scanSingleMessageRow(rows, false)
		if err != nil {
			log.Errorf("Error scanning single message row: %v", err)
			return nil, err
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

func scanStoredMessageRows(rows *sql.Rows) ([]*entity.StoredMessage, error) {
	var res []*entity.StoredMessage
	for rows.Next() {
		m, t, err := scanSingleMessageRow(rows, true)
		if err != nil {
			log.Errorf("Error scanning single message row: %v", err)
			return nil, err
		}
		res = append(res, &entity.StoredMessage{m, t})
	}
	err := rows.Err()
	if err != nil {
		log.Errorf("Error getting next message row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func scanSingleMessageRow(rows *sql.Rows, scanTimeStored bool) (*entity.Message, time.Time, error) {
	var rawID []byte
	var subj string
	var text string
	var wrdate time.Time
	var encodedSig []byte
	var rawAuthID []byte
	var rawParID []byte
	var topicStr string
	var tmStored time.Time
	var err error
	if scanTimeStored {
		err = rows.Scan(
			&rawID,
			&subj,
			&text,
			&wrdate,
			&rawAuthID,
			&rawParID,
			&encodedSig,
			&topicStr,
			&tmStored)
	} else {
		err = rows.Scan(
			&rawID,
			&subj,
			&text,
			&wrdate,
			&rawAuthID,
			&rawParID,
			&encodedSig,
			&topicStr)
	}
	if err != nil {
		log.Errorf("Error scanning message row: %v", err)
		return nil, time.Time{}, errors.DBOperFailed
	}

	sig, err := crypto.ParseSignature(encodedSig)
	if err != nil {
		log.Errorf("Can't parse signature fetched from DB: %v", err)
		return nil, time.Time{}, errors.Parsing
	}

	var id, authID, parID entity.ID
	parsOK := id.ParseSlice(rawID) == nil &&
		authID.ParseSlice(rawAuthID) == nil &&
		parID.ParseSlice(rawParID) == nil
	if !parsOK {
		log.Error("Can't parse an ID fetched from DB")
		return nil, time.Time{}, errors.Parsing
	}
	log.Debugf("Found message id %s", &id)

	topic, err := subs.NewTopic(topicStr)
	if err != nil {
		log.Errorf("The topic '%s' fetched from DB is invalid", topicStr)
		return nil, time.Time{}, errors.InconsistentDB
	}
	m, err := entity.NewMessage(subj, text, &authID, &parID, wrdate, sig, topic)
	if err != nil {
		log.Errorf("The message '%s' fetched from DB is invalid", m)
		return nil, time.Time{}, errors.InconsistentDB
	}
	return m, tmStored, nil
}

func (d *EntityDatabase) GetRootMessages(offset, limit int) ([]*entity.Message, error) {
	log.Debugf("Fetching root messages from the database")
	query := `
	SELECT Messages.Id,
	       Messages.Subject,
	       Messages.Content,
	       Messages.Timestamp,
	       Messages.Author_id,
	       Messages.Parent_id,
	       Messages.Signature,
	       GROUP_CONCAT(Tags.Name)
	FROM Messages
	INNER JOIN Message_Tags on Messages.Id=Message_Tags.Message_id
	INNER JOIN Tags on Tags.Id=Message_Tags.Tag_id
	WHERE Messages.Parent_id=?
	GROUP BY Messages.Id
	ORDER BY Messages.Timestamp ASC
	LIMIT ? OFFSET ?
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, entity.ZeroID[:], limit, offset)
	if err != nil {
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanMessageRows(rows)
}

func (d *EntityDatabase) GetTopicMessages(topic subs.Topic, offset, limit int) ([]*entity.Message, error) {
	log.Debugf("Fetching topic messages from the database")
	query := `
	SELECT Messages.Id,
	       Messages.Subject,
	       Messages.Content,
	       Messages.Timestamp,
	       Messages.Author_id,
	       Messages.Parent_id,
	       Messages.Signature,
	       GROUP_CONCAT(Tags.Name)
	FROM Messages
	INNER JOIN Message_Tags on Messages.Id=Message_Tags.Message_id
	INNER JOIN Tags on Tags.Id=Message_Tags.Tag_id
	WHERE Messages.Id IN (
		SELECT Message_Tags.Message_id
		FROM Message_Tags
		JOIN Tags on Message_Tags.Tag_id = Tags.Id
		WHERE Tags.Name IN (%s)
		GROUP BY Message_Tags.Message_id
		HAVING COUNT(DISTINCT Tags.Name) = %d
	)
	GROUP BY Messages.Id
	ORDER BY Messages.Timestamp ASC
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
	return scanMessageRows(rows)
}

func (d *EntityDatabase) GetMessagesStoredAfter(ts time.Time, limit int) ([]*entity.StoredMessage, error) {
	log.Debugf("Fetching messages since %s from the database", ts.Format(time.RFC3339))
	query := `
	SELECT Messages.Id,
	       Messages.Subject,
	       Messages.Content,
	       Messages.Timestamp,
	       Messages.Author_id,
	       Messages.Parent_id,
	       Messages.Signature,
	       GROUP_CONCAT(Tags.Name),
	       Messages.TimeStored
	FROM Messages
	INNER JOIN Message_Tags on Messages.Id=Message_Tags.Message_id
	INNER JOIN Tags on Tags.Id=Message_Tags.Tag_id
	WHERE Messages.TimeStored>=?
	GROUP BY Messages.Id
	ORDER BY Messages.Timestamp ASC
	LIMIT ?
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, ts, limit)
	if err != nil {
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanStoredMessageRows(rows)
}

func (d *EntityDatabase) GetReplies(eid *entity.ID) ([]*entity.Message, error) {
	log.Debugf("Fetching replies for '%s' from the database", eid.Shorten())
	query := `
	SELECT Messages.Id,
	       Messages.Subject,
	       Messages.Content,
	       Messages.Timestamp,
	       Messages.Author_id,
	       Messages.Parent_id,
	       Messages.Signature,
	       ''
	FROM Messages
	WHERE Messages.Parent_id=?
	ORDER BY Messages.Timestamp ASC
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, eid[:])
	if err != nil {
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanMessageRows(rows)
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
	log.Debugf("Checking whether DB contains message with id '%s'", eid)

	// FIXME: This is definitely not the most efficient implementation.
	var subj string
	query := `SELECT Subject FROM Messages WHERE Id=?`
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

	query := "INSERT INTO Tags (Name) VALUES (?)"
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
	INSERT INTO Message_Tags
        ( Message_id, Tag_id )
        VALUES (?, (SELECT Id FROM Tags WHERE Name=?))
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
	INSERT INTO Operations_on_Messages
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
	INSERT INTO Operations_on_Users
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

func (d *EntityDatabase) PutOperation(oper *entity.Operation, ts time.Time) error {
	log.Debugf("Adding operation '%s' to the database", oper.ShortID())
	query := `
	INSERT INTO Operations
	( Id,
	  Type,
	  Reason,
	  Comment,
	  Author_id,
	  Timestamp,
	  Signature,
	  TimeStored )
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		ts,
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
			oper, oper.ObjectID.Shorten())
		return errors.DBOperFailed
	}
	return nil
}

func scanStoredOperationRows(rows *sql.Rows, objID *entity.ID) ([]*entity.StoredOperation, error) {
	var res []*entity.StoredOperation
	for rows.Next() {
		o, t, err := scanSingleOperationRow(rows, objID, true)
		if err != nil {
			log.Errorf("Error scanning single operation row: %v", err)
			return nil, err
		}
		res = append(res, &entity.StoredOperation{o, t})
	}
	err := rows.Err()
	if err != nil {
		log.Errorf("Error getting next operation row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func scanOperationRows(rows *sql.Rows, objID *entity.ID) ([]*entity.Operation, error) {
	var res []*entity.Operation
	for rows.Next() {
		o, _, err := scanSingleOperationRow(rows, objID, false)
		if err != nil {
			log.Errorf("Error scanning single operation row: %v", err)
			return nil, err
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

func scanSingleOperationRow(
	rows *sql.Rows,
	objID *entity.ID,
	scanTimeStored bool,
) (*entity.Operation, time.Time, error) {
	var rawID []byte
	var typ int
	var reason int
	var comment string
	var rawAuthID []byte
	var perfDate time.Time
	var encodedSig []byte
	var rawMsgID []byte
	var rawUserID []byte
	var tmStored time.Time
	var err error
	if objID == nil {
		if scanTimeStored {
			err = rows.Scan(
				&rawID, &typ, &reason, &comment, &rawAuthID, &perfDate, &encodedSig,
				&rawMsgID,
				&rawUserID,
				&tmStored)
		} else {
			err = rows.Scan(
				&rawID, &typ, &reason, &comment, &rawAuthID, &perfDate, &encodedSig,
				&rawMsgID,
				&rawUserID)
		}
	} else {
		if scanTimeStored {
			err = rows.Scan(
				&rawID, &typ, &reason, &comment, &rawAuthID, &perfDate, &encodedSig,
				&tmStored)
		} else {
			err = rows.Scan(
				&rawID, &typ, &reason, &comment, &rawAuthID, &perfDate, &encodedSig)
		}
	}
	if err != nil {
		log.Errorf("Error scanning operation row: %v", err)
		return nil, time.Time{}, errors.DBOperFailed
	}

	var id, authID entity.ID
	parsOK := id.ParseSlice(rawID) == nil && authID.ParseSlice(rawAuthID) == nil
	if !parsOK {
		log.Error("Can't parse an ID fetched from DB")
		return nil, time.Time{}, errors.Parsing
	}
	log.Debugf("Found operation id %s", &id)

	var oID entity.ID
	if objID == nil {
		var rawObjID []byte
		switch {
		case len(rawMsgID) > 0:
			rawObjID = rawMsgID
		case len(rawUserID) > 0:
			rawObjID = rawUserID
		default:
			log.Errorf("Failed to fetch object ID of the oper %s", id.Shorten())
			return nil, time.Time{}, errors.InconsistentDB
		}
		parsOK = oID.ParseSlice(rawObjID) == nil
		if !parsOK {
			log.Error("Can't parse object ID fetched from the entity DB")
			return nil, time.Time{}, errors.Parsing
		}
	} else {
		oID = *objID
	}

	sig, err := crypto.ParseSignature(encodedSig)
	if err != nil {
		log.Errorf("Can't parse signature fetched from DB: %v", err)
		return nil, time.Time{}, errors.Parsing
	}
	o, err := entity.NewOperation(
		(entity.OperationType)(typ),
		(entity.OperationReason)(reason),
		comment,
		&authID,
		&oID,
		perfDate,
		sig)
	if err != nil {
		log.Errorf("The operation '%s' fetched from DB is invalid", o)
		return nil, time.Time{}, errors.InconsistentDB
	}
	return o, tmStored, nil
}

func (d *EntityDatabase) GetOperationsOnUser(uid *entity.ID) ([]*entity.Operation, error) {
	log.Debugf("Fetching operations on user %s from the database", uid.Shorten())
	query := `
	SELECT Operations.Id,
	       Operations.Type,
	       Operations.Reason,
	       Operations.Comment,
	       Operations.Author_id,
	       Operations.Timestamp,
	       Operations.Signature
	FROM Operations
	INNER JOIN Operations_on_Users on Operations.Id=Operations_on_Users.Operation_id
	WHERE Operations_on_Users.User_id=?
	GROUP BY Operations.Id
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, uid[:])
	if err != nil {
		log.Errorf("Error fetching operations from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanOperationRows(rows, uid)
}

func (d *EntityDatabase) GetOperationsOnMessage(mid *entity.ID) ([]*entity.Operation, error) {
	log.Debugf("Fetching operations on message %s from the database", mid.Shorten())
	query := `
	SELECT Operations.Id,
	       Operations.Type,
	       Operations.Reason,
	       Operations.Comment,
	       Operations.Author_id,
	       Operations.Timestamp,
	       Operations.Signature
	FROM Operations
	INNER JOIN Operations_on_Messages on Operations.Id=Operations_on_Messages.Operation_id
	WHERE Operations_on_Messages.Message_id=?
	GROUP BY Operations.Id
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, mid[:])
	if err != nil {
		log.Errorf("Error fetching operations from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanOperationRows(rows, mid)
}

func (d *EntityDatabase) GetOperation(oid *entity.ID) (*entity.Operation, error) {
	log.Debugf("Fetching operation with id '%s' from the database", oid)
	query := `
	SELECT Operations.Id,
	       Operations.Type,
	       Operations.Reason,
	       Operations.Comment,
	       Operations.Author_id,
	       Operations.Timestamp,
	       Operations.Signature,
	       Operations_on_Messages.Message_id,
	       Operations_on_Users.User_id
	FROM Operations
	LEFT JOIN Operations_on_Messages on Operations.Id=Operations_on_Messages.Operation_id
	LEFT JOIN Operations_on_Users on Operations.Id=Operations_on_Users.Operation_id
	WHERE Operations.Id=?
	GROUP BY Operations.Id
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, oid[:])
	if err != nil {
		log.Errorf("Error fetching operation from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	opers, err := scanOperationRows(rows, nil)
	switch {
	case err != nil:
		log.Errorf("Error fetching message from the database: %v", err)
		return nil, errors.DBOperFailed
	case opers == nil:
		log.Debug("No message with that ID.")
		return nil, errors.NoSuchEntity
	case len(opers) > 1:
		log.Warningf("Multiple operations found with id=%s", oid)
	default:
		log.Debug("The message found successfully")
	}
	return opers[0], nil
}

func (d *EntityDatabase) GetOperationsStoredAfter(ts time.Time, limit int) ([]*entity.StoredOperation, error) {
	log.Debugf("Fetching operations since %s from the database", ts.Format(time.RFC3339))
	query := `
	SELECT Operations.Id,
	       Operations.Type,
	       Operations.Reason,
	       Operations.Comment,
	       Operations.Author_id,
	       Operations.Timestamp,
	       Operations.Signature,
	       Operations_on_Messages.Message_id,
	       Operations_on_Users.User_id,
	       Operations.TimeStored
	FROM Operations
	LEFT JOIN Operations_on_Messages on Operations.Id=Operations_on_Messages.Operation_id
	LEFT JOIN Operations_on_Users on Operations.Id=Operations_on_Users.Operation_id
	WHERE Operations.TimeStored>=?
	GROUP BY Operations.Id
	LIMIT ?
	`
	db := (*sql.DB)(d)
	rows, err := db.Query(query, ts, limit)
	if err != nil {
		log.Errorf("Error fetching operations from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanStoredOperationRows(rows, nil)
}

func (d *EntityDatabase) HasOperation(eid *entity.ID) (bool, error) {
	log.Debugf("Checking whether DB contains operation with id '%s'", eid)

	// FIXME: This is definitely not the most efficient implementation.
	var typ int
	query := `SELECT Type FROM Operations WHERE Id=?`
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
