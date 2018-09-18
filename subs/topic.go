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
	"regexp"
	"strings"
	"vminko.org/dscuss/errors"
	"vminko.org/dscuss/log"
)

type Topic []string

// TBD: precompile regular expressions for performance benefits
const (
	tagRegex string = "^[a-z0-9_]+$"
)

func NewTopic(str string) (Topic, error) {
	t := Topic(strings.Split(str, ","))
	if !t.IsValid() {
		log.Warningf("This is not a valid topic string: '%s'", str)
		return nil, errors.Parsing
	}
	return t, nil
}

func (t Topic) String() string {
	return strings.Join(t, ",")
}

func (t Topic) IsEqual(target Topic) bool {
	return t.String() == target.String()
}

func (t Topic) ContainsTopic(subtop Topic) bool {
	for _, tag := range t {
		if !subtop.ContainsTag(tag) {
			return false
		}
	}
	return true
}

func (t Topic) ContainsTag(target string) bool {
	for _, tag := range t {
		if tag == target {
			return true
		}
	}
	return false
}

func isTagValid(tag string) bool {
	var tagRe = regexp.MustCompile(tagRegex)
	return tagRe.MatchString(tag)
}

func (t Topic) IsValid() bool {
	if t == nil || len(t) == 0 {
		return false
	}
	for _, tag := range t {
		if !isTagValid(tag) {
			return false
		}
	}
	return true
}

func (t Topic) AddTag(tag string) (Topic, error) {
	if !isTagValid(tag) {
		log.Warningf("Attempt to add invalid tag: '%s'", tag)
		return nil, errors.Parsing
	}
	if t.ContainsTag(tag) {
		log.Warningf("Attempt to add duplicated tag: '%s'", tag)
		return t, nil
	}
	return append(t, tag), nil
}

func (t Topic) Copy() Topic {
	res := make([]string, len(t))
	copy(res, t)
	return res
}
