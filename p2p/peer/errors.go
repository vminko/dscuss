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

package peer

import (
	"fmt"
	"vminko.org/dscuss/entity"
)

type needIDError struct {
	ID *entity.ID
}

func (e *needIDError) Error() string {
	return fmt.Sprintf("need entity with ID %s", e.ID.Shorten())
}

type banSenderError struct {
}

func (e *banSenderError) Error() string {
	return fmt.Sprintf("ban sender of the entity")
}

type banIDError struct {
	ID *entity.ID
}

func (e *banIDError) Error() string {
	return fmt.Sprintf("ban user with ID %s", e.ID.Shorten())
}

type bannedError struct {
}

func (e *bannedError) Error() string {
	return fmt.Sprintf("author of parent entity is banned")
}

type skipError struct {
}

func (e *skipError) Error() string {
	return fmt.Sprintf("skip the entity")
}
