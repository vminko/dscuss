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

type StateSending struct {
	p              *Peer
	outgoingEntity entity.Entity
}

func newStateSending(p *Peer, e entity.Entity) *StateSending {
	return &StateSending{p, e}
}

func (s *StateSending) perform() (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", s.p.Desc(), s.Name())

	// TBD: check if outgoingEntity is relevant for this peer

	err = s.sendAnnounce(s.outgoingEntity.ID())
	if err != nil {
		log.Errorf("Failed to send announce for %s: %v",
			s.outgoingEntity.ID().Shorten(), err)
		return nil, err
	}

	pkt, err := s.p.Conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return nil, err
	}
	if !pkt.VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature", s.p.Desc())
		return nil, errors.ProtocolViolation
	}

	verifyType := func(t packet.Type) bool {
		return t == packet.TypeAck || t == packet.TypeReq || t == packet.TypeAnnounce
	}
	pkt.VerifyHeaderFull(verifyType, s.p.User.ID())

	switch pkt.Body.Type {
	case packet.TypeAnnounce:
		// Collision detected: both peers tried to send announce
		// simultaneously.
		return newStateIdle(s.p), nil
	case packet.TypeAck:
		// Nothing to do
	case packet.TypeReq:
		err = s.processReq(pkt)
		if err != nil {
			log.Errorf("Error processing req from peer %s: %v", s.p.Desc(), err)
			return nil, err
		}
	default:
		log.Fatal("BUG: packet type validation failed.")
	}

	return newStateIdle(s.p), nil
}

func (s *StateSending) sendAnnounce(id *entity.ID) error {
	pld := packet.NewPayloadAnnounce(id)
	pkt := packet.New(packet.TypeAnnounce, s.p.User.ID(), pld, s.p.owner.Signer)
	err := s.p.Conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateSending) processReq(pkt *packet.Packet) error {
	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of a req '%s': %v", pkt.Desc(), err)
		return errors.Parsing
	}
	r, ok := (i).(*packet.PayloadReq)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}

	var e entity.Entity = s.outgoingEntity
	if r.ID != *e.ID() {
		e, err = s.p.storage.GetEntity(&r.ID)
		if err != nil {
			log.Errorf("Failed to get requested entity from the DB: %v", err)
			return err
		}
	}

	err = s.sendEntity(e)
	if err != nil {
		log.Infof("Failed to send outgoing entity to '%s': %v", s.p.Desc(), err)
		return err
	}

	err = s.readAndProcessAck()
	if err != nil {
		log.Infof("Failed to receive ack for entity from '%s': %v", s.p.Desc(), err)
		return err
	}

	return nil
}

func (s *StateSending) sendEntity(e entity.Entity) error {
	var t packet.Type
	switch e.Type() {
	case entity.TypeMessage:
		t = packet.TypeMessage
	case entity.TypeOperation:
		t = packet.TypeOperation
	default:
		log.Fatal("BUG: user entities are not to be advertised.")
	}
	pkt := packet.New(t, s.p.User.ID(), e, s.p.owner.Signer)
	err := s.p.Conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateSending) readAndProcessAck() error {
	pkt, err := s.p.Conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}
	if !pkt.VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature", s.p.Desc())
		return errors.ProtocolViolation
	}
	if pkt.VerifyHeader(packet.TypeAck, s.p.User.ID()) != nil {
		log.Infof("Peer %s sent packet with invalid header", s.p.Desc())
		return errors.ProtocolViolation
	}

	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of packet '%s': %v", pkt.Desc(), err)
		return errors.Parsing
	}
	_, ok := (i).(*packet.PayloadAck)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	return nil
}

func (s *StateSending) Name() string {
	return "Sending"
}

func (s *StateSending) ID() StateID {
	return StateIDSending
}
