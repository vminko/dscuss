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
	"vminko.org/dscuss/subs"
)

type ID entity.ID

var ZeroID ID

// Peer is responsible for communication with other nodes.
// Implements the Dscuss protocol.
type Peer struct {
	conn          *connection.Connection
	owner         *owner.Owner
	storage       *storage.Storage
	validator     Validator
	goneChan      chan *Peer
	stopChan      chan struct{}
	outEntityChan chan entity.Entity
	wg            sync.WaitGroup
	State         State
	User          *entity.User
	Subs          subs.Subscriptions
}

// Info is a static Peer description for UI.
type Info struct {
	ShortID         string
	ID              string
	LocalAddr       string
	RemoteAddr      string
	AssociatedAddrs []string
	Nickname        string
	StateName       string
	Subscriptions   []string
}

type Validator interface {
	ValidatePeer(*Peer) bool
}

const (
	outEntityQueueCapacity int    = 100
	unknownValue           string = "[unknown]"
)

func (i *ID) String() string {
	s := unknownValue
	if i != nil {
		eid := (*entity.ID)(i)
		s = eid.String()
	}
	return s
}

func New(
	conn *connection.Connection,
	owner *owner.Owner,
	storage *storage.Storage,
	validator Validator,
	goneChan chan *Peer,
) *Peer {
	p := &Peer{
		conn:          conn,
		owner:         owner,
		storage:       storage,
		validator:     validator,
		goneChan:      goneChan,
		stopChan:      make(chan struct{}),
		outEntityChan: make(chan entity.Entity, outEntityQueueCapacity),
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
		log.Debugf("Peer %s is closing its conn", p.Desc())
		p.conn.Close()
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
		return fmt.Sprintf("%s-%s/%s-%s",
			u.Nickname, u.ShortID(), p.conn.LocalAddr(), p.conn.RemoteAddr())
	} else {
		return fmt.Sprintf("(not handshaked), %s", p.conn.RemoteAddr())
	}
}

func (p *Peer) ID() *ID {
	if p.User != nil {
		return (*ID)(p.User.ID())
	} else {
		return nil
	}
}

func (p *Peer) ShortID() string {
	shortID := unknownValue
	if p.ID() != nil {
		eid := (*entity.ID)(p.ID())
		shortID = eid.Shorten()
	}
	return shortID
}

func (p *Peer) isInterestedInMessage(m *entity.Message) bool {
	var t subs.Topic
	if m.IsReply() {
		r, err := p.storage.GetRoot(m)
		if err != nil {
			log.Fatalf("Got an error while fetching root of %s from DB: %v",
				m.Desc(), err)
		}
		t = r.Topic
	} else {
		t = m.Topic
	}
	return p.Subs.Covers(t)
}

func (p *Peer) isInterestedInEntity(ent entity.Entity) bool {
	switch e := ent.(type) {
	case *entity.Message:
		return p.isInterestedInMessage(e)
	case *entity.Operation:
		if e.OperationType() == entity.OperationTypeBanUser {
			return true
		}
		m, err := p.storage.GetMessage(&e.ObjectID)
		if err != nil {
			log.Fatalf("Got an error while fetching msg %s from DB: %v",
				e.ObjectID.Shorten(), err)
		}
		return p.isInterestedInMessage(m)
	case *entity.User:
		return false
	default:
		log.Fatal("BUG: unknown entity type")
	}
	return false
}

func (p *Peer) Info() *Info {
	nick := unknownValue
	if p.User != nil {
		nick = p.User.Nickname
	}
	subs := []string{unknownValue}
	if p.Subs != nil {
		subs = p.Subs.StringSlice()
	}
	return &Info{
		ShortID:         p.ShortID(),
		ID:              p.ID().String(),
		LocalAddr:       p.conn.LocalAddr(),
		RemoteAddr:      p.conn.RemoteAddr(),
		AssociatedAddrs: p.conn.Addresses(),
		Nickname:        nick,
		StateName:       p.State.Name(),
		Subscriptions:   subs,
	}
}

func (p *Peer) Addresses() []string {
	return p.conn.Addresses()
}

func (p *Peer) AddAddresses(new []string) {
	p.conn.AddAddresses(new)
}

func (p *Peer) ClearAddresses() {
	p.conn.ClearAddresses()
}
