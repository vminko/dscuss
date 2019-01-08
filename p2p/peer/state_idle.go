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
	"time"
	"vminko.org/dscuss/log"
)

// StateIdle implements the idle protocol (when peer is waiting for new entities
// from either side).
type StateIdle struct {
	p *Peer
}

const (
	IdleTimeout time.Duration = 1 * time.Second
)

func newStateIdle(p *Peer) *StateIdle {
	return &StateIdle{p}
}

func (s *StateIdle) perform() (nextState State, err error) {
	for {
		log.Debugf("Peer %s is trying to read packets...", s.p)
		pckt, err := s.p.conn.ReadFull(IdleTimeout)
		if err != nil {
			// Timeout means there were no new packets from this
			// peer. That's not en error, ignore timeout.
			if neterr, ok := err.(net.Error); !(ok && neterr.Timeout()) {
				log.Debugf("Peer %s failed to read packet: %v", s.p, err)
				return nil, err
			}
		} else {
			log.Debugf("Peer %s received packet %s", s.p, pckt)
			return newStateReceiving(s.p, pckt, s), nil
		}

		log.Debugf("Peer %s is checking for new outEntity...", s.p)
		select {
		case e, ok := <-s.p.outEntityChan:
			if ok {
				log.Debugf("Peer %s got new outEntity %s", s.p, e)
				return newStateSending(s.p, e, time.Now(), s), nil
			} else {
				log.Debugf("Peer %s: outEntityChan was closed", s.p)
				return nil, err
			}
		default:
			log.Debugf("Peer %s: no new entities in outEntityChan", s.p)
		}
	}
}

func (s *StateIdle) Name() string {
	return "Idle"
}

func (s *StateIdle) ID() StateID {
	return StateIDIdle
}
