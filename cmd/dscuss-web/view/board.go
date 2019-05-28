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

package view

const boardHTML = `
{{ define "content" }}

<h1> Dscussions in
{{ if .Topic }}
topic {{ .Topic }}
{{ else }}
all topics
{{ end }}
</h1>

{{ if .Threads }}
{{ range .Threads }}
<div class="thread-row" id="thread-{{ .ID }}">
	<div><a href="/thread?id={{ .ID }}">{{ .Subject }}</a></div>
	<div class="muted">
		ID: {{ .ID }}
	</div>
	<div class="muted">
		Topic: <a href="/board?topic={{ .Topic }}">{{ .Topic }}</a>
	</div>
	<div class="muted">
		by {{ .AuthorName }}" ( {{ .AuthorID }} ) at {{ .DateWritten }}
	</div>
</div>
<hr class="sep">
{{ end }}
{{ else }}
<div class="row">
	<div class="muted">No threads to show.</div>
</div>
{{ end }}

{{ end }}
`
