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

package packet

import (
	"vminko.org/dscuss/subs"
)

// PayloadHello is used for introducing peers during handshake.
// When user A sends this packet to user B, he/she
// notifies user B about topics of A's interests/
type PayloadHello struct {
	Proto int                `json:"proto"` // The version of the protocol this peer supports.
	Subs  subs.Subscriptions `json:"subs"`  // Subscriptions of the author of the payload.
}

func (p *PayloadHello) IsValid() bool {
	return p.Subs.IsValid()
}

func NewPayloadHello(p int, s subs.Subscriptions) *PayloadHello {
	return &PayloadHello{Proto: p, Subs: s}
}
