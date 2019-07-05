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
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"vminko.org/dscuss"
	"vminko.org/dscuss/cmd/dscuss-web/config"
)

type CommonData struct {
	CSRF               string
	Owner              User
	NodeName           string
	PageTitle          string
	Topic              string
	CurrentURL         template.URL
	ShowLogin          bool
	IsWritingPermitted bool
}

func readCommonData(r *http.Request, s *Session, l *dscuss.LoginHandle) *CommonData {
	cfg := config.New()
	res := CommonData{
		CSRF:               s.CSRFToken,
		NodeName:           cfg.NodeName,
		CurrentURL:         "/",
		IsWritingPermitted: false,
		ShowLogin:          true,
	}
	if s.IsAuthenticated {
		u := l.GetLoggedUser()
		res.Owner.Assign(u)
		res.IsWritingPermitted = true
	}
	if r.URL.Path != "" {
		currentURL := r.URL.Path
		if r.URL.RawQuery != "" {
			currentURL = currentURL + "?" + r.URL.RawQuery
		}
		res.CurrentURL = template.URL(url.QueryEscape(currentURL))
		res.ShowLogin = !strings.HasPrefix(currentURL, "/login")
	}
	return &res
}
