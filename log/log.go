/*
This file is part of Dscuss.
Copyright (C) 2017-2018  Vitaly Minko

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

// log provides logging facilities for Dscuss.
package log

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
)

var debug bool

// caller returns the name of the third function in the current stack.
func caller() string {
	pc := make([]uintptr, 10)
	runtime.Callers(1, pc)
	f := runtime.FuncForPC(pc[3])
	trimmedName := strings.TrimPrefix(f.Name(), "vminko.org/dscuss")
	trimmedName = strings.TrimLeft(trimmedName, "/.")
	return fmt.Sprintf("[%s]", trimmedName)
}

func SetDebug(d bool) {
	debug = d
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func debugf(format string, args ...interface{}) {
	if debug {
		log.Printf("DEBUG "+caller()+" "+format, args...)
	}
}

func Debug(msg string) {
	debugf("%s", msg)
}

func Debugf(format string, args ...interface{}) {
	debugf(format, args...)
}

func infof(format string, args ...interface{}) {
	log.Printf("INFO "+caller()+" "+format, args...)
}

func Info(msg string) {
	infof("%s", msg)
}

func Infof(format string, args ...interface{}) {
	infof(format, args...)
}

func warningf(format string, args ...interface{}) {
	log.Printf("WARNING "+caller()+" "+format, args...)
}

func Warning(msg string) {
	warningf("%s", msg)
}

func Warningf(format string, args ...interface{}) {
	warningf(format, args...)
}

func errorf(format string, args ...interface{}) {
	log.Printf("ERROR "+caller()+" "+format, args...)
}

func Error(msg string) {
	errorf("%s", msg)
}

func Errorf(format string, args ...interface{}) {
	errorf(format, args...)
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf("FATAL ERROR: "+format+" :-(\n", args...)
	log.Fatalf("FATAL "+caller()+" "+format, args...)
}

func Fatal(msg string) {
	fatalf("%s", msg)
}

func Fatalf(format string, args ...interface{}) {
	fatalf(format, args...)
}
