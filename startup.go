package paxos

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

type GlobalState struct {
	Config                *Config
	HeartBeatState        *HeartBeatState
	sock                  *net.UDPConn
	MessageQueue          chan GenericMessage
	PaxosMessageQueue     chan Message
	KeepAliveMessageQueue chan KeepAliveMessage
}

func sendInitialHeartBeat(state *GlobalState) {
	BroadcastKeepAliveMessage(state.sock, state.Config.AllPeerAddresses(),
		KeepAliveMessage{state.Config.SelfId})
}

func MessageRouter(messageQueue chan GenericMessage, paxosMessageQueue chan Message,
	keepAliveMessageQueue chan KeepAliveMessage) {
	for {
		select {
		case genericMessage := <-messageQueue:
			if genericMessage.Paxos != nil {
				log.Info(fmt.Sprintf("Received paxos message: %v", *genericMessage.Paxos))
				paxosMessageQueue <- *genericMessage.Paxos
			} else if genericMessage.KeepAlive != nil {
				log.Info(fmt.Sprintf("Received keep alive message: %v", *genericMessage.KeepAlive))
				keepAliveMessageQueue <- *genericMessage.KeepAlive
			} else {
				log.Warn("Unrecognized message")
			}
		}
	}
}

func GlobalInitialize(configData []byte) (*GlobalState, error) {
	config := ConfigFromJSON(configData)
	raddr, err := net.ResolveUDPAddr("udp", config.SelfAddress)
	if err != nil {
		return nil, err
	}
	pc, err := net.ListenUDP("udp", raddr)
	if err != nil {
		return nil, err
	}

	ch := make(chan GenericMessage)
	UDPServeGenericMessage(pc, ch)
	if err != nil {
		log.Fatal("Failed to listen on UDP")
	}
	state := &GlobalState{config, InitHeartBeatState(config), pc, ch,
		make(chan Message), make(chan KeepAliveMessage)}
	sendInitialHeartBeat(state)
	go MessageRouter(state.MessageQueue, state.PaxosMessageQueue, state.KeepAliveMessageQueue)
	go KeepAliveWorker(state.sock, state.Config, state.HeartBeatState, state.KeepAliveMessageQueue)
	return state, nil
}
