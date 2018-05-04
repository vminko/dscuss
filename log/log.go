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
)

var debug bool

// caller returns the name of the third function in the current stack.
func caller() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[2])
	return fmt.Sprintf("[%s]", f.Name())
}

func SetDebug(d bool) {
	debug = d
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func Debugf(format string, args ...interface{}) {
	if debug {
		log.Printf("DEBUG "+caller()+" "+format, args...)
	}
}

func Debug(msg string) {
	Debugf("%s", msg)
}

func Infof(format string, args ...interface{}) {
	log.Printf("INFO "+caller()+" "+format, args...)
}

func Info(msg string) {
	Infof("%s", msg)
}

func Warningf(format string, args ...interface{}) {
	log.Printf("WARNING "+caller()+" "+format, args...)
}

func Warning(msg string) {
	Warningf("%s", msg)
}

func Errorf(format string, args ...interface{}) {
	log.Printf("ERROR "+caller()+" "+format, args...)
}

func Error(msg string) {
	Errorf("%s", msg)
}

func Fatalf(format string, args ...interface{}) {
	fmt.Printf("FATAL ERROR: "+format+"\n", args...)
	log.Fatalf("FATAL "+caller()+" "+format, args...)
}

func Fatal(msg string) {
	Fatalf("%s", msg)
}
