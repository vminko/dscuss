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
	"vminko.org/dscuss/subs"
)

func handleCreateThread(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) > 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	topic := r.FormValue("topic")
	if r.Method == "POST" {
		topic, err := url.QueryUnescape(topic)
		if err != nil {
			BadRequestHandler(w, r, topic+" is not a valid URL-encoded string.")
			return
		}
	}
	var msg string
	var subj string
	var text string
	if r.Method == "POST" {
		t, err := subs.NewTopic(topic)
		if err != nil {
			msg = "Specified topic is unacceptable: " + err.Error()
			goto render
		}
		subj = r.PostFormValue("subject")
		text = r.PostFormValue("text")
		if (subj == "") || (len(subj) > entity.MaxSubjectLen) {
			msg = "Specified subject is unacceptable: empty or too long."
			goto render
		}
		if (text == "") || (len(text) > entity.MaxTextLen) {
			msg = "Specified message text is unacceptable: empty or too long."
			goto render
		}
		thread, err := l.NewThread(subj, text, t)
		if err != nil {
			msg = "Error making new dscussion: " + err.Error() + "."
			goto render
		}
		err = l.PostEntity((entity.Entity)(thread))
		if err != nil {
			msg = "Error posting new dscussion: " + err.Error() + "."
			goto render
		}
		http.Redirect(w, r, "/board", http.StatusSeeOther)
		return
	}
render:
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Start new dscussion"
	cd.Topic = topic
	view.Render(w, "thread_create.html", map[string]interface{}{
		"Common":  cd,
		"Subject": subj,
		"Text":    text,
		"Message": msg,
	})
}
