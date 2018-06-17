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
	gonePeerChan    chan *peer.Peer
	stopPeersChan   chan struct{}
	addrReleaseChan chan string
	peers           []*peer.Peer
	peersMx         sync.RWMutex
	wg              sync.WaitGroup
}

func NewPeerPool(cp *ConnectionProvider, owner *owner.Owner) *PeerPool {
	addrReleaseChan := make(chan string)
	return &PeerPool{
		cp:              cp,
		owner:           owner,
		gonePeerChan:    make(chan *peer.Peer),
		stopPeersChan:   make(chan struct{}),
		addrReleaseChan: addrReleaseChan,
	}
}

func (pp *PeerPool) Start() {
	log.Debugf("Starting PeerPool")
	pp.wg.Add(2)
	go pp.watchNewConnections()
	go pp.watchGonePeers()
	pp.cp.Start()
}

func (pp *PeerPool) Stop() {
	log.Debugf("Stopping PeerPool")

	// Stop peers
	var wg sync.WaitGroup
	pp.peersMx.RLock()
	for _, p := range pp.peers {
		log.Debugf("PeerPool is closing peer %s", p.Desc())
		wg.Add(1)
		myP := p
		go func() {
			myP.Close()
			wg.Done()
			log.Debugf("PeerPool closed peer %s", myP.Desc())
		}()
	}
	pp.peersMx.RUnlock()
	wg.Wait()
	log.Debug("PeerPool closed all peers")
	close(pp.gonePeerChan)

	pp.cp.Stop()
	pp.wg.Wait()
	log.Debugf("PeerPool stopped")
}

func (pp *PeerPool) watchNewConnections() {
	defer pp.wg.Done()
	for conn := range pp.cp.newConnChan() {
		log.Debugf("New connection appeared, remote addr is %s", conn.RemoteAddr())
		peer := peer.New(conn, pp.owner, pp.validateHandshakedPeer, pp.gonePeerChan)
		pp.peersMx.Lock()
		pp.peers = append(pp.peers, peer)
		pp.peersMx.Unlock()
	}
}

func (pp *PeerPool) removePeer(p *peer.Peer) {
	isFound := false
	pp.peersMx.Lock()
	for i, ip := range pp.peers {
		if ip == p {
			pp.peers[len(pp.peers)-1], pp.peers[i] = pp.peers[i], pp.peers[len(pp.peers)-1]
			pp.peers = pp.peers[:len(pp.peers)-1]
			isFound = true
			break
		}
	}
	pp.peersMx.Unlock()
	if !isFound {
		log.Errorf("Failed to remove %s from the PeerPool", p.Desc())
	}
}

func (pp *PeerPool) watchGonePeers() {
	defer pp.wg.Done()
	for cpeer := range pp.gonePeerChan {
		for _, addr := range cpeer.Conn.Addresses() {
			log.Debugf("PP is releasing address %s", addr)
			//pp.addrReleaseChan <- addr
		}
		log.Debugf("Removing peer %s from PP", cpeer.Desc())
		pp.removePeer(cpeer)
		cpeer.Close()
		log.Debugf("Peer %s is removed from PP", cpeer.Desc())
	}
}

func (pp *PeerPool) validateHandshakedPeer(newPeer *peer.Peer) bool {
	newPid, err := newPeer.ID()
	if err != nil {
		log.Fatalf("Handshaked peer %s has no ID", newPeer.Desc())
	}
	pp.peersMx.RLock()
	defer pp.peersMx.RUnlock()
	for _, p := range pp.peers {
		pid, err := p.ID()
		if err == nil && pid == newPid && p != newPeer {
			log.Debugf("Found duplicated conn with %s: %s", newPeer.Desc(), p.Conn.Desc())
			p.Conn.AddAddresses(newPeer.Conn.Addresses())
			newPeer.Conn.ClearAddresses()
			return false
		}
	}
	return true
}
