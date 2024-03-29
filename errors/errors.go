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

package errors

import (
	"errors"
)

var (
	Internal            = errors.New("internal error")
	Filesystem          = errors.New("filesystem error")
	Parsing             = errors.New("error parsing input data")
	Config              = errors.New("error handling config file")
	Database            = errors.New("database error")
	CantOpenDB          = errors.New("unable to open the database")
	DBOperFailed        = errors.New("error operating on the database")
	InconsistentDB      = errors.New("the database contains inconsistent data")
	NoSuchEntity        = errors.New("can't find requested entity")
	WrongNickname       = errors.New("unacceptable nickname")
	AlreadyLoggedIn     = errors.New("another user is already logged in")
	NotLoggedIn         = errors.New("you have to login first")
	AlreadyRegistered   = errors.New("such user is already registered")
	NoSuchUser          = errors.New("can't find the specified user")
	NoSuchModerator     = errors.New("can't find the specified moderator")
	AlreadyModerator    = errors.New("the specified user is already a moderator")
	DuplicationAttempt  = errors.New("attempt to duplicate data in the database")
	NoUserHistory       = errors.New("can't find history record for the specified user")
	WrongPacketType     = errors.New("wrong packet type")
	MalformedPayload    = errors.New("payload of the packet is malformed")
	ProtocolViolation   = errors.New("protocol violation detected")
	UnsupportedProtocol = errors.New("requested version of the protocol is not supported")
	ClosedConnection    = errors.New("use of closed connection")
	InvalidPeer         = errors.New("peer validation failed")
	WrongArguments      = errors.New("wrong arguments")
	AlreadySubscribed   = errors.New("you are already subscribed to the specified topic")
	NotSubscribed       = errors.New("you are not subscribed to the specified topic")
	ForbiddenOperation  = errors.New("forbidden operation")
	UserBanned          = errors.New("user is banned")
	NoSuchTag           = errors.New("can't find requested tag")
	PacketSizeExceeded  = errors.New("the packet size exceeded the limit")
	MsgDepthExceeded    = errors.New("the thread depth exceeded the limit")
	MsgPostRateErr      = errors.New("attempt to violate the limit of the message post rate")
	OperPostRateErr     = errors.New("attempt to violate the limit of the operation post rate")
	SubsSizeExceeded    = errors.New("too many topics in the subscriptions")
)

// TBD: consider https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
