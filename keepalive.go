package paxos

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type KeepAliveMessage struct {
	Id int
}

const KeepAlivePeriod = 10 * time.Second
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
	for k, _ := range config.PeerAddress {
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
			BroadcastKeepAliveMessage(conn, state.AllAlivePeerAddresses(config),
				KeepAliveMessage{config.SelfId})
			log.Info("Sent heartbeat message")
		case <-timer.C:
			// check keep alive
			for _, i := range state.AlivePeers() {
				r1 := time.Now()
				r2 := state.lastAliveTime[i]
				r := r1.Sub(r2)
				if r > 2*KeepAlivePeriod {
					//if (time.Now().Sub(state.lastAliveTime[i])) > 2*KeepAlivePeriod {
					log.Info(fmt.Sprintf("Peer %d dead", i))
					state.setPeerAliveStatus(i, false)
				}
			}
		case m := <-ch:
			state.lastAliveTime[m.Id] = time.Now()
			log.Info(fmt.Sprintf("Received heartbeat from peer %d", m))
			state.setPeerAliveStatus(m.Id, true)
		}
	}
}

func (state *HeartBeatState) setPeerAliveStatus(id int, isAlive bool) {
	state.alivePeers[id] = isAlive
}

func (state *HeartBeatState) AlivePeers() []int {
	m := make([]int, 0, len(state.alivePeers))
	for i, isAlive := range state.alivePeers {
		if isAlive {
			m = append(m, i)
		}
	}
	return m
}

func (state *HeartBeatState) AllAlivePeerAddresses(config *Config) []string {
	alive := state.AlivePeers()
	addresses := make([]string, len(alive))
	idx := 0
	for _, i := range alive {
		addresses[idx] = config.PeerAddress[i]
		idx++
	}
	return addresses
}

func (state *HeartBeatState) CurrentLeaderId(config *Config) int {
	leader := config.SelfId
	for i := range state.AlivePeers() {
		if i > leader {
			leader = i
		}
	}
	return leader
}
