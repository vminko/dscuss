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

import "testing"

func BenchmarkPoW8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		privKey, _ := newPrivateKey()
		pow := newPowFinder(privKey.public().encodeToDER())
		pow.setComplexity(8)
		b.StartTimer()
		_ = pow.find()
	}
}

func BenchmarkPoW16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		privKey, _ := newPrivateKey()
		pow := newPowFinder(privKey.public().encodeToDER())
		pow.setComplexity(16)
		b.StartTimer()
		_ = pow.find()
	}
}
