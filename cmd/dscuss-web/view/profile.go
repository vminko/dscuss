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

const profileHTML = `
{{ define "content" }}

<h2 class="title">Node owner's profile</h1>
<table class="form">
	<tr><th>Full ID</th><td>{{ .Common.OwnerID }}</td></tr>
	<tr><th>Nickname</th><td>{{ .Common.OwnerName }}</td></tr>
	<tr><th>Additional info</th><td>{{ .Common.OwnerInfo }}</td></tr>
	<tr><th>Registration date</th><td>{{ .Common.OwnerRegDate }}</td></tr>
</table>
<hr class="sep">
<span class="subtitle">Subscriptions</span>
<form action="/sub" method="POST" enctype="multipart/form-data">
<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
<table>
	<tr>
		<th>Topic</th>
		<th>Action</th>
	</tr>
{{ range .Subscriptions }}
	<tr>
		<td>{{ . }}</td>
		<td><a href="/unsub?id={{ . }}">Remove</a></td>
	</tr>
{{ end }}
	<tr>
		<th>
			<input type="text" name="topic" placeholder="Enter new topic...">
		</th>
		<th>
			<input type="submit" name="action" class="no-double-post" value="Subscribe">
		</th>
	</tr>
</table>
</form>
<hr class="sep">
<span class="subtitle">Moderators</span>
<form action="/mkmdr" method="POST" enctype="multipart/form-data">
<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
<table>
	<tr>
		<th>User</th>
		<th>Action</th>
	</tr>
{{ range .Moderators }}
	<tr>
		<td><a href="/profile?id={{ .ID }}">{{ .Nickname }}-{{ .ShortID }}</a></td>
		<td><a href="/rmmdr?id={{ .ID }}">Remove</a></td>
	</tr>
{{ end }}
	<tr>
		<td>
			<input type="text" name="id" placeholder="Enter new full ID...">
		</td>
		<td>
			<input type="submit" name="action" class="no-double-post" value="Add">
		</td>
	</tr>
</table>
</form>
{{ if .Message }}
	<span class="alert">{{ .Message }}</span><br>
{{ end }}
{{ end }}`

/* vim: set filetype=html: */
