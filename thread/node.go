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

import (
	"vminko.org/dscuss/entity"
)

type Node struct {
	Msg      *entity.Message
	Children []*Node
	Parent   *Node
}

func New(m *entity.Message) *Node {
	return &Node{Msg: m}
}

func (n *Node) AddReply(m *entity.Message) *Node {
	child := New(m)
	n.Children = append(n.Children, child)
	child.Parent = n
	return child
}

func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
	child.Parent = n
}

func (n *Node) Depth() int {
	depth := 0
	for tmp := n; tmp.Parent != nil; depth++ {
		tmp = tmp.Parent
	}
	return depth
}

func (n *Node) IsRoot() bool {
	return n.Depth() == 0
}

func (n *Node) View(visitor *ViewingVisitor) {
	visitor.Visit(n)
}

func (n *Node) Moderate(visitor *ModeratingVisitor) (*Node, error) {
	return visitor.Visit(n)
}
