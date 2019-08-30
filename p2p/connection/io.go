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

package connection

import (
	"io"
	"vminko.org/dscuss/errors"
)

// limitReader returns a Reader that reads from r
// but stops with PacketSizeExceeded after n bytes.
func limitReader(r io.Reader, n int) io.Reader { return &limitedReader{r, n} }

type limitedReader struct {
	R io.Reader // underlying reader
	N int       // max bytes remaining
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	if (l.N <= 0) || (len(p) > l.N) {
		return 0, errors.PacketSizeExceeded
	}
	n, err = l.R.Read(p)
	l.N -= n
	return
}

// limitWriter returns a Writer that writes to w
// but stops with PacketSizeExceeded after n bytes.
type limitedWriter struct {
	W io.Writer // underlying writer
	N int       // max bytes remaining
}

func limitWriter(w io.Writer, n int) *limitedWriter { return &limitedWriter{w, n} }

func (l *limitedWriter) Write(p []byte) (n int, err error) {
	if (l.N <= 0) || (len(p) > l.N) {
		return 0, errors.PacketSizeExceeded
	}
	n, err = l.W.Write(p)
	l.N -= n
	return
}
