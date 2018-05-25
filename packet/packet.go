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
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
)

type Type string

type Body struct {
	Type    Type            `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

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
	TypeOper Type = "oper"
	// Used for introducing users during handshake.
	TypeHello Type = "hello"
	// Used for advertising new entities.
	TypeAnnounce Type = "ann"
	// Acknowledgment for an announcement.
	TypeAck Type = "ack"
	// Request for an entity.
	TypeReq Type = "req"
)

func NewPacket(t Type, p interface{}, s *crypto.Signer) *Packet {
	jp, err := json.Marshal(p)
	if err != nil {
		log.Fatal("Can't marshal packet payload: " + err.Error())
	}
	b := &Body{Type: t, Payload: jp}
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
	case TypeHello:
		pld = new(PayloadHello)
	default:
		log.Warning("Unknown payload type: " + string(p.Body.Type))
		return nil, ErrWrongType
	}
	err := json.Unmarshal(p.Body.Payload, pld)
	if err != nil {
		log.Warning("Can't unmarshal PayloadUser: " + err.Error())
		return nil, ErrMalformedPayload
	}
	return pld, nil
}

func (p *Packet) Verify(pubKey *crypto.PublicKey) bool {
	jbody, err := json.Marshal(p.Body)
	if err != nil {
		log.Fatal("Can't marshal packet body: " + err.Error())
	}
	return pubKey.Verify(jbody, p.Sig)
}

func (p *Packet) Desc() string {
	return fmt.Sprintf("type  %s", p.Body.Type)
}
