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
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
	"vminko.org/dscuss/log"
)

var sessionPool *SessionPool
var password string

func init() {
	sessionPool = newSessionPool()
}

func SetPassword(p string) {
	password = p
}

func newRandString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Unable to generate random string: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func OpenSession(w http.ResponseWriter, r *http.Request) *Session {
	cookie, err := r.Cookie("sessionid")
	if err == nil {
		sessionID := cookie.Value
		if s, ok := sessionPool.load(sessionID); ok {
			if s.UpdatedDate.After(time.Now().Add(-maxSessionLife)) {
				s.UpdatedDate = time.Now()
				return s
			} else {
				log.Debugf("Session %s outdated and will be removed.", sessionID)
			}
		} else {
			log.Debugf("Session %s not found.", sessionID)
		}
	}
	sessionPool.removeExpiredSessions()
	s := &Session{
		newRandString(sessionIDLength),
		false,
		newRandString(csrfTokenLength),
		time.Now(),
		time.Now(),
	}
	sessionPool.add(s)
	http.SetCookie(w, &http.Cookie{Name: "sessionid", Path: "/", Value: s.ID, HttpOnly: true})
	http.SetCookie(w, &http.Cookie{Name: "csrftoken", Path: "/", Value: s.CSRFToken})
	return s
}

func CloseSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionid")
	if err == nil {
		sessionID := cookie.Value
		sessionPool.removeSession(sessionID)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionid",
		Value:    "",
		Expires:  time.Now().Add(-300 * time.Hour),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "csrftoken",
		Value:   "",
		Expires: time.Now().Add(-300 * time.Hour),
	})
}
