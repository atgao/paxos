package paxos

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

type request struct {
	Id   string
	addr *net.UDPAddr
}

type LockState struct {
	mu         sync.Mutex
	lockholder []*string
	requests   [][]request
	waiting    map[string][]int
}

func MakeLockState(num int) *LockState {
	return &LockState{
		lockholder: make([]*string, num),
		requests:   make([][]request, num),
		waiting:    make(map[string][]int),
	}
}

/*
type LockTransitionResultType int
const (
	T_SUCCESS      LockTransitionResultType = 0
	T_NO_SUCH_LOCK LockTransitionResultType = 1
	T_ALREADY_HELD LockTransitionResultType = 2
	T_WAIT         LockTransitionResultType = 3
	T_FAIL         LockTransitionResultType = 4
)
*/

type LockTransitionResult struct {
	Type LockResult
	addr *net.UDPAddr
}

func (lockState *LockState) transition(msg LockRelayMessage) []LockTransitionResult {
	lockState.mu.Lock()
	defer lockState.mu.Unlock()
	if msg.LockId < 0 || msg.LockId >= len(lockState.lockholder) {
		return []LockTransitionResult{{NO_SUCH_LOCK, msg.ClientAddr}}
	}
	if msg.Lock == true {
		if lockState.lockholder[msg.LockId] == nil {
			lockState.lockholder[msg.LockId] = &msg.ClientID
			return []LockTransitionResult{{SUCCESS, msg.ClientAddr}}
		} else if *lockState.lockholder[msg.LockId] == msg.ClientID {
			return []LockTransitionResult{{ALREADY_HELD, msg.ClientAddr}}
		} else {
			// TODO: Detect deadlock
			waitingSet := make(map[string]bool)
			queue := []string{*lockState.lockholder[msg.LockId]}

			for len(queue) > 0 {
				hd := queue[0]
				queue = queue[1:]

				if !waitingSet[hd] {
					for _, l := range lockState.waiting[hd] {
						holder := *lockState.lockholder[l]
						if holder == msg.ClientID {
							return []LockTransitionResult{{DEADLOCK, msg.ClientAddr}}
						}
						queue = append(queue, holder)
					}
					waitingSet[hd] = true
				}
			}
			lockState.requests[msg.LockId] =
				append(lockState.requests[msg.LockId], request{msg.ClientID, msg.ClientAddr})
			if lockState.waiting[msg.ClientID] == nil {
				lockState.waiting[msg.ClientID] = make([]int, 0)
			}
			lockState.waiting[msg.ClientID] = append(lockState.waiting[msg.ClientID], msg.LockId)
			return []LockTransitionResult{}
		}
	} else {
		if lockState.lockholder[msg.LockId] == nil {
			return []LockTransitionResult{{FAIL, msg.ClientAddr}}
		} else if *lockState.lockholder[msg.LockId] == msg.ClientID {
			if len(lockState.requests[msg.LockId]) == 0 {
				lockState.lockholder[msg.LockId] = nil
				// more
				return []LockTransitionResult{{SUCCESS, msg.ClientAddr}}
			} else {
				lockState.lockholder[msg.LockId] = &lockState.requests[msg.LockId][0].Id
				ret := []LockTransitionResult{
					{SUCCESS, msg.ClientAddr},
					{SUCCESS, lockState.requests[msg.LockId][0].addr},
				}
				lockState.requests[msg.LockId] = lockState.requests[msg.LockId][1:]
				return ret
			}
		} else {
			return []LockTransitionResult{{FAIL, msg.ClientAddr}}
		}
	}
}

func (lockState *LockState) transitionLog(log []LockRelayMessage) [][]LockTransitionResult {
	ret := make([][]LockTransitionResult, len(log))
	for i, v := range log {
		ret[i] = lockState.transition(v)
	}
	return ret
}

func ResponseLockRelayMessages(servConn *net.UDPConn, selfId int, lockLog []LockRelayMessage, lockRes [][]LockTransitionResult) {
	if len(lockLog) != len(lockRes) {
		panic("log should have the same length as lockRes")
	}
	for _, rl := range lockRes {
		for _, r := range rl {
			buffer, err := json.Marshal(LockReplyMessage{r.Type, nil})
			if err != nil {
				log.Fatal("Failed to encode lock result: " + err.Error())
			}
			if err := sendOneAddr(servConn, r.addr, buffer); err != nil {
				log.Warn("Failed to send result to the client: " + err.Error())
			}
		}
	}
}

func CommitLog(lockState *LockState, servConn *net.UDPConn, selfId int, lockLog []LockRelayMessage) {
	lockRes := lockState.transitionLog(lockLog)
	ResponseLockRelayMessages(servConn, selfId, lockLog, lockRes)
}
