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
	"time"
)

// UnsignedUser identifies and describes a user. UnsignedUser has to be signed
// (converted to the User) before sending to the network.
// Implements Entity interface.
type UnsignedUser struct {
	Entity
	PubKey   PublicKey
	Proof    ProofOfWork
	Nickname string
	Info     string
	RegDate  time.Time
}

func newUnsignedUser(
	nickname string,
	info string,
	pubkey *PublicKey,
	proof ProofOfWork,
	regdate time.Time) *UnsignedUser {

	return &UnsignedUser{
		Entity: Entity{
			Type: EntityTypeUser,
			ID:   newEntityID(pubkey.encodeToDER()),
		},
		PubKey:   *pubkey,
		Proof:    proof,
		Nickname: nickname,
		Info:     info,
		RegDate:  regdate,
	}
}

func (u *UnsignedUser) Description() string {
	return u.Nickname
}
