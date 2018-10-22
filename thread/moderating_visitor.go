/*
This file is part of Dscuss.
Copyright (C) 2018  Vitaly Minko

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

import (
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
)

type ModeratingVisitor struct {
	handler ModeratingHandler
}

type ModeratingHandler interface {
	Handle(n *Node) (*entity.Message, error)
}

func NewModeratingVisitor(handler ModeratingHandler) *ModeratingVisitor {
	return &ModeratingVisitor{handler: handler}
}

func (mv *ModeratingVisitor) Visit(n *Node) (*Node, error) {
	if n == nil {
		log.Fatal("Bug: attempt to moderate nil node")
	}
	newMsg, err := mv.handler.Handle(n)
	if err != nil {
		return nil, err

	}
	if newMsg == nil {
		return nil, nil
	}
	newNode := New(newMsg)
	for _, c := range n.Children {
		rep, err := mv.Visit(c)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			newNode.AddChild(rep)
		}
	}
	return newNode, nil
}
