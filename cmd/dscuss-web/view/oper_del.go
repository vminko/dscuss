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

const operDelHTML = `
{{ define "content" }}

<h2 class="title">Removing message {{ .Target.ShortID }}</h2>
<form action="/oper/del" method="POST" enctype="multipart/form-data">
<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
<input type="hidden" name="id" value="{{ .Target.ID }}">
<table class="form">
	<tr>
		<td colspan="2">
			<b>{{ .Target.Subject }}</b>
			<div class="message-text">{{ .Target.Text }}</div>
			<div class="dimmed underline">
				by <a href="/user?id={{ .Target.AuthorID }}">{{ .Target.AuthorName }}-{{ .Target.AuthorShortID }}</a>
				{{ .Target.DateWritten }}
			</div>
		</td>
	</tr>
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
