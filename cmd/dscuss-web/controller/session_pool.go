/*
This file is part of Dscuss.
Copyright (C) 2019  Vitaly Minko

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

package controller

import (
	"sync"
	"time"
	"vminko.org/dscuss/log"
)

type SessionPool struct {
	ss map[string]*Session
	mx sync.RWMutex
}

func newSessionPool() *SessionPool {
	return &SessionPool{
		ss: make(map[string]*Session),
	}
}

func (sp *SessionPool) add(s *Session) {
	sp.mx.Lock()
	defer sp.mx.Unlock()
	_, ok := sp.ss[s.ID]
	if ok {
		log.Fatalf("BUG: attempt to add duplicated session %s", s.ID)
	}
	sp.ss[s.ID] = s
}

func (sp *SessionPool) load(id string) (*Session, bool) {
	sp.mx.RLock()
	defer sp.mx.RUnlock()
	val, ok := sp.ss[id]
	return val, ok
}

func (sp *SessionPool) removeSession(id string) {
	isFound := false
	sp.mx.Lock()
	for _, s := range sp.ss {
		if s.ID == id {
			delete(sp.ss, id)
			isFound = true
			break
		}
	}
	log.Debugf("Size of SessionPool became %d", len(sp.ss))
	sp.mx.Unlock()
	if !isFound {
		log.Errorf("Failed to remove %s from the SessionsPool", id)
	}
}

func (sp *SessionPool) removeExpiredSessions() {
	sp.mx.RLock()
	tmp := make(map[string]*Session)
	for k, v := range sp.ss {
		tmp[k] = v
	}
	sp.mx.RUnlock()
	for id, s := range tmp {
		if s.UpdatedDate.Before(time.Now().Add(-maxSessionLife)) {
			sp.removeSession(s.ID)
			log.Debugf("Removing expired session %s.", id)
		}
	}
}
