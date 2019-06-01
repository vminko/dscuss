package view

const loginHTML = `
{{ define "content" }}

<form action="/login" method="POST">
<input type="hidden" name="csrf" value="{{ .Common.CSRF }}">
<input type="hidden" name="next" value="{{ .next }}">
<table class="form">
	<tr>
		<th></th>
		<td>Owner authentication</td>
	</tr>
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
		<td>How to setup your own Dscuss node: <a href="/help?=setup">Help</a></td>
	</tr>
{{ if .Message }}
	<tr>
		<th></th>
		<td><span class="alert">{{ .Message }}</span></td>
	</tr>
{{ end }}
	<tr>
		<th></th>
		<td><input type="submit" value="Login"></td>
	</tr>
</table>
</form>

{{ end }}`

/* vim: set filetype=html: */
