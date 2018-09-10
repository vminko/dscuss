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
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p/connection"
	"vminko.org/dscuss/storage"
)

type ID entity.ID

var ZeroID ID

// Peer is responsible for communication with other nodes.
// Implements the Dscuss protocol.
type Peer struct {
	Conn          *connection.Connection
	owner         *owner.Owner
	storage       *storage.Storage
	validate      Validator
	goneChan      chan *Peer
	stopChan      chan struct{}
	outEntityChan chan entity.Entity
	wg            sync.WaitGroup
	State         State
	User          *entity.User
}

type Validator func(*Peer) bool

func New(
	conn *connection.Connection,
	owner *owner.Owner,
	storage *storage.Storage,
	validate Validator,
	goneChan chan *Peer,
) *Peer {
	p := &Peer{
		Conn:          conn,
		owner:         owner,
		storage:       storage,
		validate:      validate,
		goneChan:      goneChan,
		stopChan:      make(chan struct{}),
		outEntityChan: make(chan entity.Entity),
	}
	p.State = newStateHandshaking(p)
	p.storage.AttachObserver(p.outEntityChan)
	p.wg.Add(2)
	go p.run()
	go p.watchStop()
	return p
}

func (p *Peer) Close() {
	log.Debugf("Close requested for peer %s", p.Desc())
	p.storage.DetachObserver(p.outEntityChan)
	close(p.stopChan)
	p.wg.Wait()
	log.Debugf("Peer %s is closed", p.Desc())
}

func (p *Peer) watchStop() {
	defer p.wg.Done()
	select {
	case <-p.stopChan:
		log.Debugf("Peer %s is closing its Conn", p.Desc())
		p.Conn.Close()
	}
	log.Debugf("Peer %s is leaving watchStop", p.Desc())
}

func (p *Peer) run() {
	defer p.wg.Done()
	for {
		nextState, err := p.State.perform()
		if err != nil {
			if err == errors.ClosedConnection {
				log.Debugf("Connection of peer %s was closed", p.Desc())
			} else {
				log.Errorf("Error performing '%s' state: %v", p.State.Name(), err)
				p.goneChan <- p
			}
			break
		}
		log.Debugf("Switching peer %s to state %s", p.Desc(), nextState.Name())
		p.State = nextState
	}
	log.Debugf("Peer %s is leaving run", p.Desc())
}

func (p *Peer) Desc() string {
	if p.State.ID() != StateIDHandshaking {
		u := p.User
		return fmt.Sprintf("%s-%s/%s", u.Nickname(), u.ShortID(), p.Conn.RemoteAddr())
	} else {
		return fmt.Sprintf("(not handshaked), %s", p.Conn.RemoteAddr())
	}
}

func (p *Peer) ID() (*ID, error) {
	if p.User != nil {
		return (*ID)(p.User.ID()), nil
	} else {
		return &ZeroID, errors.PeerIDUnknown
	}
}

func (p *Peer) ShortID() string {
	id, err := p.ID()
	if err == nil {
		eid := (*entity.ID)(id)
		return eid.Shorten()
	} else {
		return "[unknown]"
	}
}
