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
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/packet"
)

type StateReceiving struct {
	p              *Peer
	pendingPackets []*packet.Packet
}

func newStateReceiving(p *Peer, pckt *packet.Packet) *StateReceiving {
	return &StateReceiving{p, []*packet.Packet{pckt}}
}

func (s *StateReceiving) perform() (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", s.p.Desc(), s.Name())
	if !s.pendingPackets[1].VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature: %s", s.p.Desc())
		return nil, errors.ProtocolViolation
	}
	log.Debugf("State %s is not implemented yet", s.Name())
	return newStateIdle(s.p), nil
}

func (s *StateReceiving) Name() string {
	return "Receiving"
}

func (s *StateReceiving) ID() StateID {
	return StateIDReceiving
}
