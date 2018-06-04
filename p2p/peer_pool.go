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

package p2p

import (
	"sync"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p/peer"
)

// PeerPool is responsible for managing peers. It creates new peers, accounts
// peers and manages peer life cycle. But it has nothing to do with Entity
// transferring.
type PeerPool struct {
	cp              *ConnectionProvider
	owner           *owner.Owner
	closeChan       chan *peer.Peer
	stopPeersChan   chan struct{}
	addrReleaseChan chan string
	peers           []*peer.Peer
	peerWG          sync.WaitGroup
	selfWG          sync.WaitGroup
}

func NewPeerPool(cp *ConnectionProvider, owner *owner.Owner) *PeerPool {
	addrReleaseChan := make(chan string)
	return &PeerPool{
		cp:              cp,
		owner:           owner,
		closeChan:       make(chan *peer.Peer),
		stopPeersChan:   make(chan struct{}),
		addrReleaseChan: addrReleaseChan,
	}
}

func (pp *PeerPool) Start() {
	log.Debugf("Starting PeerPool")
	pp.selfWG.Add(2)
	go pp.listenNewConnections()
	go pp.listenClosedPeers()
	pp.cp.Start()
}

func (pp *PeerPool) Stop() {
	log.Debugf("Stopping PeerPool")
	pp.cp.Stop()
	close(pp.stopPeersChan)
	pp.peerWG.Wait()
	close(pp.closeChan)
	pp.selfWG.Wait()
	log.Debugf("PeerPool stopped")
}

func (pp *PeerPool) listenNewConnections() {
	defer pp.selfWG.Done()
	for conn := range pp.cp.newConnChan() {
		log.Debugf("New connection appeared")
		pp.peerWG.Add(1)
		peer := peer.New(conn, pp.owner, pp.stopPeersChan, &pp.peerWG, pp.closeChan)
		pp.peers = append(pp.peers, peer)
	}
}

func (pp *PeerPool) listenClosedPeers() {
	defer pp.selfWG.Done()
	for cpeer := range pp.closeChan {
		for _, addr := range cpeer.Conn.AssociatedAddresses() {
			pp.addrReleaseChan <- addr
		}
	}
}
