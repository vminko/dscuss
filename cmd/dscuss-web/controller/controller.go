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

// controller responds to the user input and performs interactions on
// the Dscuss data model objects (Users, Messages, Operations, Threads,
// Subscriptions etc.) via Dscuss API.
package controller

import (
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/static"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/p2p/peer"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func BadRequestHandler(w http.ResponseWriter, r *http.Request, msg string) {
	http.Error(w, "Your browser sent a request that this server could not perform. "+msg,
		http.StatusBadRequest)
}

func ForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "403 Forbidden", http.StatusForbidden)
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	if r := recover(); r != nil {
		log.Errorf("[INFO] Recovered from panic: %s\n[INFO] Debug stack: %s\n",
			r, debug.Stack())
		http.Error(w, "Internal server error. This event has been logged.",
			http.StatusInternalServerError)
	}
}

func CSSHandler(w http.ResponseWriter, r *http.Request) {
	defer InternalServerErrorHandler(w, r)
	w.Header().Set("Content-Type", "text/css")
	w.Header().Set("Cache-Control", "max-age=31536000, public")
	io.WriteString(w, static.CSS)
}

func JavaScriptHandler(w http.ResponseWriter, r *http.Request) {
	defer InternalServerErrorHandler(w, r)
	w.Header().Set("Content-Type", "text/javascript")
	w.Header().Set("Cache-Control", "max-age=31536000, public")
	io.WriteString(w, static.JavaScript)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/board", http.StatusSeeOther)
}

func makeHandler(
	fn func(http.ResponseWriter, *http.Request, *dscuss.LoginHandle, *Session),
	l *dscuss.LoginHandle,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer InternalServerErrorHandler(w, r)
		if l == nil {
			log.Fatal("No user is logged in.")
		}
		s := OpenSession(w, r)
		if r.Method == "POST" && r.PostFormValue("csrf") != s.CSRFToken {
			log.Debugf("PostFromValue = %s, Token = %s", r.PostFormValue("csrf"), s.CSRFToken)
			ForbiddenHandler(w, r)
			return
		}
		fn(w, r, l, s)
	}
}

func userName(id *entity.ID, l *dscuss.LoginHandle) string {
	u, err := l.GetUser(id)
	switch {
	case err == errors.NoSuchEntity:
		return "[unknown user]"
	case err != nil:
		return "[error fetching user from db]"
	default:
		return u.Nickname
	}
}

type User struct {
	ID       string
	ShortID  string
	Nickname string
	Info     string
	RegDate  string
}

func (u *User) Assign(eu *entity.User) {
	u.ID = eu.ID().String()
	u.ShortID = eu.ID().Shorten()
	u.Nickname = eu.Nickname
	u.Info = eu.Info
	u.RegDate = eu.RegDate.Format(time.RFC3339)
}

type Message struct {
	ID            string
	ShortID       string
	Subject       string
	Text          string
	DateWritten   string
	AuthorName    string
	AuthorID      string
	AuthorShortID string
}

func (m *Message) Assign(em *entity.Message, l *dscuss.LoginHandle) {
	m.ID = em.ID().String()
	m.ShortID = em.ID().Shorten()
	m.Subject = em.Subject
	m.Text = em.Text
	m.DateWritten = em.DateWritten.Format(time.RFC3339)
	m.AuthorID = em.AuthorID.String()
	m.AuthorShortID = em.AuthorID.Shorten()
	m.AuthorName = userName(&em.AuthorID, l)
}

type RootMessage struct {
	Message
	Topic string
}

func (rm *RootMessage) Assign(em *entity.Message, l *dscuss.LoginHandle) {
	rm.Message.Assign(em, l)
	rm.Topic = em.Topic.String()
}

type Thread struct {
	RootMessage
	Replies []Message
}

type ComposedReply struct {
	Subject string
	Text    string
}

type ComposedRootMessage struct {
	Topic   string
	Subject string
	Text    string
}

type Operation struct {
	ID            string
	ShortID       string
	Type          string
	Reason        string
	Comment       string
	DatePerformed string
	AuthorName    string
	AuthorID      string
	AuthorShortID string
}

func (o *Operation) Assign(eo *entity.Operation, l *dscuss.LoginHandle) {
	o.ID = eo.ID().String()
	o.ShortID = eo.ID().Shorten()
	o.Type = eo.OperationType().String()
	o.Reason = eo.Reason.String()
	o.Comment = eo.Comment
	o.DatePerformed = eo.DatePerformed.Format(time.RFC3339)
	o.AuthorID = eo.AuthorID.String()
	o.AuthorShortID = eo.AuthorID.Shorten()
	o.AuthorName = userName(&eo.AuthorID, l)
}

type Peer struct {
	Nickname        string
	ID              string
	LocalAddr       string
	RemoteAddr      string
	AssociatedAddrs string
	Subscriptions   string
	State           string
}

func (p *Peer) Assign(pi *peer.Info) {
	p.Nickname = pi.Nickname
	p.ID = pi.ID
	p.LocalAddr = pi.LocalAddr
	p.RemoteAddr = pi.RemoteAddr
	p.AssociatedAddrs = strings.Join(pi.AssociatedAddrs, ",")
	p.Subscriptions = strings.Join(pi.Subscriptions, "\n")
	p.State = pi.State
}

type PeerHistory struct {
	ID            string
	Disconnected  string
	Subscriptions string
}

func (ph *PeerHistory) Assign(uh *entity.UserHistory) {
	ph.ID = uh.ID.String()
	ph.Disconnected = uh.Disconnected.Format(time.RFC3339)
	if uh.Subs != nil {
		ph.Subscriptions = strings.Join(uh.Subs.StringSlice(), "\n")
	}

}
