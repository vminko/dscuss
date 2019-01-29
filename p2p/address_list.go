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
	"bufio"
	"os"
	"vminko.org/dscuss/address"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

// AddressList provider node addresses by reading them from a text file.
// Implements AddressProvider interface
type AddressList struct {
	filepath string
	ac       AddressConsumer
}

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
		if !address.IsValid(line) {
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
