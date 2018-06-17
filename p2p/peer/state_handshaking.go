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
	"time"
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
	uPkt := packet.New(packet.TypeUser, s.p.owner.User, s.p.owner.Signer)
	err := s.p.Conn.Write(uPkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", uPkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateHandshaking) readAndProcessUser() error {
	uPkt, err := s.p.Conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}
	if uPkt.Body.Type != packet.TypeUser {
		log.Errorf("Protocol violation detected:"+
			" peer '%s' sent unexpected packet of type '%s'. Expected: %s.",
			s.p.Desc(), uPkt.Body.Type, packet.TypeUser)
		return errors.ProtocolViolation
	}
	i, err := uPkt.DecodePayload()
	if err != nil {
		log.Errorf("Failed to decode payload of packet '%s': %v", uPkt.Desc(), err)
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
	hPld := packet.NewPayloadHello(&s.u.ID)
	hPkt := packet.New(packet.TypeHello, hPld, s.p.owner.Signer)
	err := s.p.Conn.Write(hPkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", hPkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateHandshaking) readAndProcessHello() error {
	hPkt, err := s.p.Conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}
	if hPkt.Body.Type != packet.TypeHello {
		log.Errorf("Protocol violation detected:"+
			" peer '%s' sent unexpected packet of type '%s'. Expected: %s.",
			s.p.Desc(), hPkt.Body.Type, packet.TypeHello)
		return errors.ProtocolViolation
	}
	i, err := hPkt.DecodePayload()
	if err != nil {
		log.Errorf("Failed to decode payload of packet '%s': %v", hPkt.Desc(), err)
		return errors.Parsing
	}
	h, ok := (i).(*packet.PayloadHello)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	if h.ReceiverID != s.p.owner.User.ID {
		log.Errorf("Protocol violation detected:"+
			" peer '%s' sent hello packet with wrong receiver ID: '%s'.",
			s.p.Desc(), h.ReceiverID.String())
		return errors.ProtocolViolation
	}

	const MaxTimeDiscrepancy = 5 * time.Second
	if time.Since(h.Time) > MaxTimeDiscrepancy {
		log.Errorf("Protocol violation detected:"+
			" peer '%s' sent hello packet with obsolete Time: '%s'.",
			s.p.Desc(), h.Time.Format(time.RFC3339))
		return errors.ProtocolViolation
	}
	//TBD: process subscriptions
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
	s.p.User = s.u
	if !s.p.validate(s.p) {
		log.Debugf("Peer validation failed")
		return errors.InvalidPeer
	}
	return nil
}

func (s *StateHandshaking) Name() string {
	return "Handshaking"
}

func (s *StateHandshaking) ID() StateID {
	return StateIDHandshaking
}
