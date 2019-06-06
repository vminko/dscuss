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

const startHTML = `
{{ define "content" }}

<h2 id="title">Start new dscussion</h2>
<div class="row">
	<form action="/start" method="POST" enctype="multipart/form-data">
		<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
		Topic: <input type="text" name="topic" value="{{ .Topic }}"><br>
		Subject: <input type="text" name="subject" value="{{ .Subject }}">
		<textarea name="text" rows="12">{{ .Text }}</textarea>
		{{ if .Message }}
			<span class="alert">{{ .Message }}</span><br>
		{{ end }}
		<input type="submit" name="action" class="no-double-post" value="Start dscussion">
	</form>
</div>

{{ end }}`

/* vim: set filetype=html: */
