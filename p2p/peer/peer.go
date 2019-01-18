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
	"time"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p/connection"
	"vminko.org/dscuss/subs"
)

type ID entity.ID

var ZeroID ID

// Peer is responsible for communication with other nodes.
// Implements the Dscuss protocol.
type Peer struct {
	conn          *connection.Connection
	owner         *owner.Owner
	validator     Validator
	goneChan      chan *Peer
	stopChan      chan struct{}
	outEntityChan chan entity.Entity
	wg            sync.WaitGroup
	State         State
	User          *entity.User
	Subs          subs.Subscriptions
	hist          *entity.UserHistory
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
	validator Validator,
	goneChan chan *Peer,
) *Peer {
	p := &Peer{
		conn:          conn,
		owner:         owner,
		validator:     validator,
		goneChan:      goneChan,
		stopChan:      make(chan struct{}),
		outEntityChan: make(chan entity.Entity, outEntityQueueCapacity),
	}
	p.State = newStateHandshaking(p)
	p.owner.Storage.AttachObserver(p.outEntityChan)
	p.wg.Add(2)
	go p.run()
	go p.watchStop()
	return p
}

func (p *Peer) Close() {
	log.Debugf("Close requested for peer %s", p)
	isSynced := p.State.ID() != StateIDHandshaking && p.State.ID() != StateIDActiveSyncing &&
		p.State.ID() != StateIDPassiveSyncing
	if isSynced {
		log.Debugf("Saving history for peer %s", p)
		h := &entity.UserHistory{p.User.ID(), time.Now(), p.Subs}
		p.owner.Profile.PutUserHistory(h)
	}
	p.owner.Storage.DetachObserver(p.outEntityChan)
	close(p.stopChan)
	p.wg.Wait()
	log.Debugf("Peer %s is closed", p)
}

func (p *Peer) watchStop() {
	defer p.wg.Done()
	select {
	case <-p.stopChan:
		log.Debugf("Peer %s is closing its conn", p)
		p.conn.Close()
	}
	log.Debugf("Peer %s is leaving watchStop", p)
}

func (p *Peer) run() {
	defer p.wg.Done()
	for {
		nextState, err := p.State.perform()
		if err != nil {
			if err == errors.ClosedConnection {
				// Peer was deliberately stopped by PeerPool
				log.Debugf("Connection of peer %s was closed", p)
			} else {
				log.Errorf("Error performing '%s' state: %v", p.State.Name(), err)
				p.goneChan <- p
			}
			break
		}
		log.Debugf("Switching peer %s to state %s", p, nextState.Name())
		p.State = nextState
	}
	log.Debugf("Peer %s is leaving run", p)
}

func (p *Peer) String() string {
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

func (p *Peer) isInterestedInMessage(s subs.Subscriptions, m *entity.Message) bool {
	var t subs.Topic
	if m.IsReply() {
		r, err := p.owner.Storage.GetRoot(m)
		if err != nil {
			log.Fatalf("Got an error while fetching root of %s from DB: %v",
				m, err)
		}
		t = r.Topic
	} else {
		t = m.Topic
	}
	return s.Covers(t)
}

func (p *Peer) isInterestedInEntity(ent entity.Entity, stored time.Time) bool {
	subs := p.Subs
	if p.hist != nil && stored.Before(p.hist.Disconnected) {
		subs = p.Subs.Diff(p.hist.Subs)
	}
	if subs == nil {
		return false
	}
	switch e := ent.(type) {
	case *entity.Message:
		return p.isInterestedInMessage(subs, e)
	case *entity.Operation:
		if e.OperationType() == entity.OperationTypeBanUser {
			return true
		}
		m, err := p.owner.Storage.GetMessage(&e.ObjectID)
		if err != nil {
			log.Fatalf("Got an error while fetching msg %s from DB: %v",
				e.ObjectID.Shorten(), err)
		}
		return p.isInterestedInMessage(subs, m)
	case *entity.User:
		return false
	default:
		log.Fatalf("BUG: unknown entity type: %T", ent)
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
