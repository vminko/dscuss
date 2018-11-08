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

package subs

import (
	"bufio"
	"io"
	"strings"
	"vminko.org/dscuss/log"
)

type Subscriptions []Topic

func (s Subscriptions) AddTopic(t Topic) Subscriptions {
	if s.Contains(t) {
		log.Warningf("Attempt to add duplicated topic: '%s'", t)
		return s
	}
	return append(s, t)
}

func (s Subscriptions) Copy() Subscriptions {
	var res Subscriptions
	for _, t := range s {
		res = res.AddTopic(t.Copy())
	}
	return res
}

func (s Subscriptions) Contains(target Topic) bool {
	for _, t := range s {
		if t.IsEqual(target) {
			return true
		}
	}
	return false
}

func (s Subscriptions) Covers(target Topic) bool {
	for _, t := range s {
		if t.ContainsTopic(target) {
			return true
		}
	}
	return false
}

func (s Subscriptions) IsValid() bool {
	if s == nil || len(s) == 0 {
		return false
	}
	for _, t := range s {
		if !t.IsValid() {
			return false
		}
	}
	return true
}

func (s Subscriptions) String() string {
	var str string
	for _, t := range s {
		str += t.String() + "\n"
	}
	return str
}

func (s Subscriptions) StringSlice() []string {
	var slice []string
	for _, t := range s {
		slice = append(slice, t.String())
	}
	return slice
}

func ReadString(s string) (Subscriptions, error) {
	log.Debugf("Reading subscriptions from string '%s'.", s)
	return Read(strings.NewReader(s))
}

func Read(r io.Reader) (Subscriptions, error) {
	var subs []Topic
	scanner := bufio.NewScanner(r)
	num := 0
	for scanner.Scan() {
		num++
		line := scanner.Text()
		log.Debugf("Found topic '%s'", line)
		topic, err := NewTopic(line)
		if err != nil {
			log.Warningf("Malformed line #%d in the subscriptions input: '%s'."+
				" Ignoring it.", num, line)
			continue
		}
		if (Subscriptions)(subs).Contains(topic) {
			log.Warningf("Duplicated topic in the subscriptions input: '%s'!",
				line)
		} else {

			subs = append(subs, topic)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Errorf("Error scanning subscriptions input: %v", err)
		return nil, err
	}
	return Subscriptions(subs), nil
}
