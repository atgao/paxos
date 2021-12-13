package paxos

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type KeepAliveMessage struct {
	Id int
}

const KeepAlivePeriod = 3 * time.Second
const KeepAliveCheckInterval = 1 * time.Second

type HeartBeatState struct {
	mu            sync.Mutex
	lastAliveTime map[int]time.Time
	alivePeers    map[int]bool
}

func InitHeartBeatState(config *Config) *HeartBeatState {
	state := &HeartBeatState{}
	state.lastAliveTime = make(map[int]time.Time)
	state.alivePeers = make(map[int]bool)
	for k := range config.PeerAddress {
		state.lastAliveTime[k] = time.Unix(0, 0)
		state.alivePeers[k] = false
	}
	return state
}

func KeepAliveWorker(conn *net.UDPConn, config *Config, state *HeartBeatState, ch chan KeepAliveMessage) {
	timer := time.NewTicker(KeepAliveCheckInterval)
	toSendAliveTimer := time.NewTicker(KeepAlivePeriod)
	for {
		select {
		case <-toSendAliveTimer.C:
			BroadcastKeepAliveMessage(conn, config.AllPeerAddresses(),
				KeepAliveMessage{config.SelfId})
		case <-timer.C:
			// check keep alive
			for _, i := range state.AlivePeers() {
				r1 := time.Now()
				r2 := state.lastAliveTime[i]
				r := r1.Sub(r2)
				if r > 2*KeepAlivePeriod {
					log.Info(fmt.Sprintf("Peer %d dead, current alive peers %v, current leader %d", i,
						state.AlivePeers(), state.CurrentLeaderId(config)))
					state.setPeerAliveStatus(i, false)
				}
			}
		case m := <-ch:
			toPrint := false
			if state.alivePeers[m.Id] == false {
				toPrint = true
			}
			state.lastAliveTime[m.Id] = time.Now()
			state.setPeerAliveStatus(m.Id, true)
			if toPrint {
				log.Info(fmt.Sprintf("Received new heartbeat from peer %d, current alive peers %v, current leader %d", m,
					state.AlivePeers(), state.CurrentLeaderId(config)))
			}
		}
	}
}

func (state *HeartBeatState) setPeerAliveStatus(id int, isAlive bool) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.alivePeers[id] = isAlive
}

func (state *HeartBeatState) AlivePeers() []int {
	state.mu.Lock()
	defer state.mu.Unlock()
	m := make([]int, 0, len(state.alivePeers))
	for i, isAlive := range state.alivePeers {
		if isAlive {
			m = append(m, i)
		}
	}
	return m
}

/*
func (state *HeartBeatState) AllAlivePeerAddresses(config *Config) []string {
	state.mu.Lock()
	defer state.mu.Unlock()
	alive := state.AlivePeers()
	addresses := make([]string, len(alive))
	idx := 0
	for _, i := range alive {
		addresses[idx] = config.PeerAddress[i]
		idx++
	}
	return addresses
}
*/

func (state *HeartBeatState) CurrentLeaderId(config *Config) int {
	leader := config.SelfId
	for _, i := range state.AlivePeers() {
		if i > leader {
			leader = i
		}
	}
	return leader
}

func (state *HeartBeatState) CurrentLeaderAddress(config *Config) string {
	return config.PeerAddress[state.CurrentLeaderId(config)]
}
