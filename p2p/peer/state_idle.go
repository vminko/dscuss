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
	"net"
	"vminko.org/dscuss/log"
)

type StateIdle struct {
	p *Peer
}

func newStateIdle(p *Peer) *StateIdle {
	return &StateIdle{p}
}

func (s *StateIdle) Perform() (nextState State, err error) {
	for {
		pckt, err := s.p.Conn.Read()
		if err != nil {
			if neterr, ok := err.(net.Error); !(ok && neterr.Timeout()) {
				log.Debugf("Peer %s failed to read packet: %v", s.p.Desc(), err)
				return nil, err
			}
		} else {
			log.Debugf("Peer %s received packet %s", s.p.Desc(), pckt.Desc())
			return newStateReceiving(s.p, pckt), nil
		}

		select {
		case e, ok := <-s.p.outEntityChan:
			if ok {
				log.Debugf("Peer %s got new outEntity %s", s.p.Desc(), e.Desc())
				return newStateSending(s.p, e), nil
			} else {
				log.Debugf("Peer %s: outEntityChan was closed", s.p.Desc())
				return nil, err
			}
		default:
			log.Debugf("Peer %s: no new entities in outEntityChan", s.p.Desc())
		}
	}
}

func (s *StateIdle) Name() string {
	return "Idle"
}

func (s *StateIdle) ID() StateID {
	return StateIDIdle
}
