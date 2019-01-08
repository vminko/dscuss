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
package packet

import (
	"encoding/json"
	"fmt"
	"time"
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

type Type string

// Time and ReceiverID protect from replay attack.
type Body struct {
	Type         Type            `json:"type"`
	ReceiverID   entity.ID       `json:"receiver_id"`   // Id of the user this packet is designated for.
	DateComposed time.Time       `json:"date_composed"` // Date and time when the payload was composed.
	Payload      json.RawMessage `json:"payload"`
}

// Packet is a unit of raw data for communication between peers.
type Packet struct {
	Body Body             `json:"body"`
	Sig  crypto.Signature `json:"sig"`
}

const (
	// Encapsulates a user entity.
	TypeUser Type = "user"
	// Encapsulates a message entity.
	TypeMessage Type = "msg"
	// Encapsulates an operation entity.
	TypeOperation Type = "oper"
	// Used for introducing users during handshake.
	TypeHello Type = "hello"
	// Used for advertising new entities.
	TypeAnnounce Type = "ann"
	// Acknowledgment for an announcement.
	TypeAck Type = "ack"
	// Request for an entity.
	TypeReq Type = "req"
	// Done indicated that a complex process (like syncing) is over.
	TypeDone Type = "done"
)

func New(t Type, rcv *entity.ID, pld interface{}, s *crypto.Signer) *Packet {
	jp, err := json.Marshal(pld)
	if err != nil {
		log.Fatal("Can't marshal packet payload: " + err.Error())
	}
	b := &Body{Type: t, DateComposed: time.Now(), Payload: jp}
	copy(b.ReceiverID[:], rcv[:])

	jb, err := json.Marshal(b)
	if err != nil {
		log.Fatal("Can't marshal packet body: " + err.Error())
	}
	sig, err := s.Sign(jb)
	if err != nil {
		log.Fatal("Can't sign JSON-encoded body: " + err.Error())
	}
	return &Packet{Body: *b, Sig: sig}
}

func (p *Packet) DecodePayload() (interface{}, error) {
	var pld interface{}
	switch p.Body.Type {
	case TypeUser:
		pld = new(entity.User)
	case TypeMessage:
		pld = new(entity.Message)
	case TypeOperation:
		pld = new(entity.Operation)
	case TypeHello:
		pld = new(PayloadHello)
	case TypeAnnounce:
		pld = new(PayloadAnnounce)
	case TypeReq:
		pld = new(PayloadReq)
	case TypeAck:
		pld = new(PayloadAck)
	case TypeDone:
		pld = new(PayloadDone)
	default:
		log.Error("Unknown payload type: " + string(p.Body.Type))
		return nil, errors.WrongPacketType
	}
	err := json.Unmarshal(p.Body.Payload, pld)
	if err != nil {
		log.Error("Can't unmarshal Payload: " + err.Error())
		log.Debug("Dumping Payload: " + string(p.Body.Payload))
		return nil, errors.MalformedPayload
	}
	return pld, nil
}

func (p *Packet) VerifySig(pubKey *crypto.PublicKey) bool {
	jbody, err := json.Marshal(p.Body)
	if err != nil {
		log.Fatal("Can't marshal packet body: " + err.Error())
	}
	return pubKey.Verify(jbody, p.Sig)
}

func (p *Packet) VerifyHeader(t Type, rcv *entity.ID) error {
	return p.VerifyHeaderFull(func(ft Type) bool { return ft == t }, rcv)
}

func (p *Packet) VerifyHeaderFull(f func(Type) bool, rcv *entity.ID) error {
	if !f(p.Body.Type) {
		log.Infof("Got packet of unexpected type %d", p.Body.Type)
		return errors.ProtocolViolation
	}
	if p.Body.ReceiverID != *rcv {
		log.Errorf("Got packet with wrong receiver ID: '%s'.", p.Body.ReceiverID.String())
		return errors.ProtocolViolation
	}
	const MaxTimeDiscrepancy = 3 * time.Minute
	if time.Since(p.Body.DateComposed) > MaxTimeDiscrepancy {
		log.Errorf("Got packet with obsolete timestamp: '%s'.",
			p.Body.DateComposed.Format(time.RFC3339))
		return errors.ProtocolViolation
	}
	return nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("type  %s", p.Body.Type)
}

func (p *Packet) Dump() string {
	str, err := json.Marshal(p)
	if err != nil {
		return "[error: " + err.Error() + "]"
	}
	return fmt.Sprintf("%s", str)
}
