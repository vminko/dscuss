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
	"time"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/owner"
	"vminko.org/dscuss/p2p/peer"
)

// PeerPool is responsible for managing peers. It creates new peers, accounts
// peers and manages peer life cycle. But it has nothing to do with Entity
// transferring.
type PeerPool struct {
	cp          *ConnectionProvider
	owner       *owner.Owner
	stopWorkers chan bool
	stopPeers   chan struct{}
	peers       []*peer.Peer
	peersMx     sync.RWMutex
	wg          sync.WaitGroup
}

func NewPeerPool(cp *ConnectionProvider, owner *owner.Owner) *PeerPool {
	return &PeerPool{
		cp:          cp,
		owner:       owner,
		stopWorkers: make(chan bool, 1),
		stopPeers:   make(chan struct{}),
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

	// Stop workers
	pp.cp.Stop()
	pp.stopWorkers <- true
	pp.wg.Wait()
	log.Debugf("PeerPool stopped workers")

	// Stop peers
	var wg sync.WaitGroup
	pp.peersMx.RLock()
	for _, p := range pp.peers {
		log.Debugf("PeerPool is closing peer %s", p)
		wg.Add(1)
		myP := p
		go func() {
			myP.Close()
			wg.Done()
			log.Debugf("PeerPool closed peer %s", myP)
		}()
	}
	pp.peersMx.RUnlock()
	wg.Wait()
	log.Debug("PeerPool closed all peers")

	log.Debug("PeerPool is completely stopped")
}

func (pp *PeerPool) watchNewConnections() {
	defer pp.wg.Done()

	for conn := range pp.cp.newConnChan() {
		log.Debugf("New connection appeared, remote addr is %s", conn.RemoteAddr())
		peer := peer.New(
			conn,
			pp.owner,
			pp, // Validator
		)
		pp.peersMx.Lock()
		pp.peers = append(pp.peers, peer)
		pp.peersMx.Unlock()
	}

	/*
		for {
			select {
			case conn, ok := <-pp.cp.newConnChan():
				if !ok {
					log.Debugf("newConnChan is closed")
					return
				}
				log.Debugf("New connection appeared, remote addr is %s", conn.RemoteAddr())
				peer := peer.New(
					conn,
					pp.owner,
					pp, // Validator
				)
				pp.peersMx.Lock()
				pp.peers = append(pp.peers, peer)
				pp.peersMx.Unlock()
			case <-pp.stopWorkers:
				return
			case <-time.After(time.Second):
			}
		}
	*/
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
	log.Debugf("len(pp.peers) became %d", len(pp.peers))
	pp.peersMx.Unlock()
	if !isFound {
		log.Errorf("Failed to remove %s from the PeerPool", p)
	}
}

func (pp *PeerPool) watchGonePeers() {
	defer pp.wg.Done()
	latency := 1 * time.Second
	ticker := time.NewTicker(latency)
	for {
		select {
		case <-ticker.C:
			pp.peersMx.Lock()
			for _, peer := range pp.peers {
				log.Debugf("Checking if peer %s is gone", peer)
				if peer.IsGone() {
					pp.removePeer(peer)
					peer.Close()
					log.Debugf("Peer %s is removed from PP", peer)
				}
			}
			pp.peersMx.Unlock()
		case <-pp.stopWorkers:
			return
		}
	}
}

func (pp *PeerPool) ValidatePeer(newPeer *peer.Peer) bool {
	newPid := newPeer.ID()
	if newPid == nil {
		log.Fatalf("Handshaked peer %s has no ID", newPeer)
	}
	pp.peersMx.RLock()
	defer pp.peersMx.RUnlock()
	for _, p := range pp.peers {
		pid := p.ID()
		if pid != nil && *pid == *newPid && p != newPeer {
			p.AddAddresses(newPeer.Addresses())
			newPeer.ClearAddresses()
			return false
		}
	}
	return true
}

func (pp *PeerPool) ListPeers() []*peer.Info {
	pp.peersMx.RLock()
	defer pp.peersMx.RUnlock()
	res := make([]*peer.Info, len(pp.peers))
	for i, p := range pp.peers {
		res[i] = p.Info()
	}
	return res
}
