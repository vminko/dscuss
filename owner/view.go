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
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/storage"
	"vminko.org/dscuss/thread"
)

// View personalizes content for the owner. It applies operation performed by
// the owner's moderators.
type View struct {
	profile *Profile
	storage *storage.Storage
}

func NewView(profile *Profile, storage *storage.Storage) *View {
	return &View{profile, storage}
}

func (v *View) IsUserBanned(uid *entity.ID) (bool, error) {
	authOps, err := v.storage.GetOperationsOnUser(uid)
	if err != nil {
		log.Errorf("Failed to get operations on user %s: %v", uid.Shorten(), err)
		return false, err
	}
	for _, o := range authOps {
		isModer, err := v.profile.HasModerator(&o.AuthorID)
		if err != nil {
			log.Errorf("Failed to check whether %s is a moderator: %v",
				&o.AuthorID, err)
			return false, err
		}
		if isModer {
			switch o.OperationType() {
			case entity.OperationTypeBanUser:
				return true, nil
			default:
				log.Fatal("BUG: unknown entity type %T.")
			}
		}
	}
	return false, nil
}

func (v *View) applyOperationToMessage(m *entity.Message, op *entity.Operation) *entity.Message {
	switch op.OperationType() {
	case entity.OperationTypeRemoveMessage:
		return nil
	default:
		log.Fatal("BUG: unknown entity type %T.")
	}
	return m
}

func (v *View) ModerateMessage(m *entity.Message) (*entity.Message, error) {
	isBanned, err := v.IsUserBanned(&m.AuthorID)
	if err != nil {
		log.Errorf("Failed check whether %s is banned: %v", m.AuthorID.Shorten(), err)
		return nil, err
	}
	if isBanned {
		log.Debugf("Author of msg %s (user %s) is banned",
			m.ID().Shorten(), m.AuthorID.Shorten())
		return nil, nil
	}
	msgOps, err := v.storage.GetOperationsOnMessage(m.ID())
	if err != nil {
		log.Errorf("Failed to get operations on message %s: %v", m.ID().Shorten(), err)
		return nil, err
	}
	for _, o := range msgOps {
		isModer, err := v.profile.HasModerator(&o.AuthorID)
		if err != nil {
			log.Errorf("Failed to check whether %s is a moderator: %v",
				&o.AuthorID, err)
			return nil, err
		}
		if isModer {
			m = v.applyOperationToMessage(m, o)
			if m == nil {
				break
			}
		}
	}
	return m, nil
}

func (v *View) ModerateMessages(brd []*entity.Message) (res []*entity.Message, err error) {
	for _, m := range brd {
		mm, err := v.ModerateMessage(m)
		if err != nil {
			log.Errorf("Error moderating message %s: %v", m.ID().Shorten(), err)
			return nil, err
		}
		if mm != nil {
			res = append(res, mm)
		}
	}
	return res, nil
}

type ThreadModerator struct {
	v *View
}

func (tm *ThreadModerator) Handle(n *thread.Node) (*entity.Message, error) {
	m := n.Msg
	if m == nil {
		log.Fatal("BUG: thread node with nil message")
	}
	mm, err := tm.v.ModerateMessage(m)
	if err != nil {
		log.Errorf("Error moderating message %s: %v", m.ID().Shorten(), err)
		return nil, err
	}
	return mm, nil
}

func (v *View) ModerateThread(t *thread.Node) (*thread.Node, error) {
	tm := ThreadModerator{v}
	tvis := thread.NewModeratingVisitor(&tm)
	return t.Moderate(tvis)
}
