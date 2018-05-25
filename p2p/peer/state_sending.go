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

package peer

import (
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
)

type StateSending struct {
	outgoingEntity *entity.Entity
}

func newStateSending(e *entity.Entity) *StateSending {
	return &StateSending{outgoingEntity: e}
}

func (ss *StateSending) Perform(p *Peer) (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", ss.Name())
	log.Debugf("State %s is not implemented yet", ss.Name())
	return newStateIdle(), nil
}

func (ss *StateSending) Name() string {
	return "Sending"
}

func (ss *StateSending) ID() StateID {
	return StateIDSending
}
