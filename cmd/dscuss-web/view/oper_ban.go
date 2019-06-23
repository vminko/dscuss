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

const operBanHTML = `
{{ define "content" }}

<h2 class="title">Banning user {{ .Target.Nickname }}-{{ .Target.ShortID }}</h2>
<form action="/ban" method="POST" enctype="multipart/form-data">
<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
<input type="hidden" name="id" value="{{ .Target.ID }}">
<table class="form">
	<tr><th>Full ID</th><td>{{ .Target.ID }}</td></tr>
	<tr><th>Nickname</th><td>{{ .Target.Nickname }}</td></tr>
	<tr><th>Additional info</th><td>{{ .Target.Info }}</td></tr>
	<tr><th>Registration date</th><td>{{ .Target.RegDate }}</td></tr>
	<tr>
		<th>Reason:</th>
		<td>
			<select name="reason" >
				<option value="SPAM">SPAM</option>
	    			<option value="Offtopic">Off-topic</option>
	    			<option value="Abuse">Abuse</option>
	    			<option value="Duplicate">Duplicate</option>
			</select>
		</td>
	</tr>
	<tr>
		<th>Comment:</th>
		<td><textarea name="comment" rows="4" placeholder="Why do you want to do that?">{{ .Reply.Text }}</textarea></td>
	</tr>
	<tr>
		<th></th>
		<td>
			{{ if .Message }}
				<span class="alert">{{ .Message }}</span><br>
			{{ end }}
			<input type="submit" name="action" class="no-double-post" value="Submit reply">
		</td>
	</tr>

</table>
</form>

{{ end }}`

/* vim: set filetype=html: */
