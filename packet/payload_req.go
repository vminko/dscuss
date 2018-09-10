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
	"vminko.org/dscuss/entity"
)

// PayloadReq is used for requesting advertised entities.
type PayloadReq struct {
	ID entity.ID `json:"new_id"` // Id of the entity being requested
}

func NewPayloadReq(id *entity.ID) *PayloadReq {
	p := &PayloadReq{}
	copy(p.ID[:], id[:])
	return p
}
