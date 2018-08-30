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
	p              *Peer
	outgoingEntity entity.Entity
}

func newStateSending(p *Peer, e entity.Entity) *StateSending {
	return &StateSending{p, e}
}

func (s *StateSending) perform() (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", s.Name())
	log.Debugf("State %s is not implemented yet", s.Name())
	// TBD: check if outgoingEntity is relevant for this peer
	return newStateIdle(s.p), nil
}

func (s *StateSending) Name() string {
	return "Sending"
}

func (s *StateSending) ID() StateID {
	return StateIDSending
}
