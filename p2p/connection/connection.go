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

package connection

import (
	"encoding/json"
	"net"
	"time"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/packet"
)

const (
	Timeout time.Duration = 1 * time.Second
)

// Connection is responsible for transferring packets via the network.
type Connection struct {
	conn                net.Conn
	associatedAddresses []string
	isIncoming          bool
	closeHandler        func(*Connection)
}

func New(conn net.Conn, isIncoming bool) *Connection {
	return &Connection{
		conn:                conn,
		associatedAddresses: []string{conn.RemoteAddr().String()},
		isIncoming:          isIncoming,
	}
}

func (c *Connection) Read() (*packet.Packet, error) {
	c.conn.SetDeadline(time.Now().Add(Timeout))
	d := json.NewDecoder(c.conn)
	var p packet.Packet
	err := d.Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Connection) Write(p *packet.Packet) error {
	c.conn.SetDeadline(time.Now().Add(Timeout))
	e := json.NewEncoder(c.conn)
	return e.Encode(p)
}

func (c *Connection) AssociatedAddresses() []string {
	return c.associatedAddresses
}

func (c *Connection) RegisterCloseHandler(f func(*Connection)) {
	if c.closeHandler != nil {
		log.Fatal("Attempt to overwrite closeHandler")
	}
	c.closeHandler = f
}

func (c *Connection) IsIncoming() bool {
	return c.isIncoming
}

func (c *Connection) Close() {
	c.conn.Close()
	if c.closeHandler != nil {
		c.closeHandler(c)
	}
}
