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
	"net/http"
	"net/url"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
)

func profileHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) != 0 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	var subs []string
	ss := l.ListSubscriptions()
	for _, t := range ss {
		subs = append(subs, t.String())
	}
	var moders []User
	mm := l.ListModerators()
	for _, mdr := range mm {
		u, err := l.GetUser(mdr)
		if err != nil {
			panic("Failed to fetch user: " + err.Error())
		}
		moders = append(moders, User{})
		m := &moders[len(moders)-1]
		m.Assign(u, l)
	}
	var msg string
	if r.Method == "POST" {
		http.Redirect(w, r, "/board", http.StatusSeeOther)
		goto render
		return
	}
render:
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Node owner's profile"
	view.Render(w, "profile.html", map[string]interface{}{
		"Common":        cd,
		"Moderators":    moders,
		"Subscriptions": subs,
		"Message":       msg,
	})
}

func addModeratorHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) != 0 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	idStr := r.FormValue("id")
	idStr, err := url.QueryUnescape(idStr)
	if err != nil {
		BadRequestHandler(w, r, idStr+" is not a valid URL-encoded string.")
		return
	}
	var id entity.ID
	err = id.ParseString(idStr)
	if err != nil {
		BadRequestHandler(w, r, "'"+idStr+"' is not a valid entity ID.")
		return
	}
	err = l.AddModerator(&id)
	if err == errors.AlreadyModerator {
		BadRequestHandler(w, r, "Can't add new moderator: "+err.Error()+".")
	} else if err != nil {
		panic("Error making new moderator: " + err.Error() + ".")
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func removeModeratorHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) != 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	idStr := r.URL.Query().Get("id")
	var mid entity.ID
	err := mid.ParseString(idStr)
	if err != nil {
		BadRequestHandler(w, r, idStr+" is not a valid entity ID.")
		return
	}
	err = l.RemoveModerator(&mid)
	if err != nil {
		panic("Error removing moderator: " + err.Error() + ".")
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func MakeProfileHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(profileHandler, l)
}

func MakeAddModeratorHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(addModeratorHandler, l)
}

func MakeRemoveModeratorHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(removeModeratorHandler, l)
}
