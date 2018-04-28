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

package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"vminko.org/dscuss/log"
)

// Signer hides private key from caller and offers simple interface for signing
// data.
type Signer struct {
	privkey *PrivateKey
}

func NewSigner(privkey *PrivateKey) *Signer {
	return &Signer{privkey: privkey}
}

func (s *Signer) Public() *PublicKey {
	return s.privkey.Public()
}

// Sign creates a signature for the data using the Signer's private key.
func (s *Signer) Sign(data []byte) (Signature, error) {
	digest := sha256.Sum256(data)

	r, t, err := ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(s.privkey), digest[:])
	if err != nil {
		log.Errorf("Can't sign data using private key %s %v", s.privkey, err)
		return nil, ErrInternal
	}

	// Encode the signature {R, S}
	// big.Int.Bytes() will need padding in the case of leading zero bytes
	params := s.privkey.Curve.Params()
	curveOrderByteLen := params.P.BitLen() / 8
	rBytes, tBytes := r.Bytes(), t.Bytes()
	Signature := make([]byte, curveOrderByteLen*2)
	copy(Signature[curveOrderByteLen-len(rBytes):], rBytes)
	copy(Signature[curveOrderByteLen*2-len(tBytes):], tBytes)

	return Signature, nil
}
