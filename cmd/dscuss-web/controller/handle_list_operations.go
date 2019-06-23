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

func handleListOperations(w http.ResponseWriter, r *http.Request, l *dscuss.LoginHandle, s *Session) {
	if len(r.URL.Query()) != 2 {
		BadRequestHandler(w, r, "Wrong number of query parameters")
		return
	}
	idStr := r.URL.Query().Get("id")
	var id entity.ID
	err := id.ParseString(idStr)
	if err != nil {
		BadRequestHandler(w, r, "'"+idStr+"' is not a valid entity ID.")
		return
	}
	entType := r.URL.Query().Get("type")
	var operEntities []*entity.Operation
	switch entType {
	case "user":
		operEntities, err = l.ListOperationsOnUser(&id)
	case "msg":
		operEntities, err = l.ListOperationsOnMessage(&id)
	default:
		BadRequestHandler(w, r, "'"+entType+"' is not a valid entity type.")
		return
	}
	if err == errors.NoSuchEntity {
		NotFoundHandler(w, r)
	} else if err != nil {
		panic("Error fetching operations: " + err.Error())
	}

	var ops []Operation
	for _, oe := range operEntities {
		ops = append(ops, Operation{})
		o := &ops[len(ops)-1]
		o.Assign(oe, l)
	}
	cd := readCommonData(r, s, l)
	var t string
	switch entType {
	case "user":
		t = "user"
	case "msg":
		t = "message"
	}
	cd.PageTitle = "Operations on " + t + " #" + id.Shorten()
	view.Render(w, "oper_list.html", map[string]interface{}{
		"Common":     cd,
		"Operations": ops,
	})
}
