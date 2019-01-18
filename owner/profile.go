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

package owner

import (
	"sync"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/sqlite"
	"vminko.org/dscuss/subs"
)

// Profile is a proxy for the profile database, which implements caching and
// provides few additional functions for managing owner's settings.
type Profile struct {
	db         *sqlite.ProfileDatabase
	moderators []*entity.ID
	modersMx   sync.Mutex
	subs       subs.Subscriptions
	subsMx     sync.Mutex
	selfID     *entity.ID
}

func NewProfile(db *sqlite.ProfileDatabase, id *entity.ID) *Profile {
	return &Profile{db: db, selfID: id}
}

func (p *Profile) Close() error {
	return p.db.Close()
}

func (p *Profile) PutModerator(id *entity.ID) error {
	if *id == *p.selfID {
		return errors.AlreadyModerator
	}
	p.modersMx.Lock()
	defer p.modersMx.Unlock()
	p.moderators = nil
	return p.db.PutModerator(id)
}

func (p *Profile) RemoveModerator(id *entity.ID) error {
	if *id == *p.selfID {
		return errors.ForbiddenOperation
	}
	p.modersMx.Lock()
	defer p.modersMx.Unlock()
	p.moderators = nil
	return p.db.RemoveModerator(id)
}

func (p *Profile) HasModerator(id *entity.ID) (bool, error) {
	if *id == *p.selfID {
		return true, nil
	}
	mm := p.GetModerators()
	for _, m := range mm {
		if *m == *id {
			return true, nil
		}
	}
	return false, nil
}

func (p *Profile) GetModerators() []*entity.ID {
	p.modersMx.Lock()
	defer p.modersMx.Unlock()
	if p.moderators == nil {
		var err error
		p.moderators, err = p.db.GetModerators()
		if err != nil {
			log.Fatalf("Failed to fetch owner's subscriptions from "+
				"the profile database: %v", err)
		}
		p.moderators = append(p.moderators, p.selfID)
	}
	res := make([]*entity.ID, len(p.moderators))
	copy(res, p.moderators)
	return res
}

func (p *Profile) PutSubscription(t subs.Topic) error {
	p.subsMx.Lock()
	defer p.subsMx.Unlock()
	p.subs = nil
	return p.db.PutSubscription(t)
}

func (p *Profile) RemoveSubscription(t subs.Topic) error {
	p.subsMx.Lock()
	defer p.subsMx.Unlock()
	p.subs = nil
	return p.db.RemoveSubscription(t)
}

func (p *Profile) GetSubscriptions() subs.Subscriptions {
	p.subsMx.Lock()
	defer p.subsMx.Unlock()
	if p.subs == nil {
		var err error
		p.subs, err = p.db.GetSubscriptions()
		if err != nil {
			log.Fatalf("Failed to fetch owner's subscriptions from "+
				"the profile database: %v", err)
		}
	}
	return p.subs.Copy()
}

func (p *Profile) PutUserHistory(h *entity.UserHistory) error {
	return p.db.PutUserHistory(h)
}

func (p *Profile) GetUserHistory(id *entity.ID) (*entity.UserHistory, error) {
	return p.db.GetUserHistory(id)
}

func (p *Profile) GetFullHistory() []*entity.UserHistory {
	h, err := p.db.GetFullHistory()
	if err != nil {
		log.Fatalf("Failed to fetch full history from the profile database: %v", err)
	}
	return h
}
