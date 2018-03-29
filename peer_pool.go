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

package dscuss

import (
	"sync"
)

// Responsible for accounting peers and managing peer lifecycle.
type peerPool struct {
	cfg             *config
	loginCtx        *loginContext
	cp              *connectionProvider
	closeChan       chan *peer
	stopPeersChan   chan struct{}
	addrReleaseChan chan string
	peers           []*peer
	peerWG          sync.WaitGroup
	selfWG          sync.WaitGroup
}

func newPeerPool(cfg *config, dir string, loginCtx *loginContext) *peerPool {
	addrReleaseChan := make(chan string)
	return &peerPool{
		cfg:             cfg,
		loginCtx:        loginCtx,
		cp:              newConnectionProvider(cfg, dir, loginCtx, addrReleaseChan),
		closeChan:       make(chan *peer),
		stopPeersChan:   make(chan struct{}),
		addrReleaseChan: addrReleaseChan,
	}
}

func (pp *peerPool) start() {
	Logf(DEBUG, "Starting peerPool")
	pp.selfWG.Add(2)
	go pp.listenNewConnections()
	go pp.listenClosedPeers()
	pp.cp.start()
}

func (pp *peerPool) stop() {
	Logf(DEBUG, "Stopping peerPool")
	pp.cp.stop()
	close(pp.stopPeersChan)
	pp.peerWG.Wait()
	close(pp.closeChan)
	pp.selfWG.Wait()
	Logf(DEBUG, "peerPool stopped")
}

func (pp *peerPool) listenNewConnections() {
	defer pp.selfWG.Done()
	for conn := range pp.cp.newConnChan() {
		Logf(DEBUG, "New connection appeared")
		pp.peerWG.Add(1)
		peer := newPeer(conn, pp.closeChan, pp.stopPeersChan, &pp.peerWG)
		pp.peers = append(pp.peers, peer)
	}
}

func (pp *peerPool) listenClosedPeers() {
	defer pp.selfWG.Done()
	for cpeer := range pp.closeChan {
		for _, addr := range cpeer.conn.associatedAddresses {
			pp.addrReleaseChan <- addr
		}
	}
}
