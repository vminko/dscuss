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
	<a class="link-btn" href="/reply?id={{ .ThreadID }}">Reply</a>
</div>
{{end}}

<h2 id="title"><a href="/thread?id={{ .ThreadID }}">{{ .Subject }}</a></h2>
<span class="topic">in <a href="/board?topic={{ .Topic }}">{{ .Topic }}</a></span>
<div class="message-row">
	<div class="message-text">{{ .Text }}</div>
	<div class="dimmed underline">
		by <a href="/user?u={{ .AuthorID }}">{{ .AuthorName }}-{{ .AuthorShortID }}</a> {{ .DateWritten }}
		{{ if .Common.IsWritingPermitted }}
		| <a href="/ban?id={{ .AuthorID }}">ban</a>
		| <a href="/rmmsg?id={{ .ID }}">delete</a>
		| <a href="/reply?id={{ $.ID }}">reply</a>
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
		| <a href="/reply?id={{ $.ID }}">reply</a>
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

{{ if .Common.IsWritingPermitted }}
<div style="margin-top: 40px;">
<form action="/reply" method="POST" enctype="multipart/form-data">
	<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
	<input type="hidden" name="id" value="{{ .ThreadID }}">
	<textarea name="content" rows="12" placeholder="Your reply..."></textarea>
	<input type="submit" name="action" class="no-double-post" value="Post reply">
</form>
</div>
{{ end }}

{{ end }}`

/* vim: set filetype=html: */
