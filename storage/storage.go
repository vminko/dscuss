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

// Storage is a mediator between Owner and Peers. It provides subscriptions to
// new entities and functions for managing entities.
type Storage struct {
	db          *sqlite.EntityDatabase
	observers   []chan<- entity.Entity
	observersMx sync.Mutex
}

func New(db *sqlite.EntityDatabase) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AttachObserver(c chan<- entity.Entity) {
	s.observersMx.Lock()
	defer s.observersMx.Unlock()
	s.observers = append(s.observers, c)
}

func (s *Storage) DetachObserver(c chan<- entity.Entity) {
	s.observersMx.Lock()
	defer s.observersMx.Unlock()
	for i, o := range s.observers {
		if o == c {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return
		}
	}
}

func (s *Storage) notifyObservers(e entity.Entity, sender chan<- entity.Entity) {
	s.observersMx.Lock()
	defer s.observersMx.Unlock()
	for i, o := range s.observers {
		if o == sender {
			continue
		}
		log.Debugf("Notifying observer #%d", i)
		select {
		case o <- e:
			log.Debugf("Entity %s passes to observer #%d", e.Desc(), i)
		default:
			log.Debugf("Failed to pass entity %s to observer #%d", e.Desc(), i)
		}
	}
}

func (s *Storage) GetUser(eid *entity.ID) (*entity.User, error) {
	return s.db.GetUser(eid)
}

func (s *Storage) HasUser(id *entity.ID) (bool, error) {
	return s.db.HasUser(id)
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

func (s *Storage) GetRoot(m *entity.Message) (*entity.Message, error) {
	for m.IsReply() {
		p, err := s.db.GetMessage(&m.ParentID)
		if err != nil {
			return nil, err
		}
		m = p
	}
	return m, nil
}

func (s *Storage) GetOperationsOnUser(uid *entity.ID) ([]*entity.Operation, error) {
	return s.db.GetOperationsOnUser(uid)
}

func (s *Storage) GetOperationsOnMessage(mid *entity.ID) ([]*entity.Operation, error) {
	return s.db.GetOperationsOnMessage(mid)
}

func (s *Storage) GetOperation(oid *entity.ID) (*entity.Operation, error) {
	return s.db.GetOperation(oid)
}

func (s *Storage) HasEntity(eid *entity.ID) (bool, error) {
	h, err := s.db.HasUser(eid)
	if h {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	h, err = s.db.HasMessage(eid)
	if h {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return s.db.HasOperation(eid)
}

func (s *Storage) GetEntity(eid *entity.ID) (entity.Entity, error) {
	u, err := s.db.GetUser(eid)
	if err == nil {
		return (entity.Entity)(u), nil
	}
	if err != errors.NoSuchEntity {
		return nil, err
	}

	m, err := s.db.GetMessage(eid)
	if err == nil {
		return (entity.Entity)(m), nil
	}
	if err != errors.NoSuchEntity {
		return nil, err
	}

	o, err := s.db.GetOperation(eid)
	if err == nil {
		return (entity.Entity)(o), nil
	}

	return nil, err
}

func (s *Storage) PutEntity(ent entity.Entity, sender chan<- entity.Entity) error {
	var err error
	switch e := ent.(type) {
	case *entity.User:
		err = s.db.PutUser(e)
	case *entity.Message:
		err = s.db.PutMessage(e)
	case *entity.Operation:
		err = s.db.PutOperation(e)
	default:
		log.Fatalf("BUG: unknown entity type %T.", ent)
	}
	if err != nil {
		return err
	}
	s.notifyObservers(ent, sender)
	return nil
}
