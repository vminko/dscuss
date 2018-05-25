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
	"time"
	"vminko.org/dscuss/entity"
)

// PayloadHello is used for handshaking.
// When user A sends this packet to user B, he/she:
// 1. notifies user B about topics of A's interests;
// 2. proves that the user A actually has the A's private key;
// Time and ReceiverID protect from replay attack.
type PayloadHello struct {
	// Id of the user this payload is designated for.
	ReceiverID entity.ID `json:"receiver_id"`
	// Date and time when the payload was composed.
	Time time.Time `json:"time"`
	// Subscriptions of the author of the payload.
	// TBD: subscriptions []Subscription;
}
