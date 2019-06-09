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

// JavaScript is based on js from https://github.com/s-gv/orangeforum

const JavaScript = `
var inputs = document.getElementsByClassName("no-double-post");
for (var i = 0; i < inputs.length; i++) {
	if (inputs[i].type == "submit") {
		var btn = inputs[i];
		var form = btn.form;
		if (!!form && form.method == "post") {
			btn.onclick = function() {
				var form = this.form;
				if(!!form && form.checkValidity()) {
					var submitBtn = this;
				 	setTimeout(function() {
				 		submitBtn.disabled = true;
				 	}, 1);
				}
			};
		}
	}
}
`

/* vim: set filetype=javascript: */
