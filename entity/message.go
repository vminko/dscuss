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

// You have either Topic or ParentID, never both.
type MessageContent struct {
	Subject     string
	Text        string
	AuthorID    ID
	ParentID    ID
	DateWritten time.Time
	Topic       subs.Topic
}

type StoredMessage struct {
	M      *Message
	Stored time.Time
}

const (
	MaxMessageSubjectLen = 128
	MaxMessageTextLen    = 1024
	MaxMessageDepth      = 1024
	MinMessagePostDelay  = 1 * time.Minute
)

var lastMsgTimestamp time.Time

// EmergeMessage creates a new message. It should be called when owner wants to
// post a new message. Signature will be created using the provided signer.
func EmergeMessage(
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	signer *crypto.Signer,
	topic subs.Topic,
) (*Message, error) {
	if time.Since(lastMsgTimestamp) < MinMessagePostDelay {
		log.Errorf("Attempt to create a message violating the limit of the message post rate")
		return nil, errors.MsgPostRateErr
	} else {
		lastMsgTimestamp = time.Now()
	}
	um := newUnsignedMessage(subject, text, authorID, parentID, time.Now(), topic)
	if !um.isValid() {
		return nil, errors.WrongArguments
	}
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

// NewMessage composes a new message entity object from the specified data.
func NewMessage(
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	dateWritten time.Time,
	sig crypto.Signature,
	topic subs.Topic,
) (*Message, error) {
	um := newUnsignedMessage(subject, text, authorID, parentID, dateWritten, topic)
	if !um.isValid() {
		return nil, errors.WrongArguments
	}
	return &Message{UnsignedMessage: *um, Sig: sig}, nil
}

func (m *Message) Copy() *Message {
	res := *m
	if m.Topic != nil {
		res.Topic = m.Topic.Copy()
	}
	return &res
}

func (um *UnsignedMessage) ShortID() string {
	return um.Descriptor.ID.Shorten()
}

func (um *UnsignedMessage) Type() Type {
	return um.Descriptor.Type
}

func (um *UnsignedMessage) ID() *ID {
	return &um.Descriptor.ID
}

func (um *UnsignedMessage) String() string {
	shortText := strings.Replace(dstrings.Truncate(um.Text, 24), "\n", " ", -1)
	return fmt.Sprintf("%s (%s)", um.ShortID(), shortText)
}

func (um *UnsignedMessage) isValid() bool {
	correctID := um.MessageContent.ToID()
	if um.Descriptor.ID != *correctID {
		log.Debugf("Message %s has invalid ID", um)
		return false
	}
	if len(um.Subject) > MaxMessageSubjectLen {
		log.Debugf("Message %s has too long subject (%d)", um, len(um.Subject))
		return false
	}
	if len(um.Text) > MaxMessageTextLen {
		log.Debugf("Message %s has too long text (%d)", um, len(um.Text))
		return false
	}
	if um.Subject == "" || um.Text == "" {
		log.Debugf("Message %s has empty subject or text", um)
		return false
	}
	if um.Topic == nil && um.ParentID.IsZero() {
		log.Debugf("Message %s is a thread with nil topic", um)
		return false
	}
	if um.Topic != nil && !um.ParentID.IsZero() {
		log.Debugf("Message %s is a reply with non-nil topic", um)
		return false
	}
	return true
}

func (m *Message) IsUnsignedPartValid() bool {
	return m.UnsignedMessage.isValid()
}

func (m *Message) IsSigValid(pubKey *crypto.PublicKey) bool {
	jmsg, err := json.Marshal(&m.UnsignedMessage)
	if err != nil {
		log.Fatal("Can't marshal UnsignedMessage: " + err.Error())
	}
	res := pubKey.Verify(jmsg, m.Sig)
	if !res {
		log.Debugf("Message %s has invalid signature", m)
	}
	return res
}

func (m *Message) IsValid(pubKey *crypto.PublicKey) bool {
	return m.IsUnsignedPartValid() && m.IsSigValid(pubKey)
}

func (m *Message) IsReply() bool {
	return !m.ParentID.IsZero()
}

func newMessageContent(
	subject string,
	text string,
	authorID *ID,
	parentID *ID,
	dateWritten time.Time,
	topic subs.Topic,
) *MessageContent {
	mc := &MessageContent{
		Subject:     subject,
		Text:        text,
		AuthorID:    *authorID,
		ParentID:    *parentID,
		DateWritten: dateWritten,
	}
	if topic != nil {
		mc.Topic = topic.Copy()
	}
	return mc
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
