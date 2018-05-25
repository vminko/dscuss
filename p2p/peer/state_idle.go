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

type StateIdle struct{}

func newStateIdle() *StateIdle {
	return new(StateIdle)
}

func (sh *StateIdle) Perform(p *Peer) (nextState State, err error) {
	for {
		pckt, err := p.Conn.Read()
		if err != nil {
			if neterr, ok := err.(net.Error); !(ok && neterr.Timeout()) {
				log.Debugf("Peer %s failed to read packet: %v", p.Desc(), err)
				return nil, err
			}
		} else {
			log.Debugf("Peer %s received packet %s", p.Desc(), pckt.Desc())
			return newStateReceiving(pckt), nil
		}

		select {
		case e, ok := <-p.outEntityChan:
			if ok {
				log.Debugf("Peer %s got new outgoing entity %s", p.Desc(), e.Desc())
				return newStateSending(e), nil
			} else {
				log.Debugf("Peer %s: outEntityChan was closed", p.Desc())
				return nil, err
			}
		default:
			log.Debugf("Peer %s: no new entities in outEntityChan", p.Desc())
		}
	}
}

func (sh *StateIdle) Name() string {
	return "Idle"
}

func (sh *StateIdle) ID() StateID {
	return StateIDIdle
}
