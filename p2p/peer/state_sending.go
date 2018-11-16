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

// StateSending implements the entity sending protocol.
type StateSending struct {
	p              *Peer
	outgoingEntity entity.Entity
}

func newStateSending(p *Peer, e entity.Entity) *StateSending {
	return &StateSending{p, e}
}

func (s *StateSending) perform() (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", s.p, s.Name())
	if !s.p.isInterestedInEntity(s.outgoingEntity) {
		log.Debugf("Peer %s is not interested in '%s'", s.p, s.outgoingEntity)
		return newStateIdle(s.p), nil
	}
	err = s.sendAnnounce(s.outgoingEntity.ID())
	if err != nil {
		log.Errorf("Failed to send announce for %s: %v",
			s.outgoingEntity.ID().Shorten(), err)
		return nil, err
	}
	acked := false
	for !acked {
		pkt, err := s.p.conn.Read()
		if err != nil {
			log.Errorf("Error receiving packet from the peer %s: %v", s.p, err)
			return nil, err
		}
		if !pkt.VerifySig(&s.p.User.PubKey) {
			log.Infof("Peer %s sent a packet with invalid signature", s.p)
			return nil, errors.ProtocolViolation
		}
		verifyType := func(t packet.Type) bool {
			return t == packet.TypeAck || t == packet.TypeReq || t == packet.TypeAnnounce
		}
		if pkt.VerifyHeaderFull(verifyType, s.p.owner.User.ID()) != nil {
			log.Infof("Peer %s sent packet with invalid header", s.p)
			return nil, errors.ProtocolViolation
		}
		switch pkt.Body.Type {
		case packet.TypeAnnounce:
			// Collision detected: both peers tried to send announce
			// simultaneously.
			return newStateIdle(s.p), nil
		case packet.TypeAck:
			if err != nil {
				log.Errorf("Error processing ack from peer %s: %v", s.p, err)
				return nil, err
			}
			acked = true
		case packet.TypeReq:
			err = s.processReq(pkt)
			if err != nil {
				log.Errorf("Error processing req from peer %s: %v", s.p, err)
				return nil, err
			}
		default:
			log.Fatal("BUG: packet type validation failed.")
		}
	}
	return newStateIdle(s.p), nil
}

func (s *StateSending) sendAnnounce(id *entity.ID) error {
	pld := packet.NewPayloadAnnounce(id)
	pkt := packet.New(packet.TypeAnnounce, s.p.User.ID(), pld, s.p.owner.Signer)
	err := s.p.conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt, s.p, err)
		return err
	}
	return nil
}

func (s *StateSending) processReq(pkt *packet.Packet) error {
	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of a req '%s': %v", pkt, err)
		return errors.ProtocolViolation
	}
	r, ok := (i).(*packet.PayloadReq)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	var e entity.Entity = s.outgoingEntity
	if r.ID != *e.ID() {
		e, err = s.p.owner.Storage.GetEntity(&r.ID)
		if err != nil {
			log.Errorf("Failed to get requested entity from the DB: %v", err)
			return err
		}
	}
	err = s.sendEntity(e)
	if err != nil {
		log.Infof("Failed to send outgoing entity to '%s': %v", s.p, err)
		return err
	}
	return nil
}

func (s *StateSending) processAck(pkt *packet.Packet) error {
	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of packet '%s': %v", pkt, err)
		return errors.ProtocolViolation
	}
	_, ok := (i).(*packet.PayloadAck)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	return nil
}

func (s *StateSending) sendEntity(e entity.Entity) error {
	log.Debugf("DEBUG: entity type is %d", e.Type())
	var t packet.Type
	switch e.Type() {
	case entity.TypeMessage:
		t = packet.TypeMessage
	case entity.TypeOperation:
		t = packet.TypeOperation
	case entity.TypeUser:
		t = packet.TypeUser
	default:
		log.Fatal("BUG: unknown entity type.")
	}
	log.Debugf("DEBUG: packet type is %s", t)
	pkt := packet.New(t, s.p.User.ID(), e, s.p.owner.Signer)
	err := s.p.conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt, s.p, err)
		return err
	}
	return nil
}

func (s *StateSending) Name() string {
	return "Sending"
}

func (s *StateSending) ID() StateID {
	return StateIDSending
}
