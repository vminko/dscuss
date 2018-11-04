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

const (
	// Encapsulates a user entity.
	MaxPendingEntitiesNum int = 100
)

// StateReceiving implements the entity receiving protocol.
type StateReceiving struct {
	p               *Peer
	initialPacket   *packet.Packet
	pendingEntities []entity.Entity
	requestedEntity *entity.ID
}

func newStateReceiving(p *Peer, pckt *packet.Packet) *StateReceiving {
	return &StateReceiving{p, pckt, nil, nil}
}

func (s *StateReceiving) banUser(id *entity.ID, comment string) {
	o, err := entity.EmergeOperation(
		entity.OperationTypeBanUser,
		entity.OperationReasonProtocolViolation,
		comment,
		s.p.owner.User.ID(),
		id,
		s.p.owner.Signer,
	)
	if err != nil {
		log.Fatalf("Failed to create a new operation: %v", err)
	}
	err = s.p.storage.PutEntity(o, nil)
	if err != nil {
		//
		log.Errorf("Failed to put entity %s in storage: %v", o.Desc(), err)
	}
}

func (s *StateReceiving) getPendingEntity(id *entity.ID) entity.Entity {
	for _, e := range s.pendingEntities {
		if *e.ID() == *id {
			return e
		}
	}
	return nil
}

func (s *StateReceiving) getPendingUser(id *entity.ID) *entity.User {
	e := s.getPendingEntity(id)
	if e == nil {
		return nil
	}
	u, ok := (e).(*entity.User)
	if !ok {
		log.Warningf("Found entity with requested ID %s, but it's not a user (%T)",
			id.Shorten(), e)
		return nil
	}
	return u
}

func (s *StateReceiving) getPendingMessage(id *entity.ID) *entity.Message {
	e := s.getPendingEntity(id)
	if e == nil {
		return nil
	}
	m, ok := (e).(*entity.Message)
	if !ok {
		log.Warningf("Found entity with requested ID %s, but it's not a message (%T)",
			id.Shorten(), e)
		return nil
	}
	return m
}

func (s *StateReceiving) perform() (nextState State, err error) {
	log.Debugf("Peer %s is performing state %s", s.p.Desc(), s.Name())

	a, err := s.processAnnounce()
	if err != nil {
		return nil, err
	}
	has, err := s.p.storage.HasEntity(&a.ID)
	if err != nil {
		log.Fatalf("Got unexpected error while looking for a message in the DB: %v", err)
	}

	neededID := &a.ID
	allMatched := has
	for !allMatched {
		err = s.sendReq(neededID)
		if err != nil {
			log.Errorf("Failed to request entity %s: %v", neededID.Shorten(), err)
			return nil, err
		}
		e, err := s.readEntity()
		if err != nil {
			log.Infof("Failed to receive requested entity %s: %v",
				neededID.Shorten(), err)
			return nil, err
		}
		s.pendingEntities = append(s.pendingEntities, e)
		err = s.checkPendingEntities()
		if err == nil {
			for _, e := range s.pendingEntities {
				err = s.p.storage.PutEntity(e, s.p.outEntityChan)
				if err != nil {
					log.Fatalf("Failed to put entity %s into the DB: %v",
						e.Desc(), err)
				}
			}
			allMatched = true
		} else {
			switch e := err.(type) {
			case *needIDError:
				if len(s.pendingEntities) < MaxPendingEntitiesNum {
					neededID = e.ID
				} else {
					// TBD: limit max depth of threads, then
					// ban a.AuthorID here
					return nil, err
				}
			case *banSenderError:
				s.banUser(s.p.User.ID(), e.Comment)
				return nil, err
			case *banIDError:
				s.banUser(e.ID, e.Comment)
				allMatched = true
			case *bannedError:
				// TBD: check rate of banned entities
				allMatched = true
			case *skipError:
				allMatched = true
			default:
				log.Fatalf("BUG: unexpected result type %T.", err)
			}
		}
	}
	err = s.sendAck()
	if err != nil {
		log.Errorf("Failed to send ack for %s: %v", a.ID.Shorten(), err)
		return nil, err
	}
	return newStateIdle(s.p), nil
}

func (s *StateReceiving) sendReq(id *entity.ID) error {
	pld := packet.NewPayloadReq(id)
	pkt := packet.New(packet.TypeReq, s.p.User.ID(), pld, s.p.owner.Signer)
	err := s.p.conn.Write(pkt)
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
	err := s.p.conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt.Desc(), s.p.Desc(), err)
		return err
	}
	return nil
}

func (s *StateReceiving) processAnnounce() (*packet.PayloadAnnounce, error) {
	pkt := s.initialPacket
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

func (s *StateReceiving) readEntity() (entity.Entity, error) {
	pkt, err := s.p.conn.Read()
	if err != nil {
		log.Errorf("Error receiving packet from the peer %s: %v", s.p.Desc(), err)
		return nil, err
	}
	if !pkt.VerifySig(&s.p.User.PubKey) {
		log.Infof("Peer %s sent a packet with invalid signature", s.p.Desc())
		return nil, errors.ProtocolViolation
	}

	verifyType := func(t packet.Type) bool {
		return t == packet.TypeUser || t == packet.TypeMessage || t == packet.TypeOperation
	}

	if pkt.VerifyHeaderFull(verifyType, s.p.owner.User.ID()) != nil {
		log.Infof("Peer %s sent packet with invalid header", s.p.Desc())
		return nil, errors.ProtocolViolation
	}
	i, err := pkt.DecodePayload()
	if err != nil {
		log.Infof("Failed to decode payload of packet '%s': %v", pkt.Desc(), err)
		return nil, errors.Parsing
	}
	e, ok := (i).(entity.Entity)
	if !ok {
		log.Fatal("BUG: payload is not entity, when packet type asserts that it is.")
	}
	if *e.ID() != *s.requestedEntity {
		log.Infof("Peer %s sent an entity, which was not requested", s.p.Desc())
		return nil, errors.ProtocolViolation
	}

	return e, nil
}

func (s *StateReceiving) checkPendingEntities() error {
	e := s.pendingEntities[0]
	_, ok := (e).(*entity.User)
	if ok {
		log.Infof("Peer %s advertised a user entity", s.p.Desc())
		return &banSenderError{"peer advertised a user entity " + e.ID().Shorten()}
	}
	for _, ent := range s.pendingEntities {
		var err error
		switch e := ent.(type) {
		case *entity.User:
			err = s.checkUser(e)
		case *entity.Message:
			err = s.checkMessage(e)
		case *entity.Operation:
			err = s.checkOperation(e)
		default:
			log.Fatalf("BUG: unexpected type of entity: %T", e)
		}
		if err != nil {
			log.Debugf("unmatched condition: %v", err)
			return err
		}
	}
	return nil
}

func (s *StateReceiving) checkUser(u *entity.User) error {
	if !u.IsValid() {
		log.Infof("Peer %s sent malformed User entity", s.p.Desc())
		return &banSenderError{"peer sent malformed user " + u.ID().Shorten()}
	}
	isBanned, err := s.p.owner.View.IsUserBanned(u.ID())
	if err != nil {
		log.Fatalf("Failed check whether %s is banned: %v", u.ID().Shorten(), err)
	}
	if isBanned {
		return &bannedError{}
	}
	return nil
}

func (s *StateReceiving) checkMessage(m *entity.Message) error {
	if !m.IsUnsignedPartValid() {
		log.Infof("Peer %s sent malformed Message entity", s.p.Desc())
		return &banSenderError{"peer sent malformed message " + m.ID().Shorten()}
	}
	isBanned, err := s.p.owner.View.IsUserBanned(&m.AuthorID)
	if err != nil {
		log.Fatalf("Failed check whether %s is banned: %v", m.AuthorID.Shorten(), err)
	}
	if isBanned {
		return &bannedError{}
	}
	u := s.getPendingUser(&m.AuthorID)
	if u == nil {
		var err error
		u, err = s.p.storage.GetUser(&m.AuthorID)
		if err == errors.NoSuchEntity {
			log.Debugf("Need user ID (%s) - author of the message %s",
				m.AuthorID.Shorten(), m.ID().Shorten())
			return &needIDError{&m.AuthorID}
		} else if err != nil {
			log.Fatalf("Unexpected error occurred while getting user from the DB: %v", err)
		}
	}
	if !m.IsSigValid(&u.PubKey) {
		log.Infof("Peer %s sent Message entity with invalid sig", s.p.Desc())
		comment := "peer sent message " + m.ID().Shorten() + " with invalid signature"
		return &banSenderError{comment}
	}
	// TBD: check rate of messages posted by u
	if m.ParentID == entity.ZeroID {
		if !s.p.owner.Subs.Covers(m.Topic) {
			log.Infof("Peer %s sent unsolicited Message entity", s.p.Desc())
			return &banSenderError{"peer sent unsolicited message " + m.ID().Shorten()}
		}
	} else {
		has, err := s.p.storage.HasMessage(&m.ParentID)
		if err != nil {
			log.Fatalf("Unexpected error while looking for a message in the DB: %v", err)
		}
		if !has && s.getPendingMessage(&m.ParentID) == nil {
			return &needIDError{&m.ParentID}
		}
	}
	return nil
}

func (s *StateReceiving) checkOperation(o *entity.Operation) error {
	if !o.IsUnsignedPartValid() {
		log.Infof("Peer %s sent malformed Operation entity", s.p.Desc())
		return &banSenderError{"peer sent malformed operation " + o.ID().Shorten()}
	}
	isBanned, err := s.p.owner.View.IsUserBanned(&o.AuthorID)
	if err != nil {
		log.Fatalf("Failed check whether %s is banned: %v", o.AuthorID.Shorten(), err)
	}
	if isBanned {
		return &bannedError{}
	}
	u := s.getPendingUser(&o.AuthorID)
	if u == nil {
		var err error
		u, err = s.p.storage.GetUser(&o.AuthorID)
		if err == errors.NoSuchEntity {
			log.Debugf("Need user ID (%s) - author of the operation %s",
				o.AuthorID.Shorten(), o.ID().Shorten())
			return &needIDError{&o.AuthorID}
		} else if err != nil {
			log.Fatalf("Unexpected error occurred while getting user %s: %v",
				o.AuthorID.Shorten(), err)
		}
	}
	if !o.IsSigValid(&u.PubKey) {
		log.Infof("Peer %s sent Operation with invalid sig", s.p.Desc())
		comment := "peer sent operation " + o.ID().Shorten() + " with invalid signature"
		return &banSenderError{comment}
	}
	// TBD: check rate of operations performed by u
	if s.getPendingEntity(&o.ObjectID) == nil {
		has, err := s.p.storage.HasEntity(&o.ObjectID)
		if err != nil {
			log.Fatalf("Unexpected error occurred while checking for entity %s: %v",
				o.ObjectID.Shorten(), err)
		}
		if !has {
			if o.OperationType() == entity.OperationTypeBanUser {
				log.Debugf("Skipping operation on unknown user (%s)",
					o.ObjectID.Shorten())
				return &skipError{}
			} else {
				log.Debugf("Need message ID (%s) - object of the operation %s",
					o.ObjectID.Shorten(), o.ID().Shorten())
				return &needIDError{&o.ObjectID}
			}
		}
	}
	return nil
}

func (s *StateReceiving) Name() string {
	return "Receiving"
}

func (s *StateReceiving) ID() StateID {
	return StateIDReceiving
}
