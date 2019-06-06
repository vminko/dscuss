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

{{ if .Common.IsWritingPermitted }}
<div class="btn-row">
	<a class="link-btn" href="/start{{ if $.Topic }}?topic={{ $.Topic }}{{ end }}">Start</a>
</div>
{{end}}

<h1> Dscussions in
{{ if .Topic }}
topic {{ .Topic }}
{{ else }}
all topics
{{ end }}
</h1>

{{ if .Threads }}
{{ range .Threads }}
<hr class="sep">
<div class="thread-row" id="thread-{{ .ID }}">
	<div>
		<a href="/thread?id={{ .ID }}">{{ .Subject }}</a>
		{{ if not $.Topic }}
			<span class="topic">in <a class="topic" href="/board?topic={{ .Topic }}">{{ .Topic }}</a></span>
		{{ end }}
	</div>
	<div class="message-text">{{ .Text }}</div>
	<div class="dimmed underline">
		by {{ .AuthorName }}-{{ .AuthorShortID }} {{ .DateWritten }}
	</div>
</div>
{{ end }}
{{ else }}
<div class="row">
	<div class="dimmed">No threads to show.</div>
</div>
{{ end }}

{{ end }}
`

/* vim: set filetype=html: */
