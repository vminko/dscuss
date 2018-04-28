/*
This file is part of Dscuss.
Copyright (C) 2017  Vitaly Minko

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

// Config provides access to the parameters from the configuration file and
// saves them to the file.

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"vminko.org/dscuss/log"
)

type NetworkConfig struct {
	HostPort        string
	AddressProvider string
	MaxInConnCount  uint32
	MaxOutConnCount uint32
}

type config struct {
	Network NetworkConfig
}

var defaultConfig = config{
	Network: NetworkConfig{
		HostPort:        ":8004",
		AddressProvider: "addrlist",
		MaxInConnCount:  10,
		MaxOutConnCount: 10,
	},
}

func (c *config) save(path string) error {
	cfgStr, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		log.Errorf("Can't marshal config: %v", err)
		return err
	}
	err = ioutil.WriteFile(path, []byte(cfgStr), 0644)
	if err != nil {
		log.Errorf("Can't save config file to %s: %v", path, err)
		return err
	}
	return nil
}

func newConfig(path string) (*config, error) {
	var c = defaultConfig

	file, err := os.Open(path)
	if err != nil {
		log.Warningf("Can't open config file %s: %v", path, err)
		err = c.save(path)
		if err != nil {
			log.Fatalf("Can't save default config file to %s: %v", path, err)
		}
		return &c, nil
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		log.Errorf("Error decoding json file %s: %v", err)
		return nil, ErrConfig
	}

	/* TBD: validate parameters */

	return &c, nil
}
