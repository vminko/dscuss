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

const threadCreateHTML = `
{{ define "content" }}

<h1 id="title">{{ .Common.PageTitle }}</h1>
<form action="/thread/create" method="POST" enctype="multipart/form-data">
	<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
	<table class="form">
		<tr>
			<th>Topic:</th>
			<td><input type="text" name="topic" value="{{ .Common.Topic }}"></td>
		</tr>
		<tr>
			<th>Subject:</th>
			<td><input type="text" name="subject" value="{{ .Subject }}"></td>
		</tr>
		<tr>
			<th>Text:</th>
			<td><textarea name="text" rows="12">{{ .Text }}</textarea></td>
		</tr>
		<tr>
			<th></th>
			<td>
				{{ if .Message }}
					<span class="alert">{{ .Message }}</span><br>
				{{ end }}
				<input type="submit" name="action" class="btn" value="Start">
			</td>
		</tr>
	</table>
</form>

{{ end }}`

/* vim: set filetype=html tabstop=2: */
