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

// User identifies and describes a user. It's suitable for sending to the network.
// Implements Entity interface.
type User struct {
	UnsignedUser
	Sig crypto.Signature
}

type UnsignedUser struct {
	Descriptor
	UserContent
}

type UserContent struct {
	PubKey   crypto.PublicKey
	Proof    crypto.ProofOfWork
	Nickname string
	Info     string
	RegDate  time.Time
}

// EmergeUser creates a new user entity. It should only be called when
// signature is not known yet.  Signature will be created using the provided
// signer.
func EmergeUser(
	nickname string,
	info string,
	proof crypto.ProofOfWork,
	signer *crypto.Signer,
) *User {
	uu := newUnsignedUser(nickname, info, signer.Public(), proof, time.Now())
	juser, err := json.Marshal(uu)
	if err != nil {
		log.Fatal("Can't marshal UnsignedUser: " + err.Error())
	}
	log.Debugf("Signing this UnsignedUser: '%s'", juser)
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
	return u.UnsignedUser.Descriptor.ID.Shorten()
}

func (u *User) Type() Type {
	return u.UnsignedUser.Descriptor.Type
}

func (u *User) ID() *ID {
	return &u.UnsignedUser.Descriptor.ID
}

func (u *User) Desc() string {
	// TBD: add subscriptions?
	return fmt.Sprintf("(%s)", u.UnsignedUser.Nickname)
}

func (u *User) VerifySig(pubKey *crypto.PublicKey) bool {
	juser, err := json.Marshal(&u.UnsignedUser)
	if err != nil {
		log.Fatal("Can't marshal UnsignedUser: " + err.Error())
	}
	log.Debugf("Verifying signature for this UnsignedUser: '%s'", juser)
	return pubKey.Verify(juser, u.Sig)
}

func (u *User) VerifyID() bool {
	correctID := u.UnsignedUser.UserContent.ToID()
	return u.UnsignedUser.Descriptor.ID == *correctID
}

func newUserContent(
	nickname string,
	info string,
	pubkey *crypto.PublicKey,
	proof crypto.ProofOfWork,
	regdate time.Time,
) *UserContent {
	return &UserContent{
		PubKey:   *pubkey,
		Proof:    proof,
		Nickname: nickname,
		Info:     info,
		RegDate:  regdate,
	}
}

func (uc *UserContent) ToID() *ID {
	id := NewID(uc.PubKey.EncodeToDER())
	return &id
}

func newUnsignedUser(
	nickname string,
	info string,
	pubkey *crypto.PublicKey,
	proof crypto.ProofOfWork,
	regdate time.Time,
) *UnsignedUser {
	uc := newUserContent(nickname, info, pubkey, proof, regdate)
	return &UnsignedUser{
		Descriptor: Descriptor{
			Type: TypeUser,
			ID:   *uc.ToID(),
		},
		UserContent: *uc,
	}
}
