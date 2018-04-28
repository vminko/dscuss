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

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"vminko.org/dscuss/log"
)

type PrivateKey ecdsa.PrivateKey

// Some parts copied from github.com/gtank/cryptopasta/.

// NewPrivateKey generates a random P-224 ECDSA private key.
func NewPrivateKey() (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	return (*PrivateKey)(key), err
}

// ParsePrivateKeyFromDER decodes a PEM-encoded ECDSA private key.
func ParsePrivateKeyFromDER(der []byte) (*PrivateKey, error) {
	privkey, err := x509.ParseECPrivateKey(der)
	if err != nil {
		log.Errorf("Can't parse private key %v", err)
		return nil, ErrParsing
	}
	return (*PrivateKey)(privkey), nil
}

// ParsePrivateKeyFromPEM decodes a PEM-encoded ECDSA private key.
func ParsePrivateKeyFromPEM(encodedKey []byte) (*PrivateKey, error) {
	block, encodedKey := pem.Decode(encodedKey)
	if block.Type != "EC PRIVATE KEY" {
		log.Error("Failed to find EC PRIVATE KEY in PEM data")
		return nil, ErrParsing
	}
	return ParsePrivateKeyFromDER(block.Bytes)
}

// EncodeToDER encodes an ECDSA private key to DER format.
func (key *PrivateKey) EncodeToDER() []byte {
	der, err := x509.MarshalECPrivateKey((*ecdsa.PrivateKey)(key))
	if err != nil {
		log.Fatalf("MarshalECPrivateKey failed to encode private key %v", err)
	}
	return der
}

// EncodeToPEM encodes an ECDSA private key to PEM format.
func (key *PrivateKey) EncodeToPEM() []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: key.EncodeToDER(),
		},
	)
}

// public returns the public key corresponding to the private key.
func (key *PrivateKey) Public() *PublicKey {
	var pubKey *PublicKey
	cryptoPub := (*ecdsa.PrivateKey)(key).Public()
	ecdsaPub, ok := (cryptoPub).(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("Wrong cryptoPub type: %T", cryptoPub)
	}
	pubKey = (*PublicKey)(ecdsaPub)
	return pubKey
}
