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
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	dstrings "vminko.org/dscuss/strings"
	"vminko.org/dscuss/subs"
)

// Message is some text information published by a user.
type Message struct {
	UnsignedMessage
	Sig crypto.Signature
}

type UnsignedMessage struct {
	Descriptor
	MessageContent
}

type MessageContent struct {
	Subject     string
	Text        string
	AuthorID    ID
	ParentID    ID
	DateWritten time.Time
	Topic       subs.Topic
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
	topic subs.Topic,
) (*Message, error) {
	if subject == "" || text == "" || topic == nil {
		return nil, errors.WrongArguments
	}
	um := newUnsignedMessage(subject, text, authorID, parentID, time.Now(), topic)
	jmsg, err := json.Marshal(um)
	if err != nil {
		log.Fatal("Can't marshal unsigned Message: " + err.Error())
	}
	sig, err := signer.Sign(jmsg)
	if err != nil {
		log.Fatal("Can't sign JSON-encoded Message entity: " + err.Error())
	}
	return &Message{UnsignedMessage: *um, Sig: sig}, nil
}

func NewMessage(
	id *ID,
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	dateWritten time.Time,
	sig crypto.Signature,
	topic subs.Topic,
) *Message {
	um := newUnsignedMessage(subject, text, authorID, parentID, dateWritten, topic)
	return &Message{UnsignedMessage: *um, Sig: sig}
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
	shortText := strings.Replace(dstrings.Truncate(m.Text, 24), "\n", " ", -1)
	return fmt.Sprintf("%s (%s)", m.ShortID(), shortText)
}

func (m *Message) Copy() *Message {
	res := *m
	res.Topic = m.Topic.Copy()
	return &res
}

func (m *Message) IsValid(pubKey *crypto.PublicKey) bool {
	jmsg, err := json.Marshal(&m.UnsignedMessage)
	if err != nil {
		log.Fatal("Can't marshal UnsignedMessage: " + err.Error())
	}
	if !pubKey.Verify(jmsg, m.Sig) {
		log.Debugf("Message %s has invalid signature", m.Desc())
		return false
	}
	correctID := m.UnsignedMessage.MessageContent.ToID()
	if m.UnsignedMessage.Descriptor.ID != *correctID {
		log.Debugf("Message %s has invalid ID", m.Desc())
		return false
	}
	if m.Topic == nil || len(m.Topic) == 0 || m.Subject == "" || m.Text == "" {
		log.Debugf("Message %s has empty topic, subject or text", m.Desc())
		return false
	}
	return true
}

func newMessageContent(
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	dateWritten time.Time,
	topic subs.Topic,
) *MessageContent {
	return &MessageContent{
		Subject:     subject,
		Text:        text,
		AuthorID:    *authorID,
		ParentID:    *parentID,
		DateWritten: dateWritten,
		Topic:       topic.Copy(),
	}
}

func (mc *MessageContent) ToID() *ID {
	jmc, err := json.Marshal(mc)
	if err != nil {
		log.Fatal("Can't marshal MessageContent: " + err.Error())
	}
	id := NewID(jmc)
	return &id
}

func newUnsignedMessage(
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	dateWritten time.Time,
	topic subs.Topic,
) *UnsignedMessage {
	mc := newMessageContent(subject, text, authorID, parentID, dateWritten, topic)
	return &UnsignedMessage{
		Descriptor: Descriptor{
			Type: TypeMessage,
			ID:   *mc.ToID(),
		},
		MessageContent: *mc,
	}
}
