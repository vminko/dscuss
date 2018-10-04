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
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/sqlite"
	"vminko.org/dscuss/subs"
	"vminko.org/dscuss/thread"
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

func (s *Storage) GetUser(eid *entity.ID) (*entity.User, error) {
	return s.db.GetUser(eid)
}

func (s *Storage) GetMessage(eid *entity.ID) (*entity.Message, error) {
	return s.db.GetMessage(eid)
}

func (s *Storage) GetRootMessages(offset, limit int) ([]*entity.Message, error) {
	return s.db.GetRootMessages(offset, limit)
}

func (s *Storage) GetTopicMessages(topic subs.Topic, offset, limit int) ([]*entity.Message, error) {
	return s.db.GetTopicMessages(topic, offset, limit)
}

func (s *Storage) GetThread(root *entity.ID) (*thread.Node, error) {
	return s.db.GetThread(root)
}

func (s *Storage) HasMessage(id *entity.ID) (bool, error) {
	return s.db.HasMessage(id)
}

func (s *Storage) GetEntity(eid *entity.ID) (entity.Entity, error) {
	m, err := s.db.GetMessage(eid)
	if err == errors.NoSuchEntity {
		u, err := s.db.GetUser(eid)
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

func (s *Storage) PutEntity(ent entity.Entity, sender chan<- entity.Entity) error {
	var err error
	switch e := ent.(type) {
	case *entity.Message:
		err = s.db.PutMessage(e)
	case *entity.User:
		err = s.db.PutUser(e)
	// TBD: case *entity.Operation:
	default:
		log.Fatal("BUG: unexpected entity type %T.")
	}
	if err != nil {
		return err
	}
	s.notifyObservers(ent, sender)
	return nil
}
