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
	"os"
)

type config struct {
	Port uint16
}

func NewConfig(path string) (*config, error) {
	// Default config values are defined here.
	var c = config{
		Port: 8004,
	}

	file, err := os.Open(path)
	if err != nil {
		Logf(ERROR, "Can't open config file %s: %v", path, err)
		return nil, ErrFilesystem
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		Logf(ERROR, "Can't decode json file %s: %v", err)
		return nil, ErrConfig
	}

	return &c, nil
}
