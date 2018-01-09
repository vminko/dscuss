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
	"encoding/json"
	"time"
)

// User is an UnsignedUser with a signature. It also identifies and describes a
// user. But in contrast to UnsignedUser, it's suitable for sending to the
// network.
// Implements Entity interface.
type User struct {
	UnsignedUser
	Sig Signature
}

// EmergeUser creates a new user entity. It should only be called when
// signature is not known yet.  Signature will be created using the provided
// signer.
func EmergeUser(
	nickname string,
	info string,
	proof ProofOfWork,
	regdate time.Time,
	signer *Signer) (*User, error) {

	uu, err := newUnsignedUser(nickname, info, signer.public(), proof, regdate)
	if err != nil {
		Logf(ERROR, "Can't create UnsignedUser %s: %v", err)
		return nil, err
	}
	juser, err := json.Marshal(uu)
	if err != nil {
		Log(ERROR, "Can't marshal UnsignedUser: "+err.Error())
		return nil, ErrInternal
	}
	Logf(DEBUG, "Dumping Unsigned User %s", nickname)
	Log(DEBUG, string(juser))
	sig, err := signer.sign(juser)
	if err != nil {
		Log(ERROR, "Can't sign JSON-encoded user: "+err.Error())
		return nil, ErrInternal
	}

	return &User{UnsignedUser: *uu, Sig: sig}, nil
}
