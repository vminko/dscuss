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

const baseHTML = `
<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" type="text/css" href="/static/dscuss.css">
	<title>
		{{ if .Common.PageTitle }}
			{{ .Common.PageTitle }}
		{{ else }}
			{{ .Common.NodeName }}
		{{ end }}
	</title>
	{{ block "head" . }}{{ end }}
</head>

<body>
	<div id="header" class="clearfix">
		<div id="navleft">
			<a href="/">{{ .Common.NodeName }}</a>
		</div>
		<div id="navright">
			{{ if .Common.OwnerName }}
			{{ .Common.OwnerName }} ({{ .Common.OwnerID }}) <a href="/logout">Logout</a>
			{{ else }}
			<a href="/login?next={{ .Common.CurrentURL }}">Login</a>
			{{ end }}
		</div>
	</div>
	<hr>
	<div id="container">
		<div id="content">
		{{ block "content" . }}{{ end }}
		</div>
	</div>
	<script src="/static/dscuss.js"></script>
</body>
</html>
`
