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
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ConnectionProviderLatency time.Duration = 1 // in seconds
	AddressListFileName       string        = "addresses"
	HostPortRegex             string        = "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9]):\\d+$"
	IPPortRegex               string        = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):\\d+$"
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

func (a *addressMap) Store(key string, value bool) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.m[key] = value
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
type connectionProvider struct {
	cfg          *config
	loginCtx     *loginContext
	hostport     string
	listener     *net.TCPListener
	wg           sync.WaitGroup
	stopChan     chan struct{}
	outChan      chan *connection
	releaseChan  chan string
	outAddrs     *addressMap
	inConnCount  uint32
	outConnCount uint32
}

func newConnectionProvider(
	cfg *config,
	dir string,
	loginCtx *loginContext,
	releaseChan chan string,
) *connectionProvider {
	cp := &connectionProvider{
		cfg:          cfg,
		loginCtx:     loginCtx,
		hostport:     cfg.Network.Address,
		outChan:      make(chan *connection),
		stopChan:     make(chan struct{}),
		outAddrs:     &addressMap{m: make(map[string]bool)},
		releaseChan:  releaseChan,
		inConnCount:  0,
		outConnCount: 0,
	}
	setDefaultBootstrapAddresses(cp.outAddrs)
	for _, bootstp := range cfg.Network.Bootstrappers {
		switch bootstp {
		case "addrlist":
			addrFilePath := filepath.Join(
				dir,
				AddressListFileName,
			)
			err := readAddresses(addrFilePath, cp.outAddrs)
			if err != nil {
				Logf(WARNING, "Can't read node addresses from file %s", addrFilePath)
			}
		/* TBD:
		case "dht":
			icase "dns":
		*/
		default:
			panic("Unknown bootstrapper type " + bootstp)
		}
	}
	return cp
}

func setDefaultBootstrapAddresses(outAddrs *addressMap) {
	outAddrs.Store(DefaultBootstrapAddress, false)
}

func readAddresses(path string, outAddrs *addressMap) error {
	file, err := os.Open(path)
	if err != nil {
		Logf(ERROR, "Can't open file %s: %v", path, err)
		return ErrFilesystem
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var hostPortRe = regexp.MustCompile(HostPortRegex)
		var ipPortRe = regexp.MustCompile(IPPortRegex)
		if !hostPortRe.MatchString(line) && !ipPortRe.MatchString(line) {
			Logf(WARNING, "'%s' is not a valid peer address, ignoring it.", line)
			Log(WARNING, "Valid peer address is either host:port or ip:port.")
			continue
		}

		if _, ok := outAddrs.Load(line); ok {
			Logf(WARNING, "Duplicated peer address: '%s'!", line)
			continue
		}
		Logf(DEBUG, "Found peer address %s", line)
		outAddrs.Store(line, false)
	}

	if err := scanner.Err(); err != nil {
		Logf(ERROR, "Error scanning file %s: %v", path, err)
		return ErrFilesystem
	}
	return nil
}

func (cp *connectionProvider) start() {
	Logf(DEBUG, "Starting connectionProvider")
	cp.wg.Add(3)
	go cp.listenIncomingConnections()
	go cp.establishOutgoingConnections()
	go cp.handleClosedConnections()
}

func (cp *connectionProvider) stop() {
	Logf(DEBUG, "Stopping connectionProvider")
	close(cp.stopChan)
	if cp.listener != nil {
		cp.listener.Close()
	}
	cp.wg.Wait()
	close(cp.outChan)
	Logf(DEBUG, "connectionProvider stopped")
}

func (cp *connectionProvider) newConnChan() chan *connection {
	return cp.outChan
}

func (cp *connectionProvider) listenIncomingConnections() {
	var maxInConnCount = cp.cfg.Network.MaxInConnCount
	defer cp.wg.Done()
	tcpAddr, err := net.ResolveTCPAddr("tcp4", cp.hostport)
	if err != nil {
		Logf(FATAL, "Can't resolve %s: %v", cp.hostport, err)
	}
	cp.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		Logf(FATAL, "Can't start listening on %s: %v", cp.hostport, err)
	}
	for {
		Logf(DEBUG, "Listening incoming connections on %s.", cp.hostport)
		select {
		case <-cp.stopChan:
			Log(DEBUG, "Stop requested")
			return
		default:
		}
		if atomic.LoadUint32(&cp.inConnCount) >= maxInConnCount {
			Log(DEBUG, "Reached maxInConnCount, skipping Accept()")
			time.Sleep(time.Second * time.Duration(ConnectionProviderLatency))
			continue
		}
		Log(DEBUG, "Trying to accept incoming connection...")
		conn, err := cp.listener.Accept()
		if err != nil {
			Logf(WARNING, "Error accepting connection: %v", err)
			continue
		}
		Logf(INFO, "Established new connection with %s", conn.RemoteAddr().String())
		atomic.AddUint32(&cp.inConnCount, 1)
		dconn := newConnection(conn, false)
		cp.outChan <- dconn
	}
}

func (cp *connectionProvider) tryToConnect(addr string, isUsed bool) bool {
	var maxOutConnCount = cp.cfg.Network.MaxOutConnCount
	if isUsed {
		Logf(DEBUG, "%s is already used, skipping it", addr)
		return true
	}
	Logf(DEBUG, "Trying to connect to %s", addr)
	d := net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ConnectionProviderLatency)
	defer cancel()
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		Logf(INFO, "Can't establish TCP connection with %s: %v", addr, err)
		return true
	}
	Logf(INFO, "Established new connection with %s", conn.RemoteAddr().String())
	atomic.AddUint32(&cp.outConnCount, 1)
	cp.outAddrs.Store(addr, true)
	dconn := newConnection(conn, true)
	cp.outChan <- dconn
	if atomic.LoadUint32(&cp.outConnCount) == maxOutConnCount {
		Log(DEBUG, "Reached maxOutConnCount, breaking dialing loop")
		return false
	}
	return true
}

func (cp *connectionProvider) establishOutgoingConnections() {
	var maxOutConnCount = cp.cfg.Network.MaxOutConnCount
	defer cp.wg.Done()
	for {
		select {
		case <-cp.stopChan:
			Log(DEBUG, "Stop requested")
			return
		default:
		}
		if atomic.LoadUint32(&cp.outConnCount) >= maxOutConnCount {
			Log(DEBUG, "Reached maxOutConnCount, skipping dialing loop")
		} else {
			cp.outAddrs.Range(func(addr string, isUsed bool) bool {
				select {
				case <-cp.stopChan:
					Log(DEBUG, "Stop requested")
					return false
				default:
					return cp.tryToConnect(addr, isUsed)
				}
			})
		}
		time.Sleep(time.Second * ConnectionProviderLatency)
	}
}

func (cp *connectionProvider) handleClosedConnections() {
	defer cp.wg.Done()
	for {
		Log(DEBUG, "Handling closed connections...")
		select {
		case <-cp.stopChan:
			Log(DEBUG, "Stop requested")
			return
		case addr := <-cp.releaseChan:
			Log(DEBUG, "Releasing address "+addr)
			isUsed, ok := cp.outAddrs.Load(addr)
			if ok {
				if !isUsed {
					Logf(ERROR, "Attempt to release unused address %s", addr)
				}
				cp.outAddrs.Store(addr, false)
				// decrement outConnCount
				atomic.AddUint32(&cp.outConnCount, ^uint32(0))
			} else {
				// decrement inConnCount
				atomic.AddUint32(&cp.inConnCount, ^uint32(0))
			}
		default:
			Log(DEBUG, "Nothing to close...")
			time.Sleep(time.Second * ConnectionProviderLatency)
			continue
		}
	}
}
