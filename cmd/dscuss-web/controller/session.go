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
package controller

import (
	"errors"
	"time"
	"vminko.org/dscuss"
	"vminko.org/dscuss/log"
)

type Session struct {
	ID              string
	IsAuthenticated bool
	CSRFToken       string
	CreatedDate     time.Time
	UpdatedDate     time.Time
}

const (
	sessionIDLength = 32
	csrfTokenLength = 32
	maxSessionLife  = 200 * time.Hour
)

func (s *Session) Authenticate(loginHandle *dscuss.LoginHandle, user string, pass string) error {
	u := loginHandle.GetLoggedUser()
	if !(user == u.Nickname && pass == password) {
		log.Debugf("VMINKO incorrect u or p: %s %s %s %s", user, u.Nickname, pass, password)
		return errors.New("Incorrect username or password")
	}
	s.IsAuthenticated = true
	log.Debug("VMINKO auth ok")
	return nil
}
