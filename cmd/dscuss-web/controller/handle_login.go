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
	"html/template"
	"net/http"
	"net/url"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
	"vminko.org/dscuss/entity"
)

const (
	MaxPasswordLen = 64
)

func handleLogin(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) > 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	var err error
	redirectURL := r.FormValue("next")
	if r.Method == "POST" {
		// FormValue() returns URL-decoded value for GET methods
		redirectURL, err = url.QueryUnescape(redirectURL)
	}
	if err != nil || redirectURL == "" || redirectURL[0] != '/' {
		redirectURL = "/"
	}
	if s.IsAuthenticated {
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}
	var msg string
	if r.Method == "POST" {
		user := r.PostFormValue("username")
		pass := r.PostFormValue("password")
		if len(user) > entity.MaxUsernameLen || len(pass) > MaxPasswordLen {
			msg = "Specified username or password is too long."
		} else if err = s.Authenticate(l, user, pass); err != nil {
			msg = err.Error()
		} else {
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}
	}
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Owner authentication"
	view.Render(w, "login.html", map[string]interface{}{
		"Common":  cd,
		"next":    template.URL(url.QueryEscape(redirectURL)),
		"Message": msg,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		NotFoundHandler(w, r)
		return
	}
	defer InternalServerErrorHandler(w, r)
	CloseSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
