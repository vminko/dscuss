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
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
)

func handleUser(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) != 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	uidStr := r.URL.Query().Get("id")
	var uid entity.ID
	err := uid.ParseString(uidStr)
	if err != nil {
		BadRequestHandler(w, r, "'"+uidStr+"' is not a valid entity ID.")
		return
	}
	ent, err := l.GetUser(&uid)
	if err == errors.NoSuchEntity {
		NotFoundHandler(w, r)
	} else if err != nil {
		panic("Got an error while fetching user " + uid.Shorten() +
			" from DB: " + err.Error())
	}
	var u User
	u.Assign(ent)
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Profile of " + u.Nickname + "-" + u.ShortID
	view.Render(w, "user.html", map[string]interface{}{
		"Common": cd,
		"User":   u,
	})
}
