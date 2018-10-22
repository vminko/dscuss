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
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

// ProfileDatabase stores personal settings of the owner.
// It includes owner's subscriptions (TBD), list of chosen moderators. flags
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
	//exec("PRAGMA journal_mode=WAL")
	exec("CREATE TABLE IF NOT EXISTS  Moderator (" +
		"  User_id              BLOB PRIMARY KEY)")
	// TBD: create indexes?
	if execErr != nil {
		log.Errorf("Unable to initialize the database: %s", execErr.Error())
		return nil, errors.DBOperFailed
	}

	return (*ProfileDatabase)(db), nil
}

func (s *ProfileDatabase) Close() error {
	db := (*sql.DB)(s)
	err := db.Close()
	if err != nil {
		log.Errorf("Unable to close the database: %v", err)
		return errors.DBOperFailed
	}
	return nil
}

func (s *ProfileDatabase) PutModerator(id *entity.ID) error {
	log.Debugf("Adding moderator `%s' to the database", id.Shorten())
	query := `INSERT INTO Moderator ( User_id ) VALUES (?)`
	db := (*sql.DB)(s)
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

func (s *ProfileDatabase) RemoveModerator(id *entity.ID) error {
	log.Debugf("Removing moderator `%s' from the database", id.Shorten())
	query := `DELETE FROM Moderator WHERE User_ID=?`
	db := (*sql.DB)(s)
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

func (s *ProfileDatabase) GetModerators() ([]*entity.ID, error) {
	log.Debugf("Fetching moderators from the database")
	query := `SELECT User_id FROM Moderator`
	db := (*sql.DB)(s)
	rows, err := db.Query(query)
	if err != nil {
		log.Errorf("Error fetching moderators from the database: %v", err)
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
			log.Error("Can't parse an ID fetched from DB")
			return nil, errors.Parsing
		}
		log.Debugf("Found moderator id %s", id.String())
		res = append(res, &id)
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Error getting nexturow with moderator row: %v", err)
		return nil, errors.DBOperFailed
	}
	return res, nil
}
