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
	"vminko.org/dscuss/sqlite"
)

// Profile is a proxy for the profile database, which implements caching and
// provides few additional functions for managing owner's settings.
type Profile struct {
	db         *sqlite.ProfileDatabase
	moderators []*entity.ID
	modersMx   sync.Mutex
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
	mm, err := p.GetModerators()
	if err != nil {
		return false, err
	}
	for _, m := range mm {
		if *m == *id {
			return true, nil
		}
	}
	return false, nil
}

func (p *Profile) GetModerators() ([]*entity.ID, error) {
	p.modersMx.Lock()
	defer p.modersMx.Unlock()
	if p.moderators == nil {
		var err error
		p.moderators, err = p.db.GetModerators()
		if err != nil {
			return nil, err
		}
		p.moderators = append(p.moderators, p.selfID)
	}
	res := make([]*entity.ID, len(p.moderators))
	copy(res, p.moderators)
	return res, nil
}
