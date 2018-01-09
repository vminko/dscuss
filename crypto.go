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

// Some parts copied from github.com/gtank/cryptopasta/.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
)

type privateKey ecdsa.PrivateKey
type PublicKey ecdsa.PublicKey
type Signature []byte

// Signer hides private key from caller and offers simple interface for signing
// data.
type Signer struct {
	privkey *privateKey
}

func (s *Signer) public() *PublicKey {
	return s.privkey.public()
}

func (s *Signer) sign(data []byte) (Signature, error) {
	return sign(data, s.privkey)
}

// newPrivateKey generates a random P-224 ECDSA private key.
func newPrivateKey() (*privateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	return (*privateKey)(key), err
}

// parsePrivateKey decodes a PEM-encoded ECDSA private key.
func parsePrivateKey(encodedKey []byte) (*privateKey, error) {
	var block *pem.Block

	for {
		block, encodedKey = pem.Decode(encodedKey)
		if block == nil {
			Log(ERROR, "Failed to find EC PRIVATE KEY in PEM data")
			return nil, ErrParsing
		}
		if block.Type == "EC PRIVATE KEY" {
			break
		} else {
			continue
		}
	}

	privkey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		Logf(ERROR, "Can't parse private key %v", err)
		return nil, ErrParsing
	}

	return (*privateKey)(privkey), nil
}

// encode encodes an ECDSA private key to PEM format.
func (key *privateKey) encode() ([]byte, error) {
	derKey, err := x509.MarshalECPrivateKey((*ecdsa.PrivateKey)(key))
	if err != nil {
		Logf(ERROR, "Can't encode private key %v", err)
		return nil, ErrInternal
	}

	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: derKey,
	}

	return pem.EncodeToMemory(keyBlock), nil
}

// public returns the public key corresponding to the private key.
func (key *privateKey) public() *PublicKey {
	cryptoPub := (*ecdsa.PrivateKey)(key).Public()
	if ecdsaPub, ok := (cryptoPub).(*ecdsa.PublicKey); ok {
		pubKey := (*PublicKey)(ecdsaPub)
		return pubKey
	} else {
		panic(fmt.Sprintf("wrong cryptoPub type: %T", cryptoPub))
	}
}

// parsePublicKey decodes a DER-encoded ECDSA public key.
func parsePublicKey(encodedKey []byte) (*PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(encodedKey)
	if err != nil {
		Logf(WARNING, "Can't parse public key %v", err)
		return nil, ErrParsing
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		panic("data was not an ECDSA public key")
	}

	return (*PublicKey)(ecdsaPub), nil
}

// encode encodes an ECDSA public key to DER format.
func (key *PublicKey) encode() ([]byte, error) {
	derBytes, err := x509.MarshalPKIXPublicKey((*ecdsa.PublicKey)(key))
	if err != nil {
		Logf(ERROR, "MarshalPKIXPublicKey failed: %v", err)
		return nil, ErrInternal
	}

	return derBytes, nil
}

// MarshalJSON returns the JSON encoded key.
func (key *PublicKey) MarshalJSON() ([]byte, error) {
	der, err := key.encode()
	if err != nil {
		return nil, err
	}
	b64der := base64.RawURLEncoding.EncodeToString(der)
	return []byte(`"` + string(b64der) + `"`), nil
}

// UnmarshalJSON decodes b and sets result to *key.
func (key *PublicKey) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	der, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		Logf(WARNING, "Can't decode base64-encoded pubkey '%s'", s)
		return ErrParsing
	}
	res, err := parsePublicKey(der)
	if res != nil {
		*key = *res
	}
	return err
}

// sign creates a signature for the data using the specified privkey.
func sign(data []byte, privkey *privateKey) (Signature, error) {
	digest := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(privkey), digest[:])
	if err != nil {
		Logf(ERROR, "Can't sign data using private key %s %v", privkey, err)
		return nil, ErrInternal
	}

	// Encode the signature {R, S}
	// big.Int.Bytes() will need padding in the case of leading zero bytes
	params := privkey.Curve.Params()
	curveOrderByteLen := params.P.BitLen() / 8
	rBytes, sBytes := r.Bytes(), s.Bytes()
	Signature := make([]byte, curveOrderByteLen*2)
	copy(Signature[curveOrderByteLen-len(rBytes):], rBytes)
	copy(Signature[curveOrderByteLen*2-len(sBytes):], sBytes)

	return Signature, nil
}

// verify checks whether the sig of the data corresponds the specified pubkey.
func verify(data []byte, sig Signature, pubkey *PublicKey) bool {
	digest := sha256.Sum256(data)

	curveOrderByteLen := pubkey.Curve.Params().P.BitLen() / 8

	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sig[:curveOrderByteLen])
	s.SetBytes(sig[curveOrderByteLen:])

	return ecdsa.Verify((*ecdsa.PublicKey)(pubkey), digest[:], r, s)
}

// encode encodes an ECDSA signature according to
// https://tools.ietf.org/html/rfc7515#appendix-A.3.1
func (sig Signature) encode() string {
	return base64.RawURLEncoding.EncodeToString(sig)
}

// parseSignature decodes an ECDSA signature according to
// https://tools.ietf.org/html/rfc7515#appendix-A.3.1
func parseSignature(b64sig string) (Signature, error) {
	sig, err := base64.RawURLEncoding.DecodeString(b64sig)
	if err != nil {
		Logf(ERROR, "Can't decode base64-encoded signture %x", b64sig)
		return nil, ErrParsing
	}
	return sig, nil
}

// marshaljson returns the json encoded key.
func (sig Signature) MarshalJSON() ([]byte, error) {
	return []byte(sig.encode()), nil
}

// unmarshaljson decodes b and sets result to *sig.
func (sig *Signature) UnmarshalJSON(b []byte) error {
	res, err := parseSignature(string(b))
	if err == nil {
		*sig = res
	}
	return err
}
