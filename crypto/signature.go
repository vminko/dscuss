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
	"encoding/base64"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

type Signature []byte

// Encode encodes an ECDSA signature according to
// https://tools.ietf.org/html/rfc7515#appendix-A.3.1
func (sig Signature) Encode() string {
	return base64.RawURLEncoding.EncodeToString(sig)
}

// ParseSignature decodes an ECDSA signature according to
// https://tools.ietf.org/html/rfc7515#appendix-A.3.1
func ParseSignature(b64sig []byte) (Signature, error) {
	sig := make([]byte, base64.RawURLEncoding.DecodedLen(len(b64sig)))
	_, err := base64.RawURLEncoding.Decode(sig, b64sig)
	if err != nil {
		log.Errorf("Can't decode base64-encoded signture %x", b64sig)
		return nil, errors.Parsing
	}
	return sig, nil
}

// MarshalJSON returns the JSON-encoded key.
func (sig Signature) MarshalJSON() ([]byte, error) {
	return []byte(`"` + sig.Encode() + `"`), nil
}

// UnmarshalJSON decodes b and sets result to *sig.
func (sig *Signature) UnmarshalJSON(b []byte) error {
	trimmed := bytes.Trim(b, "\"")
	res, err := ParseSignature(trimmed)
	if err == nil {
		copy([]byte(*sig), res)
	}
	return err
}
