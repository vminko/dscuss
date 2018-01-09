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

// Logging facilities for Dscuss.

import (
	"fmt"
	"log"
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func Logf(level int, format string, args ...interface{}) {
	switch level {
	case DEBUG:
		if debug {
			log.Printf("DEBUG: "+format, args...)
		}
	case INFO:
		log.Printf("INFO: "+format, args...)
	case WARNING:
		log.Printf("WARNING: "+format, args...)
	case ERROR:
		log.Printf("ERROR: "+format, args...)
	case FATAL:
		fmt.Printf("FATAL ERROR: "+format+"\n", args...)
		log.Fatalf("FATAL: "+format, args...)
	default:
		panic("Unknown log level.")
	}
}

func Log(level int, msg string) {
	Logf(level, "%s", msg)
}

func panicf() {

}
