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

func handleRemoveMessage(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) > 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	tidStr := r.FormValue("id")
	if r.Method == "POST" {
		// FormValue() returns URL-decoded value for GET methods
		tidStr, err := url.QueryUnescape(tidStr)
		if err != nil {
			BadRequestHandler(w, r, tidStr+" is not a valid URL-encoded string.")
			return
		}
	}
	var tid entity.ID
	err := tid.ParseString(tidStr)
	if err != nil {
		BadRequestHandler(w, r, "'"+tidStr+"' is not a valid entity ID.")
		return
	}

	var msg string
	var tg Message
	var op Operation
	m, err := l.GetMessage(&tid)
	if err == errors.NoSuchEntity {
		NotFoundHandler(w, r)
	} else if err != nil {
		panic("Got an error while fetching msg " + tid.Shorten() +
			" from DB: " + err.Error())
	}
	tg.Assign(m, l)
	root, err := l.GetRootMessage(m)
	if err != nil {
		panic("Got an error while fetching root for msg " + tid.Shorten() +
			" from DB:" + err.Error())
	}

	if r.Method == "POST" {
		op.Reason = r.PostFormValue("reason")
		op.Comment = r.PostFormValue("comment")
		if len(op.Comment) > entity.MaxOperationCommentLen {
			msg = "Specified comment is too long."
			goto render
		}
		var reason entity.OperationReason
		err = reason.ParseString(op.Reason)
		if err != nil {
			BadRequestHandler(w, r, op.Reason+" is not a valid operation reason.")
			return
		}
		oper, err := l.NewOperation(entity.OperationTypeRemoveMessage, reason, op.Comment, &tid)
		if err != nil {
			panic("Error making new operation: " + err.Error() + ".")
		}
		err = l.PostEntity((entity.Entity)(oper))
		if err != nil {
			panic("Error posting new operation: " + err.Error() + ".")
		}
		http.Redirect(
			w, r,
			"/thread?id="+url.QueryEscape(root.ID().String()),
			http.StatusSeeOther,
		)
		return
	}
render:
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Removing message #" + tg.ShortID
	cd.Topic = root.Topic.String()
	view.Render(w, "oper_del.html", map[string]interface{}{
		"Common":    cd,
		"Target":    tg,
		"Operation": op,
		"Message":   msg,
	})
}
