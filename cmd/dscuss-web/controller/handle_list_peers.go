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
	"sort"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/view"
)

func handleListPeers(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if !s.IsAuthenticated {
		ForbiddenHandler(w, r)
		return
	}
	if len(r.URL.Query()) != 0 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	var peers []Peer
	peerEntities := l.ListPeers()
	if len(peerEntities) > 0 {
		sort.Sort(dscuss.ByNickname(peerEntities))
		for _, pe := range peerEntities {
			peers = append(peers, Peer{})
			p := &peers[len(peers)-1]
			p.Assign(pe)
		}
	}
	cd := readCommonData(r, s, l)
	cd.PageTitle = "Connected peers"
	view.Render(w, "peer_list.html", map[string]interface{}{
		"Common": cd,
		"Peers":  peers,
	})
}
