/*
This file is part of Dscuss.
Copyright (C) 2017  Vitaly Minko

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

package dscuss

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type EntityType int

const (
	// User registers, post messages and performs operations.
	EntityTypeUser EntityType = iota
	// Some information published by a user.
	EntityTypeMessage
	// An action performed on a user or a message.
	EntityTypeOperation
)

type EntityID [32]byte

func newEntityID(data []byte) EntityID {
	return sha256.Sum256(data)
}

func (eid EntityID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + base64.StdEncoding.EncodeToString(eid[:]) + `"`), nil
}

func (eid *EntityID) UnmarshalJSON(b []byte) error {
	res, err := base64.StdEncoding.DecodeString(string(b))
	copy(eid[:], res[:])
	return err
}

// Entity is a logical unit of data for communication between peers.
type Entity struct {
	Type EntityType
	ID   EntityID
}

/*type Entity interface {
	Type() EntityType
	ID() EntityID
	Description() string
}*/

func (e *Entity) Description() string {
	return fmt.Sprintf("entity type %d, id [%x]", e.Type, e.ID)
}
