/*
This file is part of Dscuss.
Copyright (C) 2018-2019  Vitaly Minko

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

package peer

import (
	"time"
	"vminko.org/dscuss/entity"
	"vminko.org/dscuss/log"
	"vminko.org/dscuss/packet"
)

// StateActiveSyncing implements the active part of the sync protocol.
// In case of success, switches peer to either Idle (for peers connected via
// passive connections) or StatePassiveSyncing (for peers connected via active
// connections).
type StateActiveSyncing struct {
	p                *Peer
	messagesToSync   []*entity.StoredMessage
	operationsToSync []*entity.StoredOperation
	startTime        time.Time
}

const (
	MaxSyncDuration           time.Duration = time.Hour * 24 * 30 // a month
	MaxSyncNumberOfMessages   int           = 1000
	MaxSyncNumberOfOperations int           = 1000
)

func newStateActiveSyncing(p *Peer) *StateActiveSyncing {
	return &StateActiveSyncing{p: p}
}

func (s *StateActiveSyncing) perform() (nextState State, err error) {
	if s.startTime.IsZero() {
		s.startTime = s.getStartTime()
	}
	msgSynced, err := s.messagesSynced()
	if err != nil {
		return nil, err
	}
	if !msgSynced {
		return s.syncMessage()
	}
	operSynced, err := s.operationsSynced()
	if err != nil {
		return nil, err
	}
	if !operSynced {
		return s.syncOperation()
	}
	err = s.sendDone()
	if err != nil {
		log.Errorf("Failed to send done to the peer %s: %v", s.p, err)
		return nil, err
	}
	if s.p.conn.IsActive() {
		return newStatePassiveSyncing(s.p), nil
	}
	return newStateIdle(s.p), nil
}

func (s *StateActiveSyncing) getStartTime() time.Time {
	var res time.Time
	if s.p.hist == nil {
		res = time.Now().Add(-MaxSyncDuration)
	} else {
		res = s.p.hist.Disconnected
		if s.p.Subs.Diff(s.p.hist.Subs) != nil {
			res = res.Add(-MaxSyncDuration)
		}
		log.Debugf("Synchronizing entities with peer %s starting from %s",
			s.p, res.Format(time.RFC3339))
	}
	return res
}

func (s *StateActiveSyncing) messagesSynced() (bool, error) {
	if s.messagesToSync == nil {
		mm, err := s.p.owner.Storage.GetMessagesStoredAfter(s.startTime, MaxSyncNumberOfMessages)
		if err != nil {
			log.Errorf("Failed to fetch messages since %s", s.startTime.Format(time.RFC3339))
			return false, err
		}
		log.Debugf("Found %d message(s) to synchronize with peer %s", len(mm), s.p)
		s.messagesToSync = mm
	}
	return len(s.messagesToSync) == 0, nil
}

func (s *StateActiveSyncing) syncMessage() (nextState State, err error) {
	if len(s.messagesToSync) == 0 {
		log.Fatal("BUG: syncMessage() is called when messagesToSync is empty")
	}
	log.Debugf("%d message(s) left to synchronize with peer %s", len(s.messagesToSync), s.p)
	var m *entity.StoredMessage
	m, s.messagesToSync = s.messagesToSync[0], s.messagesToSync[1:]
	log.Debugf("Sending message %s to peer %s", m.M, s.p)
	return newStateSending(s.p, m.M, m.Stored, s), nil
}

func (s *StateActiveSyncing) operationsSynced() (bool, error) {
	if s.operationsToSync == nil {
		oo, err := s.p.owner.Storage.GetOperationsStoredAfter(s.startTime, MaxSyncNumberOfOperations)
		if err != nil {
			log.Errorf("Failed to fetch operations since %s", s.startTime.Format(time.RFC3339))
			return false, err
		}
		log.Debugf("Found %d operation(s) to synchronize with peer %s", len(oo), s.p)
		s.operationsToSync = oo
	}
	return len(s.operationsToSync) == 0, nil
}

func (s *StateActiveSyncing) syncOperation() (nextState State, err error) {
	if len(s.operationsToSync) == 0 {
		log.Fatal("BUG: syncOperation() is called when operationsToSync is empty")
	}
	log.Debugf("%d operation(s) left to synchronize with peer %s", len(s.operationsToSync), s.p)
	var o *entity.StoredOperation
	o, s.operationsToSync = s.operationsToSync[0], s.operationsToSync[1:]
	log.Debugf("Sending operation %s to peer %s", o.O, s.p)
	return newStateSending(s.p, o.O, o.Stored, s), nil
}

func (s *StateActiveSyncing) sendDone() error {
	pld := packet.NewPayloadDone()
	pkt := packet.New(packet.TypeDone, s.p.User.ID(), pld, s.p.owner.Signer)
	err := s.p.conn.Write(pkt)
	if err != nil {
		log.Errorf("Error sending %s to the peer %s: %v", pkt, s.p, err)
		return err
	}
	return nil
}

func (s *StateActiveSyncing) Name() string {
	return "StateActiveSyncing"
}

func (s *StateActiveSyncing) ID() StateID {
	return StateIDActiveSyncing
}
