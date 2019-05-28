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

// controller responds to the user input and performs interactions on
// the Dscuss data model objects (Users, Messages, Operations, Threads,
// Subscriptions etc.) via Dscuss API.
package controller

import (
	"io"
	"net/http"
	"runtime/debug"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/static"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	if r := recover(); r != nil {
		log.Errorf("[INFO] Recovered from panic: %s\n[INFO] Debug stack: %s\n",
			r, debug.Stack())
		http.Error(w, "Internal server error. This event has been logged.",
			http.StatusInternalServerError)
	}
}

func BadRequestHandler(w http.ResponseWriter, r *http.Request, msg string) {
	http.Error(w, "Your browser sent a request that this server could not understand. "+msg,
		http.StatusBadRequest)
}

func ForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "403 Forbidden", http.StatusForbidden)
}

func CSSHandler(w http.ResponseWriter, r *http.Request) {
	defer InternalServerErrorHandler(w, r)
	w.Header().Set("Content-Type", "text/css")
	w.Header().Set("Cache-Control", "max-age=31536000, public")
	io.WriteString(w, static.CSS)
}

func MakeRootHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(
		func(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
			defer InternalServerErrorHandler(w, r)
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/board", http.StatusSeeOther)
		},
		l,
	)
}

func makeHandler(
	fn func(http.ResponseWriter, *http.Request, *dscuss.LoginHandle, *Session),
	l *dscuss.LoginHandle,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer InternalServerErrorHandler(w, r)
		if l == nil {
			log.Fatal("No user is logged in.")
		}
		s := OpenSession(w, r)
		if r.Method == "POST" && r.PostFormValue("csrf") != s.CSRFToken {
			log.Debugf("PostFromValue = %s, Token = %s", r.PostFormValue("csrf"), s.CSRFToken)
			ForbiddenHandler(w, r)
			return
		}
		fn(w, r, l, s)
	}
}

func userName(l *dscuss.LoginHandle, id *entity.ID) string {
	u, err := l.GetUser(id)
	switch {
	case err == errors.NoSuchEntity:
		return "[unknown user]"
	case err != nil:
		return "[error fetching user from db]"
	default:
		return u.Nickname
	}
}
