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
	"sync"
	"time"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/p2p/connection"
)

// peer is responsible for communication with other nodes.
// Implements the Dscuss protocol.
type Peer struct {
	Conn      *connection.Connection
	closeChan chan *Peer
	stopChan  chan struct{}
	wg        *sync.WaitGroup
	state     State
}

func New(
	conn *connection.Connection,
	closeChan chan *Peer,
	stopChan chan struct{},
	wg *sync.WaitGroup,
) *Peer {
	p := &Peer{
		Conn:      conn,
		closeChan: closeChan,
		stopChan:  stopChan,
		wg:        wg,
	}
	go p.run()
	return p
}

func (p *Peer) run() {
	defer p.Conn.Close()
	defer p.wg.Done()
	pulser := time.NewTicker(time.Second * 3)
	defer pulser.Stop()
	for {
		select {
		case <-p.stopChan:
			log.Debug("Stop requested")
			return
		case <-pulser.C:
			log.Debug("Peer is idle...")
		}
	}
}
