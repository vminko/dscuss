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
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

type PublicKey ecdsa.PublicKey

// Some parts copied from github.com/gtank/cryptopasta/.

// ParsePublicKeyFromDER decodes a DER-encoded ECDSA public key.
func ParsePublicKeyFromDER(encodedKey []byte) (*PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(encodedKey)
	if err != nil {
		log.Warningf("Can't parse public key %v", err)
		return nil, errors.Parsing
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Data was not an ECDSA public key")
	}

	return (*PublicKey)(ecdsaPub), nil
}

// ParsePublicKeyFromPEM decodes a PEM-encoded ECDSA public key.
func ParsePublicKeyFromPEM(encodedKey []byte) (*PublicKey, error) {
	block, encodedKey := pem.Decode(encodedKey)
	if block.Type != "EC PUBLIC KEY" {
		log.Error("Failed to find EC PUBLIC KEY in PEM data")
		return nil, errors.Parsing
	}
	return ParsePublicKeyFromDER(block.Bytes)
}

// EncodeToDER encodes an ECDSA public key to DER format.
func (key *PublicKey) EncodeToDER() []byte {
	derBytes, err := x509.MarshalPKIXPublicKey((*ecdsa.PublicKey)(key))
	if err != nil {
		log.Fatalf("MarshalPKIXPublicKey failed to encode public key: : %v", err)
	}
	return derBytes
}

// EncodeToPEM encodes an ECDSA public key to PEM format.
func (key *PublicKey) EncodeToPEM() []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "EC PUBLIC KEY",
			Bytes: key.EncodeToDER(),
		},
	)
}

// MarshalJSON returns the JSON encoded key.
func (key *PublicKey) MarshalJSON() ([]byte, error) {
	der := key.EncodeToDER()
	b64der := base64.RawURLEncoding.EncodeToString(der)
	return []byte(`"` + string(b64der) + `"`), nil
}

// UnmarshalJSON decodes b and sets result to *key.
func (key *PublicKey) UnmarshalJSON(b []byte) error {
	trimmed := bytes.Trim(b, "\"")
	der := make([]byte, base64.RawURLEncoding.DecodedLen(len(trimmed)))
	_, err := base64.RawURLEncoding.Decode(der, trimmed)
	if err != nil {
		log.Warningf("Can't decode base64-encoded pubkey '%s'", trimmed)
		return errors.Parsing
	}
	res, err := ParsePublicKeyFromDER(der)
	if res != nil {
		*key = *res
	}
	return err
}

// Verify checks whether the sig of the data corresponds the public key.
func (key *PublicKey) Verify(data []byte, sig Signature) bool {
	digest := sha256.Sum256(data)

	curveOrderByteLen := key.Curve.Params().P.BitLen() / 8

	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sig[:curveOrderByteLen])
	s.SetBytes(sig[curveOrderByteLen:])

	return ecdsa.Verify((*ecdsa.PublicKey)(key), digest[:], r, s)
}
