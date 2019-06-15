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

// View presents the controller responds in HTML format.
package view

import (
	"html/template"
	"net/http"
	"vminko.org/dscuss/log"
)

type templateMap map[string]*template.Template

func (tm templateMap) Add(baseTmpl *template.Template, tmplName string, tmplSrc string) {
	fileName := tmplName + ".html"
	templates[fileName] = template.Must(baseTmpl.Clone())
	template.Must(templates[fileName].New(tmplName).Parse(tmplSrc))
}

var templates templateMap = make(map[string]*template.Template)

func init() {
	base := template.Must(template.New("base").Parse(baseHTML))
	templates.Add(base, "login", loginHTML)
	templates.Add(base, "board", boardHTML)
	templates.Add(base, "thread", threadHTML)
	templates.Add(base, "reply", replyHTML)
	templates.Add(base, "start", startHTML)
	templates.Add(base, "profile", profileHTML)
	templates.Add(base, "rmmsg", rmmsgHTML)
	templates.Add(base, "ban", banHTML)
}

func Render(w http.ResponseWriter, tmplName string, data interface{}) {
	err := templates[tmplName].Execute(w, data)
	if err != nil {
		log.Errorf("Error rendering %s: %s\n", tmplName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
