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
	"regexp"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/subs"
)

func handleBoard(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) > 1 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	var validURI = regexp.MustCompile("^/board(topic=[a-z,]*)?$")
	m := validURI.FindStringSubmatch(r.URL.Path)
	if m == nil {
		NotFoundHandler(w, r)
		return
	}
	topicStr := r.URL.Query().Get("topic")
	var err error
	var topic subs.Topic
	if topicStr != "" {
		topic, err = subs.NewTopic(topicStr)
		if err != nil {
			BadRequestHandler(w, r, "Unacceptable topic: "+err.Error()+".")
			return
		}
	}

	const boardSize = 10
	var messages []*entity.Message
	if topic != nil {
		messages, err = l.ListTopic(topic, 0, boardSize)
	} else {
		messages, err = l.ListBoard(0, boardSize)
	}
	if err != nil {
		panic("Can't list board: " + err.Error() + ".")
	}

	var threads []RootMessage
	for _, msg := range messages {
		threads = append(threads, RootMessage{})
		t := &threads[len(threads)-1]
		t.Assign(msg, l)
	}

	cd := readCommonData(r, s, l)
	cd.PageTitle = "Board"
	cd.Topic = topicStr
	view.Render(w, "board.html", map[string]interface{}{
		"Common":  cd,
		"Threads": threads,
	})
}
