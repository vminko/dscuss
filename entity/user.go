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
	"encoding/json"
	"fmt"
	"time"
	"vminko.org/dscuss/crypto"
	"vminko.org/dscuss/log"
)

// User is an UnsignedUser with a signature. Like UnsignedUser, it also
// identifies and describes a user. But in contrast to UnsignedUser, it's
// suitable for sending to the network.
type User struct {
	UnsignedUser
	Sig crypto.Signature
}

// emergeUser creates a new user entity. It should only be called when
// signature is not known yet.  Signature will be created using the provided
// signer.
func EmergeUser(
	nickname string,
	info string,
	proof crypto.ProofOfWork,
	regdate time.Time,
	signer *crypto.Signer,
) *User {
	uu := newUnsignedUser(nickname, info, signer.Public(), proof, regdate)
	juser, err := json.Marshal(uu)
	if err != nil {
		log.Fatal("Can't marshal UnsignedUser: " + err.Error())
	}
	sig, err := signer.Sign(juser)
	if err != nil {
		log.Fatal("Can't sign JSON-encoded user: " + err.Error())
	}

	return &User{UnsignedUser: *uu, Sig: sig}
}

func NewUser(
	nickname string,
	info string,
	pubkey *crypto.PublicKey,
	proof crypto.ProofOfWork,
	regdate time.Time,
	sig crypto.Signature,
) *User {
	uu := newUnsignedUser(nickname, info, pubkey, proof, regdate)
	return &User{UnsignedUser: *uu, Sig: sig}
}

func (u *User) String() string {
	userStr, err := json.Marshal(u)
	if err != nil {
		log.Errorf("Can't marshal the user %s: %v", u.Nickname, err)
		return "[Failed to marshal the user]"
	}
	return string(userStr)
}

func (u *User) Nickname() string {
	return u.UnsignedUser.Nickname
}

func (u *User) ShortID() string {
	return u.UnsignedUser.Entity.ID.Shorten()
}

func (u *User) Desc() string {
	// TBD: add subscriptions?
	return fmt.Sprintf("(%s)", u.UnsignedUser.Nickname)
}

// UnsignedUser identifies and describes a user. UnsignedUser has to be signed
// (converted to the User) before sending to the network.
// Implements Entity interface.
type UnsignedUser struct {
	Entity
	PubKey   crypto.PublicKey
	Proof    crypto.ProofOfWork
	Nickname string
	Info     string
	RegDate  time.Time
}

func newUnsignedUser(
	nickname string,
	info string,
	pubkey *crypto.PublicKey,
	proof crypto.ProofOfWork,
	regdate time.Time,
) *UnsignedUser {
	return &UnsignedUser{
		Entity: Entity{
			Type: TypeUser,
			ID:   NewID(pubkey.EncodeToDER()),
		},
		PubKey:   *pubkey,
		Proof:    proof,
		Nickname: nickname,
		Info:     info,
		RegDate:  regdate,
	}
}
