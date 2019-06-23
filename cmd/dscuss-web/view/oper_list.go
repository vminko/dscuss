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

const operListHTML = `
{{ define "content" }}

<h1 id="title">{{ .Common.PageTitle }}</h1>
{{ if .Operations }}
	{{ range .Operations }}
	<hr class="sep">
	<div class="operation-row" id="operation-{{ .ID }}">
		<div>
			<b>Operation {{ .Type }}</b><br>
			Reason: {{ .Reason }}<br>
		</div>
		{{ if .Comment }}
			<div class="comment">{{ .Comment }}</div>
		{{ end }}
		<div class="dimmed underline">
			by <a href="/user?id={{ .AuthorID }}">{{ .AuthorName }}-{{ .AuthorShortID }}</a>
			{{ .DatePerformed }}
		</div>
	</div>
{{ end }}
{{ else }}
	<div class="row">
		<div class="dimmed">No operations to show.</div>
	</div>
{{ end }}

{{ end }}
`

/* vim: set filetype=html tabstop=2: */
