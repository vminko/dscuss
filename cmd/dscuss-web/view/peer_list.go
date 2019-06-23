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

const peerListHTML = `
{{ define "content" }}

<h1>{{ .Common.PageTitle }}</h1>

{{ if .Peers }}
{{ range .Peers }}
<hr class="sep">
<div class="peer-row" id="peer-{{ .ID }}">
<table class="form">
	<tr><th>Nickname</th><td>{{ .Nickname }}</td></tr>
	<tr><th>ID</th><td>{{ .ID }}</td></tr>
	<tr><th>State</th><td>{{ .State }}</td></tr>
	<tr><th>Local address</th><td>{{ .LocalAddr }}</td></tr>
	<tr><th>Remove address</th><td>{{ .RemoteAddr }}</td></tr>
	<tr><th>Associated addresses</th><td>{{ .AssociatedAddrs }}</td></tr>
	<tr>
		<th>Subscriptions</th>
		<td><div class="subs">{{ .Subscriptions }}</div></td>
	</tr>
</table>
</div>
{{ end }}
{{ else }}
<div class="row">
	<div class="dimmed">There are no peers connected.</div>
</div>
{{ end }}
<div><a href="/peer/history">Show peer history</a>.</div>
{{ end }}
`

/* vim: set filetype=html: */
