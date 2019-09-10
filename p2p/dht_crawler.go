/*
This file is part of Dscuss.
Copyright (C) 2019  Vitaly Minko

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
	"crypto/sha1"
	"fmt"
	"github.com/nictuku/dht"
	"net"
	"sync"
	"time"
	"vminko.org/dscuss/address"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/subs"
)

// DHTCrawler discovers new peers via DHT.
// Implements AddressProvider interface
type DHTCrawler struct {
	dht       *dht.DHT
	addr      string
	port      int
	bootstrap string
	advPort   int
	subs      subs.Subscriptions
	ac        AddressConsumer
	stopChan  chan struct{}
	wg        sync.WaitGroup
	processed map[string]struct{}
}

const (
	DHTCrawlerTimeout time.Duration = 5 * time.Second
)

func NewDHTCrawler(
	addr string,
	port int,
	bootstrap string,
	advPort int,
	s subs.Subscriptions,
) *DHTCrawler {
	return &DHTCrawler{
		addr:      addr,
		port:      port,
		bootstrap: bootstrap,
		advPort:   advPort,
		subs:      s.ToCombinations(),
		stopChan:  make(chan struct{}),
		processed: make(map[string]struct{}),
	}
}

func (dc *DHTCrawler) RegisterAddressConsumer(ac AddressConsumer) {
	dc.ac = ac
}

func (dc *DHTCrawler) Start() {
	log.Debugf("Starting DHTCrawler on: %s:%d", dc.addr, dc.port)
	if dc.ac == nil {
		log.Fatal("Attempt to start providing addresses when AddressConsumer is not set")
	}
	cfg := dht.NewConfig()
	cfg.Address = dc.addr
	cfg.Port = dc.port
	cfg.RateLimit = -1
	cfg.ClientPerMinuteLimit = 10000
	cfg.DHTRouters = dc.bootstrap
	cfg.SaveRoutingTable = false
	log.Debugf("Using these DHTRouters: %s", cfg.DHTRouters)

	var err error
	dc.dht, err = dht.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create a DHT node: %v", err)
	}
	if log.IsDebugEnabled() {
		dl := &dhtLogger{}
		dc.dht.DebugLogger = dl
	}
	err = dc.dht.Start()
	if err != nil {
		log.Fatalf("Failed to start DHT node: %v", err)
	}
	log.Debugf("DHTCrawler started on port: %d", dc.dht.Port())
	dc.wg.Add(2)
	go dc.requestAddresses()
	go dc.drainAddresses()
}

func (dc *DHTCrawler) Stop() {
	log.Debug("Stopping DHTCrawler")
	close(dc.stopChan)
	dc.dht.Stop()
	log.Debug("DHT stopped")
	dc.wg.Wait()
	log.Debug("DHTCrawler stopped")
}

func (dc *DHTCrawler) requestAddresses() {
	defer dc.wg.Done()
	log.Debugf("Starting requestAddresses loop for subs %s", dc.subs)
	for {
		select {
		case <-dc.stopChan:
			log.Debug("Leaving requestAddresses")
			return
		case <-time.Tick(DHTCrawlerTimeout):
			for _, t := range dc.subs {
				ih := calcInfoHash(t.String())
				log.Debugf("Requesting addresses for topic %s", t)
				dc.dht.PeersRequestPort(string(ih), true, dc.advPort)
			}
		}
	}
}

func (dc *DHTCrawler) drainAddresses() {
	defer dc.wg.Done()
	for {
		select {
		case <-dc.stopChan:
			log.Debug("Leaving drainAddresses")
			return
		case r := <-dc.dht.PeersRequestResults:
			log.Debug("Draining addresses...")
			for _, peers := range r {
				for _, x := range peers {
					a := dht.DecodePeerAddress(x)
					dc.handleNewAddress(a)
				}
			}
		}
	}
}

func (dc *DHTCrawler) handleNewAddress(a string) {
	log.Debugf("Found new address: %s", a)
	if !address.IsValid(a) {
		log.Warningf("DHT is poisoned, found malformed address %s", a)
		return
	}
	if _, ok := dc.processed[a]; ok {
		log.Debugf("Address '%s' has already been processed, skipping it", a)
		return
	}
	if !isAddressLocal(a, dc.advPort) {
		dc.ac.AddressFound(a)
	} else {
		log.Debugf("Skipping local address %s", a)
	}
	dc.processed[a] = struct{}{}
}

func localIPs() []string {
	var res []string
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Errorf("Failed to get network interfaces: %v", err)
		return res
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Errorf("Failed to get addresses of the interface %s: %v", i, err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			res = append(res, ip.String())
		}
	}
	return res
}

func isAddressLocal(a string, localPort int) bool {
	ip, port, err := address.Parse(a)
	if err != nil {
		log.Errorf("BUG: malformed address '%s' passed input validation: %v", a, err)
		return false
	}
	if localPort != port {
		return false
	}
	for _, lip := range localIPs() {
		if lip == ip {
			return true
		}
	}
	return false
}

func calcInfoHash(s string) dht.InfoHash {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	ih, err := dht.DecodeInfoHash(fmt.Sprintf("%x", bs))
	if err != nil {
		log.Fatalf("Failed to decode InfoHash for %s error: %v\n", s, err)
	}
	return ih
}

type dhtLogger struct{}

func (dl *dhtLogger) Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func (dl *dhtLogger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func (dl *dhtLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
