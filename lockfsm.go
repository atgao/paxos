package paxos

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

type LockState struct {
	mu         sync.Mutex
	lockholder *net.UDPAddr
}

func UDPAddrEq(addr1 net.UDPAddr, addr2 net.UDPAddr) bool {
	return addr1.String() == addr2.String()
}

func (lockState *LockState) transition(msg LockRelayMessage) bool {
	lockState.mu.Lock()
	defer lockState.mu.Unlock()
	if msg.Lock == true {
		if lockState.lockholder == nil {
			lockState.lockholder = msg.ClientAddr
			return true
		} else {
			return false
		}
	} else {
		if lockState.lockholder == nil {
			return false
		} else if UDPAddrEq(*lockState.lockholder, *msg.ClientAddr) {
			lockState.lockholder = nil
			return true
		} else {
			return false
		}
	}
}

func (lockState *LockState) transitionLog(log []LockRelayMessage) []bool {
	ret := make([]bool, len(log))
	for i, v := range log {
		ret[i] = lockState.transition(v)
	}
	return ret
}

func ResponseLockRelayMessages(servConn *net.UDPConn, selfId int, lockLog []LockRelayMessage, lockRes []bool) {
	if len(lockLog) != len(lockRes) {
		panic("log should have the same length as lockRes")
	}
	for i := range lockLog {
		if lockLog[i].OriginServerId == selfId {
			buffer, err := json.Marshal(LockReplyMessage{lockRes[i], nil})
			if err != nil {
				log.Fatal("Failed to encode lock result: " + err.Error())
			}
			if err := sendOneAddr(servConn, lockLog[i].ClientAddr, buffer); err != nil {
				log.Warn("Failed to send result to the client: " + err.Error())
			}
		}
	}
}

func CommitLog(lockState *LockState, servConn *net.UDPConn, selfId int, lockLog []LockRelayMessage) {
	lockRes := lockState.transitionLog(lockLog)
	ResponseLockRelayMessages(servConn, selfId, lockLog, lockRes)
}
