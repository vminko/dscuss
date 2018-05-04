/*
This file is part of Dscuss.
Copyright (C) 2017-2018  Vitaly Minko

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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Type int
type ID [32]byte

const (
	// User registers, post messages and performs operations.
	TypeUser Type = iota
	// Some information published by a user.
	TypeMessage
	// An action performed on a user or a message.
	TypeOperation
)

// Entity is a logical unit of data for communication between peers.
type Entity struct {
	Type Type
	ID   ID
}

func (e *Entity) Description() string {
	return fmt.Sprintf("entity type %d, id [%x]", e.Type, e.ID)
}

type EntityProvider interface {
	AddEntityConsumer(ec EntityConsumer)
	RemoveEntityConsumer(ec EntityConsumer)
}

type EntityConsumer interface {
	EntityReceived(e *Entity)
}

type EntityStorage interface {
	PutEntity(e *Entity) error
	GetEntity(id ID) (*Entity, error)
	//GetRootMessages(mi MessageIterator)
	//GetMessageReplies(id ID, mi MessageIterator)
}

func NewID(data []byte) ID {
	return sha256.Sum256(data)
}

func (i ID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + base64.StdEncoding.EncodeToString(i[:]) + `"`), nil
}

func (i *ID) UnmarshalJSON(b []byte) error {
	res, err := base64.StdEncoding.DecodeString(string(b))
	copy(i[:], res[:])
	return err
}