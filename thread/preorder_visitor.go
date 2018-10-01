/*
This file is part of Dscuss.
Copyright (C) 2017-2018  Vitaly Minko

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

package thread

type PreOrderVisitor struct {
	handler Handler
}

func NewPreOrderVisitor(handler Handler) *PreOrderVisitor {
	return &PreOrderVisitor{handler: handler}
}

func (v *PreOrderVisitor) Visit(n *Node) bool {
	if n == nil {
		return true
	}
	if !v.handler.Handle(n) {
		return false
	}
	for _, r := range n.Replies {
		if !v.Visit(r) {
			return false
		}
	}
	return true
}
