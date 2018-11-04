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
	"time"
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

type OperationType int
type OperationReason int

const (
	OperationTypeRemoveMessage OperationType = iota
	OperationTypeBanUser
	// TBD: OperationTypeBanUserTemporarily
	// TBD: OperationTypeEditMessageTopic
	// TBD: OperationTypeEditMessageSubject
	// TBD: OperationTypeEditMessageText
)

const (
	OperationTypeRemoveMessageStr string = "RemoveMessage"
	OperationTypeBanUserStr       string = "BanUser"
)

const (
	OperationReasonProtocolViolation OperationReason = iota
	OperationReasonSpam
	OperationReasonOfftopic
	OperationReasonAbuse
	OperationReasonDuplicate
)

const (
	OperationReasonProtocolViolationStr string = "ProtocolViolation"
	OperationReasonSpamStr              string = "SPAM"
	OperationReasonOfftopicStr          string = "Offtopic"
	OperationReasonAbuseStr             string = "Abuse"
	OperationReasonDuplicateStr         string = "Duplicate"
)

// Operation is an action performed on a user or a message.
type Operation struct {
	UnsignedOperation
	Sig crypto.Signature
}

type UnsignedOperation struct {
	Descriptor
	OperationContent
}

type OperationContent struct {
	Type          OperationType
	Reason        OperationReason
	Comment       string
	AuthorID      ID
	ObjectID      ID
	DatePerformed time.Time
}

func (ot OperationType) String() string {
	switch ot {
	case OperationTypeRemoveMessage:
		return OperationTypeRemoveMessageStr
	case OperationTypeBanUser:
		return OperationTypeBanUserStr
	default:
		return "unknown operation type"
	}
}

func (or OperationReason) String() string {
	switch or {
	case OperationReasonProtocolViolation:
		return OperationReasonProtocolViolationStr
	case OperationReasonSpam:
		return OperationReasonSpamStr
	case OperationReasonOfftopic:
		return OperationReasonOfftopicStr
	case OperationReasonAbuse:
		return OperationReasonAbuseStr
	case OperationReasonDuplicate:
		return OperationReasonDuplicateStr
	default:
		return "unknown operation reason"
	}
}

func (or *OperationReason) ParseString(s string) error {
	switch s {
	case OperationReasonProtocolViolationStr:
		*or = OperationReasonProtocolViolation
	case OperationReasonSpamStr:
		*or = OperationReasonSpam
	case OperationReasonOfftopicStr:
		*or = OperationReasonOfftopic
	case OperationReasonAbuseStr:
		*or = OperationReasonAbuse
	case OperationReasonDuplicateStr:
		*or = OperationReasonDuplicate
	default:
		return errors.Parsing
	}
	return nil
}

func (uo *UnsignedOperation) ShortID() string {
	return uo.Descriptor.ID.Shorten()
}

func (uo *UnsignedOperation) OperationType() OperationType {
	return uo.OperationContent.Type
}

func (uo *UnsignedOperation) Type() Type {
	return uo.Descriptor.Type
}

func (uo *UnsignedOperation) ID() *ID {
	return &uo.Descriptor.ID
}

func (uo *UnsignedOperation) String() string {
	return fmt.Sprintf("%s (%s performed oper type %s reason %s on %s)",
		uo.ShortID(), uo.AuthorID.Shorten(), uo.OperationType().String(),
		uo.Reason, uo.ObjectID.Shorten())
}

func (uo *UnsignedOperation) isValid() bool {
	correctID := uo.OperationContent.ToID()
	if uo.Descriptor.ID != *correctID {
		log.Debugf("Operation %s has invalid ID", uo)
		return false
	}
	t := uo.OperationContent.Type
	if t != OperationTypeRemoveMessage && t != OperationTypeBanUser {
		log.Debugf("Operation %s has invalid type %d", uo, t)
		return false
	}
	isReasonOK := uo.Reason == OperationReasonSpam || uo.Reason == OperationReasonOfftopic ||
		uo.Reason == OperationReasonAbuse || uo.Reason == OperationReasonDuplicate ||
		uo.Reason == OperationReasonProtocolViolation
	if !isReasonOK {
		log.Debugf("Operation %s has invalid reason %d", uo, uo.Reason)
		return false
	}
	if uo.AuthorID == ZeroID {
		log.Debugf("Operation %s has empty author", uo, uo.AuthorID)
		return false
	}
	if uo.ObjectID == ZeroID {
		log.Debugf("Operation %s has empty objectr", uo, uo.ObjectID)
		return false
	}
	return true
}

func (o *Operation) IsUnsignedPartValid() bool {
	return o.UnsignedOperation.isValid()
}

func (o *Operation) IsSigValid(pubKey *crypto.PublicKey) bool {
	jop, err := json.Marshal(&o.UnsignedOperation)
	if err != nil {
		log.Fatal("Can't marshal UnsignedOperation: " + err.Error())
	}
	res := pubKey.Verify(jop, o.Sig)
	if !res {
		log.Debugf("Operation %s has invalid signature", o)
	}
	return res
}

func (o *Operation) IsValid(pubKey *crypto.PublicKey) bool {
	return o.IsUnsignedPartValid() && o.IsSigValid(pubKey)
}

// EmergeOperation creates a new operation. It should be called when owner wants to
// perform a new operation. Signature will be created using the provided signer.
func EmergeOperation(
	typ OperationType,
	reason OperationReason,
	comment string,
	authorID *ID,
	objectID *ID,
	signer *crypto.Signer,
) (*Operation, error) {
	uo := newUnsignedOperation(typ, reason, comment, authorID, objectID, time.Now())
	if !uo.isValid() {
		return nil, errors.WrongArguments
	}
	joper, err := json.Marshal(uo)
	if err != nil {
		log.Fatal("Can't marshal unsigned Operation: " + err.Error())
	}
	sig, err := signer.Sign(joper)
	if err != nil {
		log.Fatal("Can't sign JSON-encoded Operation entity: " + err.Error())
	}
	return &Operation{UnsignedOperation: *uo, Sig: sig}, nil
}

// NewOperation composes a new operation entity object from the specified data.
func NewOperation(
	typ OperationType,
	reason OperationReason,
	comment string,
	authorID *ID,
	objectID *ID,
	datePerformed time.Time,
	sig crypto.Signature,
) (*Operation, error) {
	uo := newUnsignedOperation(typ, reason, comment, authorID, objectID, datePerformed)
	if !uo.isValid() {
		return nil, errors.WrongArguments
	}
	return &Operation{UnsignedOperation: *uo, Sig: sig}, nil
}

func newOperationContent(
	typ OperationType,
	reason OperationReason,
	comment string,
	authorID *ID,
	objectID *ID,
	datePerformed time.Time,
) *OperationContent {
	return &OperationContent{
		Type:          typ,
		Reason:        reason,
		Comment:       comment,
		AuthorID:      *authorID,
		ObjectID:      *objectID,
		DatePerformed: datePerformed,
	}
}

func (oc *OperationContent) ToID() *ID {
	joc, err := json.Marshal(oc)
	if err != nil {
		log.Fatal("Can't marshal OperationContent: " + err.Error())
	}
	id := NewID(joc)
	return &id
}

func newUnsignedOperation(
	typ OperationType,
	reason OperationReason,
	comment string,
	authorID *ID,
	objectID *ID,
	datePerformed time.Time,
) *UnsignedOperation {
	oc := newOperationContent(typ, reason, comment, authorID, objectID, datePerformed)
	return &UnsignedOperation{
		Descriptor: Descriptor{
			Type: TypeOperation,
			ID:   *oc.ToID(),
		},
		OperationContent: *oc,
	}
}
