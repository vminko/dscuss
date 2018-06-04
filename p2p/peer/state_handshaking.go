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
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/packet"
)

type StateHandshaking struct {
	p *Peer
	u *entity.User
}

func newStateHandshaking(p *Peer) *StateHandshaking {
	return &StateHandshaking{p: p}
}

func (s *StateHandshaking) Perform() (nextState State, err error) {
	var perfErr error
	perfUnlessErr := func(f func() error) {
		if perfErr != nil {
			return
		}
		perfErr = f()
	}
	if s.p.Conn.IsIncoming() {
		perfUnlessErr(s.readAndProcessUser)
		perfUnlessErr(s.sendUser)
		perfUnlessErr(s.readAndProcessHello)
		perfUnlessErr(s.sendHello)
	} else {
		perfUnlessErr(s.sendUser)
		perfUnlessErr(s.readAndProcessUser)
		perfUnlessErr(s.sendHello)
		perfUnlessErr(s.readAndProcessHello)
	}
	perfUnlessErr(s.finalize)
	if perfErr != nil {
		log.Errorf("Failed to handshake with %s %v", s.p.Desc(), perfErr)
		return nil, perfErr
	}
	return newStateIdle(s.p), nil
}

func (s *StateHandshaking) sendUser() error {
	uPckt := packet.New(packet.TypeUser, s.p.owner.User, s.p.owner.Signer)
	err := s.p.Conn.Write(uPckt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", uPckt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateHandshaking) readAndProcessUser() error {
	uPckt, err := s.p.Conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}
	if uPckt.Body.Type != packet.TypeUser {
		log.Errorf("Protocol violation detected:"+
			" peer '%s' sent unexpected packet of type '%s'. Expected: %s.",
			s.p.Desc(), uPckt.Body.Type, packet.TypeUser)
		return errors.ProtocolViolation
	}
	i, err := uPckt.DecodePayload()
	if err != nil {
		log.Errorf("Failed to decode payload of packet '%s': %v", uPckt.Desc(), err)
		return errors.Parsing
	}
	u, ok := (i).(*entity.User)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	s.u = u
	return nil
}

func (s *StateHandshaking) sendHello() error {
	// TBD
	return nil
}

func (s *StateHandshaking) readAndProcessHello() error {
	// TBD
	return nil
}

func (s *StateHandshaking) finalize() error {
	_, err := s.p.owner.DB.GetUser(&s.u.ID)
	if err == errors.NoSuchEntity {
		s.p.owner.DB.PutUser(s.u)
	} else if err != nil {
		log.Errorf("Unexpected error occurred while getting user from the DB: %v", err)
		return errors.Database
	}
	return nil
}

func (s *StateHandshaking) Name() string {
	return "Handshaking"
}

func (s *StateHandshaking) ID() StateID {
	return StateIDHandshaking
}
