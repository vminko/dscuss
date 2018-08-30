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

package storage

import (
	"sync"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/sqlite"
)

// Storage is a proxy for the entity database, which provides subscriptions to
// new entities.
type Storage struct {
	db        *sqlite.Database
	observers []chan<- entity.Entity
	mx        sync.Mutex
}

func New(db *sqlite.Database) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AttachObserver(c chan<- entity.Entity) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.observers = append(s.observers, c)
}

func (s *Storage) DetachObserver(c chan<- entity.Entity) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for i, o := range s.observers {
		if o == c {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return
		}
	}
}

func (s *Storage) notifyObservers(e entity.Entity, sender chan<- entity.Entity) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for i, o := range s.observers {
		if o != sender {
			log.Debugf("Notifying observer #%d", i)
			o <- e
		}
	}

}

func (s *Storage) PutUser(u *entity.User, sender chan<- entity.Entity) error {
	err := s.db.PutUser(u)
	if err != nil {
		return err
	}
	s.notifyObservers((entity.Entity)(u), sender)
	return nil
}

func (s *Storage) GetUser(eid *entity.ID) (*entity.User, error) {
	return s.db.GetUser(eid)
}

/*
func (s *Storage) PutMessage(m *entity.Message) error {
	err := s.db.PutMessage(m)
	if err != nil {
		return err
	}
	s.notifyObservers((*entity.Entity)(m))
	return nil
}

func (s *Storage) GetMessage(eid *entity.ID) (*entity.Message, error) {
	return s.db.GetMessage(eid)
}
*/

//TBD:
//GetRootMessages(mi MessageIterator)
//GetMessageReplies(id ID, mi MessageIterator)
