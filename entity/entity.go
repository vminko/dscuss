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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

type Type int

const (
	// User registers, post messages and performs operations.
	TypeUser Type = iota
	// Some information published by a user.
	TypeMessage
	// An action performed on a user or a message.
	TypeOperation
)

type ID [32]byte

// Entity is a logical unit of data for communication between peers.
type Entity interface {
	Type() Type
	ID() *ID
	Desc() string
}

type EntityProvider interface {
	AttachObserver(c chan<- Entity)
	DetachObserver(c chan<- Entity)
}

type Descriptor struct {
	Type Type
	ID   ID
}

var ZeroID ID

func NewID(data []byte) ID {
	return sha256.Sum256(data)
}

func (i ID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + base64.StdEncoding.EncodeToString(i[:]) + `"`), nil
}

func (i *ID) UnmarshalJSON(b []byte) error {
	trimmed := bytes.Trim(b, "\"")
	res, err := base64.StdEncoding.DecodeString(string(trimmed))
	copy(i[:], res[:])
	return err
}

func (i *ID) String() string {
	return hex.EncodeToString(i[:])
}

func (i *ID) Shorten() string {
	idStr := i.String()
	return idStr[:8]
}
