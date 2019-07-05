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
)

func handlePeerHistory(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	if len(r.URL.Query()) != 0 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	var hist []PeerHistory
	hh := l.ListUserHistory()
	if len(hh) > 0 {
		for _, hr := range hh {
			hist = append(hist, PeerHistory{})
			h := &hist[len(hist)-1]
			h.Assign(hr)
		}
	}
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Peer history"
	view.Render(w, "peer_history.html", map[string]interface{}{
		"Common":  cd,
		"History": hist,
	})
}
