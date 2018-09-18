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

type StateReceiving struct {
	p               *Peer
	pendingPackets  []*packet.Packet
	requestedEntity *entity.ID
}

func newStateReceiving(p *Peer, pckt *packet.Packet) *StateReceiving {
	return &StateReceiving{p, []*packet.Packet{pckt}, nil}
}

func (s *StateReceiving) perform() (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", s.p.Desc(), s.Name())

	a, err := s.processAnnounce()
	if err != nil {
		return nil, err
	}
	has, err := s.p.storage.HasMessage(&a.ID)
	if err != nil {
		log.Errorf("Got unexpected error while looking for a message in the DB: %v", err)
		return nil, errors.Database
	}
	if has {
		err = s.sendAck()
		if err != nil {
			log.Errorf("Failed to send ack for %s: %v", a.ID.Shorten(), err)
			return nil, err
		}
	} else {
		err = s.sendReq(&a.ID)
		if err != nil {
			log.Errorf("Failed to request entity %s: %v", a.ID.Shorten(), err)
			return nil, err
		}
		err = s.readAndProcessMessage()
		if err != nil {
			log.Infof("Failed to receive requested entity %s: %v", a.ID.Shorten(), err)
			return nil, err
		}
	}
	return newStateIdle(s.p), nil
}

func (s *StateReceiving) sendReq(id *entity.ID) error {
	pld := packet.NewPayloadReq(id)
	pkt := packet.New(packet.TypeReq, s.p.User.ID(), pld, s.p.owner.Signer)
	err := s.p.Conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt.Desc(), s.p.Desc(), err)
		return err
	}
	s.requestedEntity = id
	return nil
}

func (s *StateReceiving) sendAck() error {
	pld := packet.NewPayloadAck()
	pkt := packet.New(packet.TypeAck, s.p.User.ID(), pld, s.p.owner.Signer)
	err := s.p.Conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateReceiving) processAnnounce() (*packet.PayloadAnnounce, error) {
	pkt := s.pendingPackets[0]
	if !pkt.VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature: %s", s.p.Desc())
		return nil, errors.ProtocolViolation
	}
	err := pkt.VerifyHeader(packet.TypeAnnounce, s.p.owner.User.ID())
	if err != nil {
		log.Infof("Peer %s sent packet with invalid header: %v", s.p.Desc(), err)
		return nil, errors.ProtocolViolation
	}

	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of announce '%s': %v", pkt.Desc(), err)
		return nil, errors.Parsing
	}
	a, ok := (i).(*packet.PayloadAnnounce)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}

	return a, nil
}

func (s *StateReceiving) readAndProcessMessage() error {
	pkt, err := s.p.Conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return err
	}
	if !pkt.VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature", s.p.Desc())
		return errors.ProtocolViolation
	}
	if pkt.VerifyHeader(packet.TypeMessage, s.p.owner.User.ID()) != nil {
		log.Infof("Peer %s sent packet with invalid header", s.p.Desc())
		return errors.ProtocolViolation
	}

	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of packet '%s': %v", pkt.Desc(), err)
		return errors.Parsing
	}
	m, ok := (i).(*entity.Message)
	if !ok {
		log.Fatal("BUG: packet type does not match type of successfully decoded payload.")
	}
	if !m.IsValid(&s.p.User.PubKey) {
		log.Infof("Peer %s sent malformed Message entity", s.p.Desc())
		return errors.ProtocolViolation
	}
	if m.Descriptor.ID != *s.requestedEntity {
		log.Infof("Peer %s sent unsolicited Message entity", s.p.Desc())
		return errors.ProtocolViolation
	}

	if !s.p.Subs.Covers(m.Topic) {
		log.Infof("Peer %s sent unsolicited Message entity", s.p.Desc())
		return errors.ProtocolViolation
	}

	err = s.p.storage.PutMessage(m, s.p.outEntityChan)
	if err != nil {
		log.Errorf("Failed to put user into the DB: %v", err)
		return errors.Database
	}

	return nil
}

func (s *StateReceiving) Name() string {
	return "Receiving"
}

func (s *StateReceiving) ID() StateID {
	return StateIDReceiving
}
