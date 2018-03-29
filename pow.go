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
	"runtime"
	"sync"
	"sync/atomic"
)

const (
	powTargetBits  = 8 /*TBD: must be set to 16 in release version */
	powKeyLenBytes = 32
	powSalt        = "dscuss-proof-of-work"
)

type ProofOfWork uint64

// powFinder implements proof-of-work algorithm (also known as Hashcash).
// Proof-of-work is used in Dscuss to protect against Sybil attack.
type powFinder struct {
	data    []byte
	target  *big.Int
	counter uint64
}

func newPowFinder(data []byte) *powFinder {
	target := big.NewInt(1)
	target.Lsh(target, uint(powKeyLenBytes*8-powTargetBits))
	pf := &powFinder{data, target, 0}
	return pf
}

// setComplexity changes number of target bits in the proof-of-work.
// The higher value you set, the harder it will be to find PoW.
// The maximum value of TargetBitNum is powKeyLenBytes*8.
// Should only be used for benchmarking.
func (pf *powFinder) setComplexity(targetBitNum int) {
	pf.target = big.NewInt(1)
	pf.target.Lsh(pf.target, uint(powKeyLenBytes*8-targetBitNum))
}

func (pf *powFinder) worker(
	workerID int,
	resultChan chan uint64,
	stopChan chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for nonce := uint64(0); nonce < math.MaxUint64; {
		nonce := atomic.AddUint64(&pf.counter, 1)
		select {
		case <-stopChan:
			return
		default:
			Logf(DEBUG, "Worker #%d is trying PoW %d", workerID, nonce)
			if pf.validate(nonce) {
				Logf(DEBUG, "Worker #%d has found PoW: \"%d\"", workerID, nonce)
				resultChan <- nonce
				return
			}
		}
	}
	resultChan <- 0
}

func (pf *powFinder) find() ProofOfWork {
	Logf(DEBUG, "Looking for proof-of-work for %x", pf.data)
	resultChan := make(chan uint64)
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		Logf(DEBUG, "Starting PoW wroker #%d", i)
		wg.Add(1)
		go pf.worker(i, resultChan, stopChan, &wg)
	}
	proof := <-resultChan
	close(stopChan)
	wg.Wait()
	if proof == 0 {
		// The probability of this case is very close to 0.
		// It's OK to panic here in the proof-of-concept version.
		Log(FATAL, "Failed to find proof-of-work")
	}
	Logf(DEBUG, "PoW is found: %d", proof)
	return ProofOfWork(proof)
}

func (pf *powFinder) validate(nonce uint64) bool {
	var keyInt big.Int
	var key []byte
	data := pf.prepareData(nonce)
	// The recommended parameters for interactive logins as of 2017.
	key, err := scrypt.Key(data, []byte(powSalt), 32768, 8, 1, powKeyLenBytes)
	if err != nil {
		Log(FATAL, "Incorrect key derivation parameters")
	}
	keyInt.SetBytes(key[:])
	return keyInt.Cmp(pf.target) == -1
}

func (pf *powFinder) prepareData(nonce uint64) []byte {
	nbuf := make([]byte, 8)
	binary.BigEndian.PutUint64(nbuf, nonce)
	data := bytes.Join(
		[][]byte{
			pf.data,
			nbuf,
		},
		[]byte{},
	)
	return data
}
