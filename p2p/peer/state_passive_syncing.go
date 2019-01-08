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

package peer

import (
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/packet"
)

// StatePassiveSyncing implements the passive part of the sync protocol.
// In case of success, switches peer to either Idle (for peers connected via
// active connections) or StateActiveSyncing (for peers connected via passive
// connections).
type StatePassiveSyncing struct {
	p *Peer
}

func newStatePassiveSyncing(p *Peer) *StatePassiveSyncing {
	return &StatePassiveSyncing{p: p}
}

func (s *StatePassiveSyncing) perform() (nextState State, err error) {
	log.Debugf("Peer %s is trying to read packets...", s.p)
	pkt, err := s.p.conn.Read()
	if err != nil {
		log.Debugf("Peer %s failed to read packet: %v", s.p, err)
		return nil, err
	}
	log.Debugf("Peer %s received packet %s", s.p, pkt)
	if !pkt.VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature", s.p)
		return nil, errors.ProtocolViolation
	}
	err = pkt.VerifyHeader(packet.TypeDone, s.p.owner.User.ID())
	if err != nil {
		return newStateReceiving(s.p, pkt, s), nil
	} else {
		if s.p.conn.IsActive() {
			return newStateIdle(s.p), nil
		} else {
			return newStateActiveSyncing(s.p), nil
		}
	}
}

func (s *StatePassiveSyncing) Name() string {
	return "StatePassiveSyncing"
}

func (s *StatePassiveSyncing) ID() StateID {
	return StateIDPassiveSyncing
}
