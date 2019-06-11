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

func replyHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) != 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	pidStr := r.FormValue("id")
	if r.Method == "POST" {
		// FormValue() returns URL-decoded value for GET methods
		pidStr, err := url.QueryUnescape(pidStr)
		if err != nil {
			BadRequestHandler(w, r, pidStr+" is not a valid URL-encoded string.")
			return
		}
	}
	var pid entity.ID
	err := pid.ParseString(pidStr)
	if err != nil {
		BadRequestHandler(w, r, "'"+pidStr+"' is not a valid entity ID.")
		return
	}

	var msg string
	var showSubj bool
	var rm RootMessage
	var pm Message
	var rpl ComposedReply
	m, err := l.GetMessage(&pid)
	if err == errors.NoSuchEntity {
		NotFoundHandler(w, r)
	} else if err != nil {
		panic("Got an error while fetching msg " + pid.Shorten() +
			" from DB: " + err.Error())
	}
	pm.Assign(m, l)

	root, err := l.GetRootMessage(m)
	if err != nil {
		panic("Got an error while fetching root for msg " + pid.Shorten() +
			" from DB:" + err.Error())
	}
	rm.Assign(root, l)

	if r.Method == "POST" {
		rpl.Subject = r.PostFormValue("subject")
		rpl.Text = r.PostFormValue("text")
		if (rpl.Subject == "") || (len(rpl.Subject) > entity.MaxSubjectLen) {
			msg = "Specified subject is unacceptable: empty or too long."
			goto render
		}
		if (rpl.Text == "") || (len(rpl.Text) > entity.MaxTextLen) {
			msg = "Specified message text is unacceptable: empty or too long."
			goto render
		}
		rplMsg, err := l.NewReply(rpl.Subject, rpl.Text, &pid)
		if err != nil {
			msg = "Error making new reply: " + err.Error() + "."
			goto render
		}
		err = l.PostEntity((entity.Entity)(rplMsg))
		if err != nil {
			msg = "Error posting new reply: " + err.Error() + "."
			goto render
		}
		http.Redirect(w, r, "/thread?id="+url.QueryEscape(rm.ID), http.StatusSeeOther)
		return
	}
render:
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Reply to " + rm.Subject
	cd.Topic = rm.Topic
	view.Render(w, "reply.html", map[string]interface{}{
		"Common":            cd,
		"Thread":            rm,
		"Parent":            pm,
		"Reply":             rpl,
		"ShowParentSubject": showSubj,
		"Message":           msg,
	})
}

func MakeReplyHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(replyHandler, l)
}
