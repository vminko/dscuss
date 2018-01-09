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
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/scrypt"
	"math"
	"math/big"
)

const (
	powTargetBits  = 4 /*TBD: customise for release version */
	powKeyLenBytes = 32
	powSalt        = "dscuss-proof-of-work"
)

type ProofOfWork uint64

// powFinder implements proof-of-work algorithm (also known as Hashcash).
// Proof-of-work is used in Dscuss to protect against Sybil attack.
type powFinder struct {
	data   []byte
	target *big.Int
}

func newPowFinder(data []byte) *powFinder {
	target := big.NewInt(1)
	target.Lsh(target, uint(powKeyLenBytes*8-powTargetBits))

	pf := &powFinder{data, target}

	return pf
}

func (pf *powFinder) find() ProofOfWork {
	var proof uint64 = 0

	Logf(DEBUG, "Looking for proof-of-work for \"%s\"...", pf.data)
	for proof < math.MaxUint64 {
		if pf.isValid(ProofOfWork(proof)) {
			Logf(DEBUG, "PoW is found: \"%d\"", proof)
			Logf(DEBUG, "PoW hash is: [% x]", pf.data[:])
			break
		} else {
			Logf(DEBUG, "PoW: trying %d", proof)
			proof++
		}
	}

	if proof == math.MaxUint64 {
		// The probability of this case is very close to 0.
		// It's OK to panic here in the proof-of-concept version.
		panic("Failed to find proof-of-work")
	}

	return ProofOfWork(proof)
}

func (pf *powFinder) isValid(proof ProofOfWork) bool {
	var keyInt big.Int
	var key []byte
	data := pf.prepareData(proof)
	// The recommended parameters for interactive logins as of 2017.
	key, err := scrypt.Key(data, []byte(powSalt), 32768, 8, 1, powKeyLenBytes)
	if err != nil {
		panic("Incorrect key derivation parameters")
	}
	keyInt.SetBytes(key[:])
	Logf(DEBUG, "Pow: scrypt key is %s, expected %s", keyInt.String(), (*pf.target).String())
	return keyInt.Cmp(pf.target) == -1
}

func (pf *powFinder) prepareData(nonce ProofOfWork) []byte {
	nbuf := make([]byte, 8)
	binary.BigEndian.PutUint64(nbuf, uint64(nonce))
	data := bytes.Join(
		[][]byte{
			pf.data,
			nbuf,
		},
		[]byte{},
	)
	return data
}
