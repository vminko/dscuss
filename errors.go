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
	"errors"
)

var (
	ErrInternal     = errors.New("internal error")
	ErrDatabase     = errors.New("database error")
	ErrFilesystem   = errors.New("filesystem error")
	ErrParsing      = errors.New("error parsing input data")
	ErrConfig       = errors.New("error handling config file")
	ErrNoSuchEntity = errors.New("can't find requested entity")
	// User-related errors
	ErrWrongNickname     = errors.New("unacceptable nickname")
	ErrAlreadyLoggedIn   = errors.New("another user is already logged in")
	ErrAlreadyRegistered = errors.New("such user is already registered")
	ErrNoSuchUser        = errors.New("can't find specified user")
)
