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
	"fmt"
	"net"
	"sync"
	"time"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/packet"
)

const (
	DefaultTimeout time.Duration = 5 * time.Second
	// Limits the size of a packet. Anything larger will cause
	// a decoding error.
	MaxPacketSize int = 10 * 1024
)

// Connection is responsible for transferring packets via the network.
type Connection struct {
	conn         net.Conn
	addresses    []string
	addrMx       sync.RWMutex
	isIncoming   bool
	closeHandler func(*Connection)
}

func New(conn net.Conn, isIncoming bool) *Connection {
	return &Connection{
		conn:       conn,
		addresses:  []string{conn.RemoteAddr().String()},
		isIncoming: isIncoming,
	}
}

func fixErrClosedConnection(err error) error {
	closedConnText := "use of closed network connection"
	if e, ok := err.(*net.OpError); ok {
		if e.Err.Error() == closedConnText {
			return errors.ClosedConnection
		}
	}
	return err
}

func (c *Connection) ReadFull(timeout time.Duration) (*packet.Packet, error) {
	c.conn.SetDeadline(time.Now().Add(timeout))
	d := json.NewDecoder(limitReader(c.conn, MaxPacketSize))
	d.DisallowUnknownFields()
	var p packet.Packet
	err := d.Decode(&p)
	if err != nil {
		return nil, fixErrClosedConnection(err)
	}
	log.Debugf("Received this packet from %s: %s", c.RemoteAddr(), p.Dump())
	return &p, nil
}

func (c *Connection) Read() (*packet.Packet, error) {
	return c.ReadFull(DefaultTimeout)
}

func (c *Connection) WriteFull(p *packet.Packet, timeout time.Duration) error {
	log.Debugf("Sending this packet to %s: %s", c.RemoteAddr(), p.Dump())
	c.conn.SetDeadline(time.Now().Add(timeout))
	e := json.NewEncoder(limitWriter(c.conn, MaxPacketSize))
	return fixErrClosedConnection(e.Encode(p))
}

func (c *Connection) Write(p *packet.Packet) error {
	return c.WriteFull(p, DefaultTimeout)
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

// Addresses returns a copy of address list associated with the connection.
func (c *Connection) Addresses() []string {
	c.addrMx.RLock()
	defer c.addrMx.RUnlock()
	res := make([]string, len(c.addresses))
	copy(res, c.addresses)
	return res
}

func (c *Connection) AddAddresses(new []string) {
	c.addrMx.Lock()
	defer c.addrMx.Unlock()
	c.addresses = append(c.addresses, new...)
}

func (c *Connection) ClearAddresses() {
	c.addrMx.Lock()
	defer c.addrMx.Unlock()
	c.addresses = nil
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

func (c *Connection) IsActive() bool {
	return !c.isIncoming
}

func (c *Connection) String() string {
	return fmt.Sprintf("inc=%t, %s", c.isIncoming, c.RemoteAddr())
}

func (c *Connection) Close() {
	c.conn.Close()
	if c.closeHandler != nil {
		c.closeHandler(c)
	}
}
