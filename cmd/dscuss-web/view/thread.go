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

const threadHTML = `
{{ define "content" }}

{{ if .Common.IsWritingPermitted }}
<div class="btn-row">
	<a class="link-btn" href="/reply?id={{ .ID }}">Reply</a>
</div>
{{end}}

<h2 class="title"><a href="/thread?id={{ .ID }}">{{ .Subject }}</a></h2>
<div class="message-row">
	<div class="message-text">{{ .Text }}</div>
	<div class="dimmed underline">
		by <a href="/user?u={{ .AuthorID }}">{{ .AuthorName }}-{{ .AuthorShortID }}</a> {{ .DateWritten }}
		{{ if .Common.IsWritingPermitted }}
		| <a href="/ban?id={{ .AuthorID }}">ban</a>
		| <a href="/rmmsg?id={{ .ID }}">delete</a>
		| <a href="/reply?id={{ .ID }}">reply</a>
		{{ end }}
	</div>
</div>

{{ if .Replies }}
{{ range .Replies }}
<hr class="sep">
<div class="message-row" id="message-{{ .ID }}">
	<b>{{ .Subject }}</b>
	<div class="message-text">{{ .Text }}</div>
	<div class="dimmed underline">
		by <a href="/user?u={{ .AuthorID }}">{{ .AuthorName }}-{{ .AuthorShortID }}</a>
		<a href="/lsop?id={{ .ID }}">{{ .DateWritten }}</a>
		{{ if $.Common.IsWritingPermitted }}
		| <a href="/ban?id={{ .AuthorID }}">ban</a>
		| <a href="/rmmsg?id={{ .ID }}">delete</a>
		| <a href="/reply?id={{ .ID }}">reply</a>
		{{ end }}
	</div>
</div>
{{ end }}
{{ else }}
<div class="row">
	<div class="dimmed">No replies to show.</div>
</div>
{{ end }}
<div id="comment-last"></div>

{{ end }}`

/* vim: set filetype=html: */
