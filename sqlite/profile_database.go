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

package sqlite

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"time"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/subs"
)

// ProfileDatabase stores personal settings of the owner.
// It includes owner's subscriptions, list of chosen moderators. flags
// that message has been read (TBD) and so on.
type ProfileDatabase sql.DB

func OpenProfileDatabase(fileName string) (*ProfileDatabase, error) {
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
	exec("PRAGMA foreign_keys=ON")
	//exec("PRAGMA journal_mode=WAL")
	exec("CREATE TABLE IF NOT EXISTS Moderators (" +
		"  User_id          BLOB PRIMARY KEY)")
	exec("CREATE TABLE IF NOT EXISTS Subscriptions (" +
		"  Id               INTEGER PRIMARY KEY AUTOINCREMENT," +
		"  Topic            TEXT NOT NULL UNIQUE)")
	exec("CREATE TABLE IF NOT EXISTS User_Histories (" +
		"  User_id          BLOB PRIMARY KEY," +
		"  TimeDisconnected TIMESTAMP NOT NULL)")
	exec("CREATE TABLE IF NOT EXISTS User_Subscriptions (" +
		"  User_id          BLOB NOT NULL REFERENCES User_Histories ON DELETE CASCADE," +
		"  Topic            TEXT NOT NULL)")
	// TBD: create indexes?
	if execErr != nil {
		log.Errorf("Unable to initialize the profile database: %s", execErr.Error())
		return nil, errors.DBOperFailed
	}

	return (*ProfileDatabase)(db), nil
}

func (pd *ProfileDatabase) Close() error {
	db := (*sql.DB)(pd)
	err := db.Close()
	if err != nil {
		log.Errorf("Unable to close the profile database: %v", err)
		return errors.DBOperFailed
	}
	return nil
}

func (pd *ProfileDatabase) PutModerator(id *entity.ID) error {
	log.Debugf("Adding moderator `%s' to the profile database", id.Shorten())
	query := `INSERT INTO Moderators ( User_id ) VALUES (?)`
	db := (*sql.DB)(pd)
	_, err := db.Exec(query, id[:])
	if err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if ok && sqliteErr.Code == sqlite3.ErrConstraint {
			log.Warningf("Attempt to duplicate a moderator: %s", err.Error())
			return errors.AlreadyModerator
		}
		log.Errorf("Can't execute 'PutModerator' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	return nil
}

func (pd *ProfileDatabase) RemoveModerator(id *entity.ID) error {
	log.Debugf("Removing moderator `%s' from the profile database", id.Shorten())
	query := `DELETE FROM Moderators WHERE User_ID=?`
	db := (*sql.DB)(pd)
	res, err := db.Exec(query, id[:])
	if err != nil {
		log.Errorf("Can't execute 'RemoveModerator' statement: %s", err.Error())
		return errors.DBOperFailed
	} else {
		count, err := res.RowsAffected()
		if err != nil {
			log.Errorf("Error getting number of affected rows: %s", err.Error())
			return err
		} else {
			if count != 1 {
				return errors.NoSuchModerator
			}
		}
	}
	return nil
}

func (pd *ProfileDatabase) GetModerators() ([]*entity.ID, error) {
	log.Debugf("Fetching moderators from the profile database")
	query := `SELECT User_id FROM Moderators`
	db := (*sql.DB)(pd)
	rows, err := db.Query(query)
	if err != nil {
		log.Errorf("Error fetching moderators from the profile database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	var res []*entity.ID
	for rows.Next() {
		var rawID []byte
		err := rows.Scan(&rawID)
		if err != nil {
			log.Errorf("Error scanning moderator row: %v", err)
			return nil, errors.DBOperFailed
		}

		var id entity.ID
		if id.ParseSlice(rawID) != nil {
			log.Error("Can't parse an ID fetched from the profile DB")
			return nil, errors.Parsing
		}
		log.Debugf("Found moderator id %s", id.String())
		res = append(res, &id)
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting next moderator row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func (pd *ProfileDatabase) PutSubscription(t subs.Topic) error {
	log.Debugf("Adding subscription `%s' to the profile database", t)
	query := `INSERT INTO Subscriptions ( Topic ) VALUES (?)`
	db := (*sql.DB)(pd)
	_, err := db.Exec(query, t.String())
	if err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if ok && sqliteErr.Code == sqlite3.ErrConstraint {
			log.Warningf("Attempt to duplicate a subscription: %s", err.Error())
			return errors.AlreadySubscribed
		}
		log.Errorf("Can't execute 'PutSubscription' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	return nil
}

func (pd *ProfileDatabase) RemoveSubscription(t subs.Topic) error {
	log.Debugf("Removing subscription `%s' from the profile database", t)
	query := `DELETE FROM Subscriptions WHERE Topic=?`
	db := (*sql.DB)(pd)
	res, err := db.Exec(query, t.String())
	if err != nil {
		log.Errorf("Can't execute 'RemoveSubscription' statement: %s", err.Error())
		return errors.DBOperFailed
	} else {
		count, err := res.RowsAffected()
		if err != nil {
			log.Errorf("Error getting number of affected rows: %s", err.Error())
			return err
		} else {
			if count != 1 {
				return errors.NotSubscribed
			}
		}
	}
	return nil
}

func (pd *ProfileDatabase) GetSubscriptions() (subs.Subscriptions, error) {
	log.Debugf("Fetching subscriptions from the profile database")
	query := `SELECT Topic FROM Subscriptions`
	db := (*sql.DB)(pd)
	rows, err := db.Query(query)
	if err != nil {
		log.Errorf("Error fetching subscriptions from the profile database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	var res subs.Subscriptions
	for rows.Next() {
		var tStr string
		err := rows.Scan(&tStr)
		if err != nil {
			log.Errorf("Error scanning subscription row: %v", err)
			return nil, errors.DBOperFailed
		}

		t, err := subs.NewTopic(tStr)
		if err != nil {
			log.Errorf("Can't parse a topic fetched from the profile DB: %v", err)
			return nil, errors.Parsing
		}
		log.Debugf("Found subscription %s", tStr)
		res = res.AddTopic(t)
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting next subscription row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func (pd *ProfileDatabase) PutUserHistory(h *entity.UserHistory) error {
	log.Debugf("Adding history of user `%s' to the profile database", h.ID.Shorten())
	query := `INSERT OR REPLACE INTO User_Histories ( User_id, TimeDisconnected ) VALUES (?,?)`
	db := (*sql.DB)(pd)
	_, err := db.Exec(query, h.ID[:], h.Disconnected)
	if err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if ok && sqliteErr.Code == sqlite3.ErrConstraint {
			log.Warningf("Attempt to duplicate a user history record: %s", err.Error())
			return errors.DuplicationAttempt
		}
		log.Errorf("Can't execute 'PutUserHistory' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	for _, t := range h.Subs {
		err := pd.putUserSubscription(h.ID, t)
		if err != nil {
			log.Errorf("The DB is corrupted. History for user %s is saved without topic %s",
				h.ID.Shorten(), t)
			return err
		}
	}

	return nil
}

func (pd *ProfileDatabase) putUserSubscription(id *entity.ID, t subs.Topic) error {
	log.Debugf("Adding user subscription `%s' to the profile database", t)
	query := `INSERT INTO User_Subscriptions (User_id, Topic) VALUES (? ,?)`
	db := (*sql.DB)(pd)
	_, err := db.Exec(query, id[:], t.String())
	if err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if ok && sqliteErr.Code == sqlite3.ErrConstraint {
			log.Warningf("Attempt to duplicate a subscription: %s", err.Error())
			return errors.DuplicationAttempt
		}
		log.Errorf("Can't execute 'putUserSubscription' statement: %s", err.Error())
		return errors.DBOperFailed
	}
	return nil
}

func (pd *ProfileDatabase) GetUserHistory(id *entity.ID) (*entity.UserHistory, error) {
	log.Debugf("Fetching history of user '%s' from the profile database", id)

	var timeDisc time.Time
	var subsStr string
	query := `
	SELECT User_Histories.TimeDisconnected,
	       GROUP_CONCAT(User_Subscriptions.Topic, "\n")
	FROM User_Histories
	LEFT JOIN User_Subscriptions on User_Histories.User_id=User_Subscriptions.User_id
	WHERE User_Histories.User_id=?
	GROUP BY User_Histories.User_id
	`
	db := (*sql.DB)(pd)
	err := db.QueryRow(query, id[:]).Scan(&timeDisc, &subsStr)
	switch {
	case err == sql.ErrNoRows:
		log.Debug("No history for that user.")
		return nil, errors.NoUserHistory
	case err != nil:
		log.Errorf("Error fetching user history from the peer database: %v", err)
		return nil, errors.DBOperFailed
	default:
		log.Debug("User history found successfully")
	}

	s, err := subs.ReadString(subsStr)
	if err != nil {
		log.Errorf("The subscriptions '%s' fetched from DB are invalid", subsStr)
		return nil, errors.InconsistentDB
	}

	return &entity.UserHistory{id, timeDisc, s}, nil
}

func scanHistoryRows(rows *sql.Rows) ([]*entity.UserHistory, error) {
	var res []*entity.UserHistory
	for rows.Next() {
		var rawID []byte
		var timeDisc time.Time
		var subsStr string
		err := rows.Scan(
			&rawID,
			&timeDisc,
			&subsStr)
		if err != nil {
			log.Errorf("Error scanning history row: %v", err)
			return nil, errors.DBOperFailed
		}

		var id entity.ID
		if id.ParseSlice(rawID) != nil {
			log.Error("Can't parse an ID fetched from DB")
			return nil, errors.Parsing
		}
		log.Debugf("Found history record for %s", &id)

		s, err := subs.ReadString(subsStr)
		if err != nil {
			log.Errorf("The subscriptions '%s' fetched from DB are invalid", subsStr)
			return nil, errors.InconsistentDB
		}
		res = append(res, &entity.UserHistory{&id, timeDisc, s})
	}
	err := rows.Err()
	if err != nil {
		log.Errorf("Error getting next history row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}

func (pd *ProfileDatabase) GetFullHistory() ([]*entity.UserHistory, error) {
	log.Debug("Fetching full history of users from the profile database")
	query := `
	SELECT User_Histories.User_id,
	       User_Histories.TimeDisconnected,
	       GROUP_CONCAT(User_Subscriptions.Topic, char(10))
	FROM User_Histories
	LEFT JOIN User_Subscriptions on User_Histories.User_id=User_Subscriptions.User_id
	GROUP BY User_Histories.User_id
	`
	db := (*sql.DB)(pd)
	rows, err := db.Query(query)
	if err != nil {
		log.Errorf("Error fetching history of users from the database: %v", err)
		return nil, errors.DBOperFailed
	}
	defer rows.Close()
	return scanHistoryRows(rows)
}
