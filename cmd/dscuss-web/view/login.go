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

const loginHTML = `
{{ define "content" }}

<h1 id="title">{{ .Common.PageTitle }}</h1>
<form action="/login" method="POST">
	<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
	<input type="hidden" name="next" value="{{ .next }}">
	<table class="form">
		<tr>
			<th>Username:</th>
			<td><input type="text" name="username" required></td>
		</tr>
		<tr>
			<th>Password:</th>
			<td><input type="password" name="password" required></td>
		</tr>
		<tr>
			<th></th>
			<td>How to setup your own Dscuss node: <a href="http://vminko.org/dscuss/setup">Help</a></td>
		</tr>
		{{ if .Message }}
			<tr>
				<th></th>
				<td><span class="alert">{{ .Message }}</span></td>
			</tr>
		{{ end }}
		<tr>
			<th></th>
			<td><input type="submit" class="btn" value="Login"></td>
		</tr>
	</table>
</form>

{{ end }}`

/* vim: set filetype=html tabstop=2: */
