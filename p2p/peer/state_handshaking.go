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
	"vminko.org/dscuss/subs"
)

// StateHandshaking implements the handshaking protocol.
type StateHandshaking struct {
	p *Peer
	u *entity.User
	s subs.Subscriptions
}

func newStateHandshaking(p *Peer) *StateHandshaking {
	return &StateHandshaking{p: p}
}

func (s *StateHandshaking) perform() (nextState State, err error) {
	var perfErr error
	perfUnlessErr := func(f func() error) {
		if perfErr != nil {
			return
		}
		perfErr = f()
	}
	if s.p.conn.IsIncoming() {
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
	pkt := packet.New(packet.TypeUser, &entity.ZeroID, s.p.owner.User, s.p.owner.Signer)
	err := s.p.conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateHandshaking) readAndProcessUser() error {
	pkt, err := s.p.conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}

	// We can't check signature of the packet at this point, because we don't know
	// the user yet. Signature will be checked below.
	if pkt.VerifyHeader(packet.TypeUser, &entity.ZeroID) != nil {
		log.Infof("Peer %s sent packet with invalid header.", s.p.Desc())
		return errors.ProtocolViolation
	}

	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of packet '%s': %v", pkt.Desc(), err)
		return errors.Parsing
	}
	u, ok := (i).(*entity.User)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	if !u.IsValid() {
		log.Infof("Peer %s sent malformed User entity", s.p.Desc())
		return errors.ProtocolViolation
	}
	if !pkt.VerifySig(&u.PubKey) {
		log.Infof("Peer %s sent Hello packet with invalid signature", s.p.Desc())
		return errors.ProtocolViolation
	}
	isBanned, err := s.p.owner.View.IsUserBanned(u.ID())
	if err != nil {
		log.Fatalf("Failed check whether %s is banned: %v", u.ID().Shorten(), err)
	}
	if isBanned {
		return errors.UserBanned
	}

	s.u = u
	return nil
}

func (s *StateHandshaking) sendHello() error {
	hPld := packet.NewPayloadHello(s.p.owner.Subs)
	hPkt := packet.New(packet.TypeHello, s.u.ID(), hPld, s.p.owner.Signer)
	err := s.p.conn.Write(hPkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", hPkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateHandshaking) readAndProcessHello() error {
	pkt, err := s.p.conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}
	if !pkt.VerifySig(&s.u.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature", s.p.Desc())
		return errors.ProtocolViolation
	}
	if pkt.VerifyHeader(packet.TypeHello, s.p.owner.User.ID()) != nil {
		log.Infof("Peer %s sent packet with invalid header", s.p.Desc())
		return errors.ProtocolViolation
	}

	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of packet '%s': %v", pkt.Desc(), err)
		return errors.Parsing
	}
	h, ok := (i).(*packet.PayloadHello)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	if !h.IsValid() {
		log.Infof("Peer %s sent malformed Hello packet", s.p.Desc())
		return errors.ProtocolViolation
	}
	s.s = h.Subs
	return nil
}

func (s *StateHandshaking) finalize() error {
	// storage.HasUser will be a bit faster in this case, but handshaking
	// should happen rarely, so GetUser won't cause significant difference
	// in performance.
	_, err := s.p.storage.GetUser(s.u.ID())
	if err == errors.NoSuchEntity {
		err = s.p.storage.PutEntity((entity.Entity)(s.u), s.p.outEntityChan)
		if err != nil {
			log.Fatalf("Failed to put user into the DB: %v", err)
		}
	} else if err != nil {
		log.Fatalf("Unexpected error occurred while getting user from the DB: %v", err)
	}
	s.p.User = s.u
	s.p.Subs = s.s
	if !s.p.validator.ValidatePeer(s.p) {
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
