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
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/p2p/connection"
)

const (
	ConnectionProviderLatency time.Duration = 1 // in seconds
	DefaultBootstrapAddress   string        = "vminko.org:8004"
)

type addressMap struct {
	mx sync.RWMutex
	m  map[string]bool
}

func (a *addressMap) Load(key string) (bool, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()
	val, ok := a.m[key]
	return val, ok
}

func (a *addressMap) Add(key string, value bool) {
	a.mx.Lock()
	defer a.mx.Unlock()
	_, ok := a.m[key]
	if !ok {
		a.m[key] = value
	}
}

func (a *addressMap) Change(key string, value bool) {
	a.mx.Lock()
	defer a.mx.Unlock()
	_, ok := a.m[key]
	if ok {
		a.m[key] = value
	} else {
		log.Errorf("Attempt to change state of unknown address %s", key)
	}
}

func (a *addressMap) Range(f func(key string, value bool) bool) {
	a.mx.RLock()
	tmp := make(map[string]bool)
	for k, v := range a.m {
		tmp[k] = v
	}
	a.mx.RUnlock()
	for k, v := range tmp {
		if !f(k, v) {
			break
		}
	}
}

// Responsible for establishing connections with other peers.
type ConnectionProvider struct {
	hostport        string
	listener        *net.TCPListener
	wg              sync.WaitGroup
	stopChan        chan struct{}
	outChan         chan *connection.Connection
	ap              AddressProvider
	outAddrs        *addressMap
	maxInConnCount  uint32
	maxOutConnCount uint32
	inConnCount     uint32
	outConnCount    uint32
}

func NewConnectionProvider(
	ap AddressProvider,
	hostport string,
	maxInConnCount uint32,
	maxOutConnCount uint32,
) *ConnectionProvider {
	cp := &ConnectionProvider{
		ap:              ap,
		maxInConnCount:  maxInConnCount,
		maxOutConnCount: maxOutConnCount,
		hostport:        hostport,
		outChan:         make(chan *connection.Connection),
		stopChan:        make(chan struct{}),
		outAddrs:        &addressMap{m: make(map[string]bool)},
		inConnCount:     0,
		outConnCount:    0,
	}
	setDefaultBootstrapAddresses(cp.outAddrs)
	return cp
}

func setDefaultBootstrapAddresses(outAddrs *addressMap) {
	outAddrs.Add(DefaultBootstrapAddress, false)
}

func (cp *ConnectionProvider) Start() {
	log.Debugf("Starting ConnectionProvider")
	cp.ap.RegisterAddressConsumer(cp)
	cp.wg.Add(2)
	go cp.listenIncomingConnections()
	go cp.establishOutgoingConnections()
	cp.ap.Start()
}

func (cp *ConnectionProvider) Stop() {
	log.Debugf("Stopping ConnectionProvider")
	cp.ap.Stop()
	close(cp.stopChan)
	if cp.listener != nil {
		cp.listener.Close()
	}
	cp.wg.Wait()
	close(cp.outChan)
	log.Debugf("ConnectionProvider stopped")
}

func (cp *ConnectionProvider) newConnChan() chan *connection.Connection {
	return cp.outChan
}

func (cp *ConnectionProvider) AddressFound(a string) {
	cp.outAddrs.Add(a, false)
}

func (cp *ConnectionProvider) ErrorFindingAddresses(err error) {
	log.Fatalf("AddressProvider failure: %v", err)
}

func (cp *ConnectionProvider) listenIncomingConnections() {
	defer cp.wg.Done()
	tcpAddr, err := net.ResolveTCPAddr("tcp4", cp.hostport)
	if err != nil {
		log.Fatalf("Can't resolve %s: %v", cp.hostport, err)
	}
	cp.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("Can't start listening on %s: %v", cp.hostport, err)
	}
	for {
		log.Debugf("Listening incoming connections on %s.", cp.hostport)
		select {
		case <-cp.stopChan:
			log.Debug("Stop requested")
			return
		default:
		}
		if atomic.LoadUint32(&cp.inConnCount) >= cp.maxInConnCount {
			log.Debug("Reached maxInConnCount, skipping Accept()")
			time.Sleep(time.Second * time.Duration(ConnectionProviderLatency))
			continue
		}
		log.Debug("Trying to accept incoming connection...")
		conn, err := cp.listener.Accept()
		if err != nil {
			log.Warningf("Error accepting connection: %v", err)
			continue
		}
		log.Infof("Established new connection with %s", conn.RemoteAddr().String())
		atomic.AddUint32(&cp.inConnCount, 1)
		dconn := connection.New(conn, true)
		dconn.RegisterCloseHandler(cp.createCloseConnHandler())
		cp.outChan <- dconn
	}
}

func (cp *ConnectionProvider) tryToConnect(addr string, isUsed bool) bool {
	if isUsed {
		log.Debugf("%s is already used, skipping it", addr)
		return true
	}
	log.Debugf("Trying to connect to %s", addr)
	d := net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ConnectionProviderLatency)
	defer cancel()
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		log.Infof("Can't establish TCP connection with %s: %v", addr, err)
		return true
	}
	log.Infof("Established new connection with %s", conn.RemoteAddr().String())
	atomic.AddUint32(&cp.outConnCount, 1)
	cp.outAddrs.Change(addr, true)
	dconn := connection.New(conn, false)
	dconn.RegisterCloseHandler(cp.createCloseConnHandler())
	cp.outChan <- dconn
	if atomic.LoadUint32(&cp.outConnCount) == cp.maxOutConnCount {
		log.Debug("Reached maxOutConnCount, breaking dialing loop")
		return false
	}
	return true
}

func (cp *ConnectionProvider) establishOutgoingConnections() {
	defer cp.wg.Done()
	for {
		select {
		case <-cp.stopChan:
			log.Debug("Stop requested")
			return
		default:
		}
		if atomic.LoadUint32(&cp.outConnCount) >= cp.maxOutConnCount {
			log.Debug("Reached maxOutConnCount, skipping dialing loop")
		} else {
			cp.outAddrs.Range(func(addr string, isUsed bool) bool {
				select {
				case <-cp.stopChan:
					log.Debug("Stop requested")
					return false
				default:
					return cp.tryToConnect(addr, isUsed)
				}
			})
		}
		time.Sleep(time.Second * ConnectionProviderLatency)
	}
}

func (cp *ConnectionProvider) createCloseConnHandler() func(*connection.Connection) {
	return func(conn *connection.Connection) {
		log.Debugf("Executing close handler for connection %s", conn.Desc())
		if conn.IsIncoming() {
			// decrement inConnCount
			atomic.AddUint32(&cp.inConnCount, ^uint32(0))
		} else {
			// decrement outConnCount
			atomic.AddUint32(&cp.outConnCount, ^uint32(0))
		}
		for _, addr := range conn.Addresses() {
			log.Debug("CP is releasing address " + addr)
			isUsed, ok := cp.outAddrs.Load(addr)
			if ok {
				if !isUsed {
					log.Errorf("Attempt to release unused address %s", addr)
				}
				cp.outAddrs.Change(addr, false)
			}
		}
		log.Debugf("Leaving closeConnHandler, inConnCount = %d, outConnCount=%d",
			atomic.LoadUint32(&cp.inConnCount), atomic.LoadUint32(&cp.outConnCount))
	}
}
