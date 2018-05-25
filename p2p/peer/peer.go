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
	"sync"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/p2p/connection"
)

// Peer is responsible for communication with other nodes.
// Implements the Dscuss protocol.
type Peer struct {
	Conn             *connection.Connection
	closeChan        chan *Peer
	stopChan         chan struct{}
	internalStopChan chan struct{}
	outEntityChan    chan *entity.Entity
	wg               *sync.WaitGroup
	internalWG       sync.WaitGroup
	state            State
	user             *entity.User
}

func New(
	conn *connection.Connection,
	closeChan chan *Peer,
	stopChan chan struct{},
	wg *sync.WaitGroup,
) *Peer {
	p := &Peer{
		Conn:             conn,
		closeChan:        closeChan,
		stopChan:         stopChan,
		internalStopChan: make(chan struct{}),
		outEntityChan:    make(chan *entity.Entity),
		wg:               wg,
		state:            newStateHandshaking(),
	}
	// TBD; storage.AddEntityConsumer(p))
	go p.run()
	return p
}

func (p *Peer) watchStop() {
	defer p.internalWG.Done()
	select {
	case <-p.stopChan:
		log.Debugf("Stop requested for peer %s, closing Conn", p.Desc())
		p.Conn.Close()
		return
	case <-p.internalStopChan:
		log.Debugf("Peer %s: run requested to stop watchStop()", p.Desc())
		return
	}
}

func (p *Peer) run() {
	p.internalWG.Add(1)
	go p.watchStop()
	for {
		nextState, err := p.state.Perform(p)
		if err != nil {
			log.Errorf("Error performing '%s' state: %v", p.state.Name(), err)
			close(p.internalStopChan)
			break
		}
		log.Debugf("Switching peer %s to state %s", p.Desc(), nextState.Name())
		p.state = nextState
	}
	// Finalize peer's internal facilities
	p.internalWG.Wait()
	p.Conn.Close()
	// Make external notifications
	p.closeChan <- p
	p.wg.Done()
}

func (p *Peer) Desc() string {
	if p.state.ID() != StateIDHandshaking {
		return fmt.Sprintf("%s-%s", p.user.Nickname(), p.user.ShortID())
	} else {
		return fmt.Sprintf("(not handshaked), %s", p.Conn.RemoteAddress())
	}
}

func (p *Peer) EntityReceived(e *entity.Entity) {
	log.Debugf("Peer %s received entity %s from the Storage", p.Desc(), e.Desc())
	p.outEntityChan <- e
}
