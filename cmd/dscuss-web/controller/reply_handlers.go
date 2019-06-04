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
	"regexp"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

const (
	maxSubjectLength = 128
	maxTextLength    = 1024
)

func replyHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	var validURI = regexp.MustCompile("^/reply(\\?id=[a-zA-Z0-9\\/+=]+)?$")
	if validURI.FindStringSubmatch(r.URL.Path) == nil {
		NotFoundHandler(w, r)
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
		BadRequestHandler(w, r, pidStr+" is not a valid entity ID.")
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
		log.Fatalf("Got an error while fetching msg %s from DB: %v",
			pid.Shorten(), err)
	}
	pm.Assign(m, l)

	root, err := l.GetRootMessage(m)
	if err != nil {
		log.Fatalf("Got an error while fetching root for msg %s from DB: %v",
			pid.Shorten(), err)
	}
	rm.Assign(root, l)

	if r.Method == "POST" {
		rpl.Subject = r.PostFormValue("subject")
		rpl.Text = r.PostFormValue("text")
		if len(rpl.Subject) > maxSubjectLength || len(rpl.Text) > maxTextLength {
			msg = "Specified subject or text is too long."
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
	}
render:
	view.Render(w, "reply.html", map[string]interface{}{
		"Common":            readCommonData(r, s, l),
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
