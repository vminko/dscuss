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
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/thread"
)

type ThreadComposer struct {
	t *Thread
	l *dscuss.LoginHandle
}

func (tc *ThreadComposer) Handle(n *thread.Node) bool {
	m := n.Msg
	if m == nil {
		return true
	}
	if n.IsRoot() {
		tc.t.RootMessage.Assign(m, tc.l)
	} else {
		tc.t.Replies = append(tc.t.Replies, Message{})
		r := &tc.t.Replies[len(tc.t.Replies)-1]
		r.Assign(m, tc.l)
	}
	return true
}

func threadHandler(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	var validURI = regexp.MustCompile("^/thread(id=[a-zA-Z0-9\\/+=]{32})?$")
	m := validURI.FindStringSubmatch(r.URL.Path)
	if m == nil {
		NotFoundHandler(w, r)
		return
	}
	idStr := r.URL.Query().Get("id")
	var tid entity.ID
	err := tid.ParseString(idStr)
	if err != nil {
		BadRequestHandler(w, r, idStr+" is not a valid entity ID.")
		return
	}
	node, err := l.ListThread(&tid)
	if err != nil {
		log.Fatal("Can't list thread: " + err.Error() + ".")
		return
	}
	var t Thread
	tc := ThreadComposer{&t, l}
	tvis := thread.NewViewingVisitor(&tc)
	node.View(tvis)
	view.Render(w, "thread.html", map[string]interface{}{
		"Common":        readCommonData(r, s, l),
		"ID":            t.ID,
		"Topic":         t.Topic,
		"Subject":       t.Subject,
		"Text":          t.Text,
		"DateWritten":   t.DateWritten,
		"AuthorName":    t.AuthorName,
		"AuthorID":      t.AuthorID,
		"AuthorShortID": t.AuthorShortID,
		"Replies":       t.Replies,
	})
}

func MakeThreadHandler(l *dscuss.LoginHandle) http.HandlerFunc {
	return makeHandler(threadHandler, l)
}
