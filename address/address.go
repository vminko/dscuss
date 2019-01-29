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

package address

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	DomainPortRegex string = "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9]):\\d+$"
	IPPortRegex     string = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):\\d+$"
)

func IsValid(a string) bool {
	var domainPortRe = regexp.MustCompile(DomainPortRegex)
	var ipPortRe = regexp.MustCompile(IPPortRegex)
	return domainPortRe.MatchString(a) || ipPortRe.MatchString(a)
}

func Parse(a string) (string, int, error) {
	idx := strings.LastIndex(a, ":")
	host := a[:idx]
	portStr := a[idx+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}
