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

package errors

import (
	"errors"
)

var (
	Internal          = errors.New("internal error")
	Filesystem        = errors.New("filesystem error")
	Parsing           = errors.New("error parsing input data")
	Config            = errors.New("error handling config file")
	Database          = errors.New("database error")
	CantOpenDB        = errors.New("unable to open the database")
	DBOperFailed      = errors.New("error operating on the database")
	NoSuchEntity      = errors.New("can't find requested entity")
	WrongNickname     = errors.New("unacceptable nickname")
	AlreadyLoggedIn   = errors.New("another user is already logged in")
	AlreadyRegistered = errors.New("such user is already registered")
	NoSuchUser        = errors.New("can't find specified user")
	WrongPacketType   = errors.New("wrong packet type")
	MalformedPayload  = errors.New("payload of the packet is malformed")
	ProtocolViolation = errors.New("protocol violation detected")
)
