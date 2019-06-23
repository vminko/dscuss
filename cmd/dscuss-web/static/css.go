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
package static

// CSS is based on style sheets from https://github.com/s-gv/orangeforum

const CSS = `
* {
	-webkit-box-sizing: border-box;
	-moz-box-sizing: border-box;
	box-sizing: border-box;
}
html, body {
	margin: 0;
	height: 100%;
}
#container {
	max-width: 960px;
	line-height: 1.58;
	margin: 0 auto;
	min-height: 100%;
	position: relative;
}
#header {
	padding-top: 10px;
}
#content {
	clear: both;
	padding-top: 20px;
	padding-bottom: 75px;
}
#footer {
	position: absolute;
	bottom: 0;
	width: 100%;
	height: 60px;
	text-align: center;
}
.clearfix {
	overflow: auto;
}
body {
	font-family: Arial, "Helvetica Neue", Helvetica, sans-serif;
	text-rendering: optimizeLegibility;
	-webkit-font-smoothing: antialiased;
	padding-left: 10px;
	padding-right: 10px;
}
img {
	max-width: 100%;
	display: block;
	margin: 0 auto;
}
a {
	text-decoration: none;
}
a:link {
	color: #07C;
}
a:hover, a:active {
	color: #3af;
}
a:visited {
	color: #005999;
}
#header a, #header a:link, #header a:hover, #header a:active, #header a:visited {
	color: #000;
}
#title a, #title a:link, #title a:hover, #title a:active, #title a:visited {
	color: #000;
}
#footer a, #footer a:link, #footer a:hover, #footer a:active, #footer a:visited {
	color: grey;
}
a.headline {
	padding: 0px 5px;
}
.link-btn, .link-btn:link, .link-btn:visited {
	color: white;
	background: #07C;
	padding: 10px 20px;
	text-align: center;
	width: 150px;
	margin-left: 20px;
	font-size: 16px;
}
.link-btn:hover {
	background: #3af;
}
.btn {
	padding: 10px 20px;
	background: #07C;
	font-size: 16px;
	color: white;
	border: none;
	margin-left: 20px;
	width: 150px;
	line-height: inherit;
}
.btn:hover {
	background: #3af;
	cursor: pointer;
}
.btn-row form, .btn-row a {
	display: inline-block;
}
.btn-row {
	text-align: right;
}
@media (max-width: 599px) {
	.btn {
		font-size: 12px;
		padding: 5px 10px;
		margin: 10px;
		width: 100px;
	}
	.link-btn, .link-btn:link, .link-btn:visited {
		font-size: 12px;
		padding: 5px 10px;
		margin: 10px;
		width: 100px;
	}
	.btn-row {
		text-align: center;
	}
}
#navleft {
	float: left;
	max-width: 70%;
}
#navright {
	float: right;
}
.dimmed {
	color: darkgrey;
}
.dimmed a, .dimmed a:link, .dimmed a:hover, .dimmed a:visited, .dimmed a:active {
	color: grey;
}
.row {
	margin-top: 20px;
}
th, td {
	text-align: left;
}
@media (min-width: 600px) {
	.form td {
		width: 360px;
	}
}
.form th {
	text-align: right;
	vertical-align: top;
	padding-top: 10px;
}
.form td {
	padding-top: 10px;
}
@media (max-width: 599px) {
	table.form {
		width: 100%;
	}
	.form th {
		display: block;
		float: left;
		text-align: left;
		max-width: 75%;
		padding-top: 20px;
	}
	.form td {
		display: block;
		text-align: right;
		padding-top: 20px;
	}
	.form input[type="submit"] {
		width: 100%;
	}
}
input[type="text"], input[type="number"], input[type="password"] {
	width: 100%;
}
textarea {
	width: 100%;
}
.sep {
	border: none;
	height: 1px;
	background-color: #e9e9e9;
}
#title {
	margin-bottom: 0px;
}
.subtitle {
	font-weight: bold;
}
.message-text, .comment {
	margin-top: 10px;
	white-space: pre-wrap;
	font-family: monospace;
}
.subs {
	white-space: pre-wrap;
}
.message-row, .thread-row, .operation-row, .peer-row, .history-row, .profile-block {
	margin-bottom: 30px;
}
.underline, .topic {
	font-size: 75%;
}
.alert {
	color: red;
}
a, .dimmed, h3, .message p {
	word-wrap: break-word;
}

.dimmed .link-button {
	border: none;
  	outline: none;
  	background: none;
  	cursor: pointer;
  	color: grey;
  	padding: 0;
  	text-decoration: none;
  	font-family: inherit;
  	font-size: inherit;
}
.dimmed .link-button:focus {
	outline: none;
}
.dimmed .link-button:active {
	color: grey;
}
`

/* vim: set filetype=css: */
