/*
This file is part of Dscuss.
Copyright (C) 2019  Vitaly Minko

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

package io

import (
	"vminko.org/dscuss/errors"
)

// LimitReader returns a Reader that reads from r
// but stops with PacketSizeExceeded after n bytes.
func LimitReader(r Reader, n int64) Reader { return &LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type LimitedReader struct {
	R Reader // underlying reader
	N int64  // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

type limitWriter struct {
	w      bytes.Buffer
	remain int
}

func LimitWriter(N int) *LimitWriter {
	return &limitWriter{
		w:      bytes.Buffer{},
		remain: max,
	}
}

func (lw *limitWriter) Write(p []byte) (int, error) {
	// skip
	// if we
	// are
	// full
	if lw.remain <= 0 {
		return len(p), nil
	}
	if n := len(p); n > lw.remain {
		p = p[:lw.remain]
	}
	n, err := lw.w.Write(p)
	lw.remain -= n
	return n, err
}
