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

package controller

import (
	"net/http"
	"vminko.org/dscuss"
)

var LoginHandler, ProfileHandler, BoardHandler, ThreadHandler, CreateThread, ReplyThreadHandler,
	AddModeratorHandler, DelModeratorHandler, SubscribeHandler, UnsubscribeHandler, UserHandler,
	RemoveMessageHandler, BanUserHandler, ListOperationsHandler, ListPeersHandler,
	PeerHistoryHandler func(w http.ResponseWriter, r *http.Request)

func InitHandlers(l *dscuss.LoginHandle) {
	LoginHandler = makeHandler(handleLogin, l)
	ProfileHandler = makeHandler(handleProfile, l)
	BoardHandler = makeHandler(handleBoard, l)
	ThreadHandler = makeHandler(handleThread, l)
	CreateThread = makeHandler(handleCreateThread, l)
	ReplyThreadHandler = makeHandler(handleReplyThread, l)
	AddModeratorHandler = makeHandler(handleAddModerator, l)
	DelModeratorHandler = makeHandler(handleDelModerator, l)
	SubscribeHandler = makeHandler(handleSubscribe, l)
	UnsubscribeHandler = makeHandler(handleUnsubscribe, l)
	UserHandler = makeHandler(handleUser, l)
	RemoveMessageHandler = makeHandler(handleRemoveMessage, l)
	BanUserHandler = makeHandler(handleBanUser, l)
	ListOperationsHandler = makeHandler(handleListOperations, l)
	ListPeersHandler = makeHandler(handleListPeers, l)
	PeerHistoryHandler = makeHandler(handlePeerHistory, l)
}
