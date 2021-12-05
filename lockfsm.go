package paxos

import (
	"bytes"
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"net"
)

type LockState struct {
	lockholder *net.UDPAddr
}

func UDPAddrEq(addr1 net.UDPAddr, addr2 net.UDPAddr) bool {
	return addr1.String() == addr2.String()
}

func (lockState LockState) transition(msg LockRelayMessage) (bool, LockState) {
	if msg.Lock == true {
		if lockState.lockholder == nil {
			return true, LockState{msg.ClientAddr}
		} else {
			return false, lockState
		}
	} else {
		if lockState.lockholder == nil {
			return false, lockState
		} else if UDPAddrEq(*lockState.lockholder, *msg.ClientAddr) {
			return true, LockState{nil}
		} else {
			return false, lockState
		}
	}
}

func (lockState LockState) transitionLog(log []LockRelayMessage) ([]bool, LockState) {
	ret := make([]bool, len(log))
	currentLockState := lockState
	for i, v := range log {
		ret[i], currentLockState = currentLockState.transition(v)
	}
	return ret, currentLockState
}

func ResponseLockRelayMessages(servConn *net.UDPConn, selfId int, lockLog []LockRelayMessage, lockRes []bool) {
	if len(lockLog) != len(lockRes) {
		panic("log should have the same length as lockRes")
	}
	for i := range lockLog {
		if lockLog[i].OriginServerId == selfId {
			var buffer bytes.Buffer
			enc := gob.NewEncoder(&buffer)
			if err := enc.Encode(lockRes[i]); err != nil {
				log.Fatal("Failed to encode lock result: " + err.Error())
			}
			if err := sendOneAddr(servConn, lockLog[i].ClientAddr, buffer); err != nil {
				log.Warn("Failed to send result to the client: " + err.Error())
			}
		}
	}
}

func CommitLog(lockState *LockState, servConn *net.UDPConn, selfId int, lockLog []LockRelayMessage) {
	lockRes, newLog := lockState.transitionLog(lockLog)
	ResponseLockRelayMessages(servConn, selfId, lockLog, lockRes)
	*lockState = newLog
}
