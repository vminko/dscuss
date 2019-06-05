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

const replyHTML = `
{{ define "content" }}

<h2 id="title">Reply to <a href="/thread?id={{ .Thread.ID }}">{{ .Thread.Subject }}</a></h2>
<span class="topic">in <a href="/board?topic={{ .Thread.Topic }}">{{ .Thread.Topic }}</a></span>
{{ if .ShowParentSubject }}
<b>{{ .Parent.Subject }}</b>
{{ end }}
<div class="message-text">{{ .Parent.Text }}</div>
<div class="dimmed underline">
	by <a href="/user?u={{ .Parent.AuthorID }}">{{ .Parent.AuthorName }}-{{ .Parent.AuthorShortID }}</a>
	{{ .Parent.DateWritten }}
</div>
<div class="row">
	<form action="/reply" method="POST" enctype="multipart/form-data">
	<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
	<input type="hidden" name="id" value="{{ .Parent.ID }}">
	Subject: <input type="text" name="subject" value="{{ .Reply.Subject }}" placeholder="Re: {{.Parent.Subject}}">
	<textarea name="text" rows="12">{{ .Reply.Text }}</textarea>
	{{ if .Message }}
		<span class="alert">{{ .Message }}</span><br>
	{{ end }}
	<input type="submit" name="action" class="no-double-post" value="Submit reply">
	</form>
</div>

{{ end }}`

/* vim: set filetype=html: */
