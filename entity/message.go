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

package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/log"
)

// Message is some text information published by a user.
type Message struct {
	Descriptor
	Subject     string
	Text        string
	AuthorID    ID
	ParentID    ID
	DateWritten time.Time
	// TBD: topic
	Sig crypto.Signature
}

// EmergeMessage creates a new message entity. It should only be called when
// signature is not known yet.  Signature will be created using the provided
// signer.
func EmergeMessage(
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	signer *crypto.Signer,
) *Message {
	m := &Message{
		Descriptor: Descriptor{
			Type: TypeMessage,
			ID:   ZeroID,
		},
		Subject:     subject,
		Text:        text,
		AuthorID:    *authorID,
		ParentID:    *parentID,
		DateWritten: time.Now(),
	}

	// Two more properties to go: ID and signature.
	// Calculate msg ID
	noidJMsg, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Can't marshal unsigned Message without ID: " + err.Error())
	}
	m.Descriptor.ID = NewID(noidJMsg)

	// Calculate msg signature
	fullJMsg, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Can't marshal unsigned Message: " + err.Error())
	}
	sig, err := signer.Sign(fullJMsg)
	if err != nil {
		log.Fatal("Can't sign JSON-encoded Message entity: " + err.Error())
	}
	m.Sig = sig

	return m
}

func NewMessage(
	id *ID,
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	dateWritten time.Time,
	sig crypto.Signature,
) *Message {
	return &Message{
		Descriptor: Descriptor{
			Type: TypeMessage,
			ID:   *id,
		},
		Subject:     subject,
		Text:        text,
		AuthorID:    *authorID,
		ParentID:    *parentID,
		DateWritten: dateWritten,
		Sig:         sig,
	}
}

func (m *Message) Copy() *Message {
	res := *m
	return &res
}

func (m *Message) VerifySig(pubKey *crypto.PublicKey) bool {
	jmsg, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Can't marshal message: " + err.Error())
	}
	return pubKey.Verify(jmsg, m.Sig)
}

func (m *Message) VerifyID() bool {
	tmp := m.Copy()
	tmp.Descriptor.ID = ZeroID
	jtmp, err := json.Marshal(tmp)
	if err != nil {
		log.Fatal("Can't marshal temporary Message with zero ID: " + err.Error())
	}
	correctID := NewID(jtmp)
	return m.Descriptor.ID == correctID
}

func (m *Message) ShortID() string {
	return m.Descriptor.ID.Shorten()
}

func (m *Message) Type() Type {
	return m.Descriptor.Type
}

func (m *Message) ID() *ID {
	return &m.Descriptor.ID
}

func (m *Message) Desc() string {
	return fmt.Sprintf("%s (%s)", m.ShortID(), strings.Trim(m.Text[:24], "\\n\\r"))
}
