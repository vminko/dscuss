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
	"regexp"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
)

const (
	maxUsernameLength = 64
	maxPasswordLength = 64
)

func loginHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	var validURI = regexp.MustCompile("^/login(next=[a-zA-Z0-9\\/+=]+)?$")
	m := validURI.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	var err error
	redirectURL := r.FormValue("next")
	// FormValue() returns URL-decoded value for GET methods
	if r.Method == "POST" {
		redirectURL, err = url.QueryUnescape(r.FormValue("next"))
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
		if len(user) > maxUsernameLength || len(pass) > maxPasswordLength {
			msg = "Specified username or password is too long."
		} else if err = s.Authenticate(l, user, pass); err != nil {
			msg = err.Error()
		} else {
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}
	}
	view.Render(w, "login.html", map[string]interface{}{
		"Common":  readCommonData(r, s, l),
		"next":    template.URL(url.QueryEscape(redirectURL)),
		"Message": msg,
	})
}

func MakeLoginHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(loginHandler, l)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		http.NotFound(w, r)
		return
	}
	defer InternalServerErrorHandler(w, r)
	CloseSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
