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
	"bufio"
	"os"
	"regexp"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

type AddressConsumer interface {
	AddressFound(a string)
	ErrorFindingAddresses(err error)
}

// AddressProvider isolates ConnectionProvider from implementations of various
// address discovery methods.
type AddressProvider interface {
	RegisterAddressConsumer(ac AddressConsumer)
	Start()
	Stop()
}

// AddressList provider node addresses by reading them from a text file.
// Implements AddressProvider interface
type AddressList struct {
	filepath string
	ac       AddressConsumer
}

const (
	HostPortRegex string = "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9]):\\d+$"
	IPPortRegex   string = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):\\d+$"
)

func NewAddressList(filepath string) *AddressList {
	return &AddressList{filepath: filepath}
}

func (al *AddressList) RegisterAddressConsumer(ac AddressConsumer) {
	al.ac = ac
}

func (al *AddressList) Start() {
	if al.ac == nil {
		log.Fatal("attempt to start providing addresses when AddressConsumer is not set")
	}
	go al.readAddresses()
}

func (al *AddressList) Stop() {
}

func (al *AddressList) readAddresses() {
	file, err := os.Open(al.filepath)
	if err != nil {
		log.Errorf("Can't open file %s: %v", al.filepath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var hostPortRe = regexp.MustCompile(HostPortRegex)
		var ipPortRe = regexp.MustCompile(IPPortRegex)
		if !hostPortRe.MatchString(line) && !ipPortRe.MatchString(line) {
			log.Warningf("'%s' is not a valid peer address, ignoring it.", line)
			log.Warning("Valid peer address is either host:port or ip:port.")
			al.ac.ErrorFindingAddresses(errors.Parsing)
			continue
		}

		log.Debugf("Found peer address %s", line)
		al.ac.AddressFound(line)
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Error scanning file %s: %v", al.filepath, err)
		al.ac.ErrorFindingAddresses(errors.Filesystem)
	}
}
