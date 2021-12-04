package paxos

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

type GlobalState struct {
	Config                *Config
	HeartBeatState        *HeartBeatState
	InterNodeUDPSock      *net.UDPConn
	ClientFacingUDPSock   *net.UDPConn
	MessageQueue          chan GenericMessage
	PaxosMessageQueue     chan Message
	KeepAliveMessageQueue chan KeepAliveMessage
	LockRelayMessageQueue chan LockRelayMessage
}

func sendInitialHeartBeat(state *GlobalState) {
	BroadcastKeepAliveMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(),
		KeepAliveMessage{state.Config.SelfId})
}

func MessageRouter(messageQueue chan GenericMessage, paxosMessageQueue chan Message,
	keepAliveMessageQueue chan KeepAliveMessage, lockRelayMessageQueue chan LockRelayMessage) {
	for {
		select {
		case genericMessage := <-messageQueue:
			if genericMessage.Paxos != nil {
				log.Info(fmt.Sprintf("Received paxos message: %v", *genericMessage.Paxos))
				paxosMessageQueue <- *genericMessage.Paxos
			} else if genericMessage.KeepAlive != nil {
				log.Info(fmt.Sprintf("Received keep alive message: %v", *genericMessage.KeepAlive))
				keepAliveMessageQueue <- *genericMessage.KeepAlive
			} else if genericMessage.LockRelay != nil {
				log.Info(fmt.Sprintf("Received lock message: %v", *genericMessage.LockRelay))
				lockRelayMessageQueue <- *genericMessage.LockRelay
			} else {
				log.Warn("Unrecognized message")
			}
		}
	}
}

func GlobalInitialize(configData []byte) (*GlobalState, error) {
	config := ConfigFromJSON(configData)
	raddr, err := net.ResolveUDPAddr("udp", config.PaxosAddr)
	if err != nil {
		return nil, err
	}
	pc, err := net.ListenUDP("udp", raddr)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Started paxos server on %s", config.PaxosAddr))

	raddr, err = net.ResolveUDPAddr("udp", config.ServerAddr)
	if err != nil {
		return nil, err
	}
	pcserv, err := net.ListenUDP("udp", raddr)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Started lock server on %s", config.ServerAddr))

	ch := make(chan GenericMessage)
	UDPServeGenericMessage(pc, ch)
	UDPServeLockMessage(config.SelfId, pcserv, ch)
	if err != nil {
		log.Fatal("Failed to listen on UDP")
	}
	state := &GlobalState{config, InitHeartBeatState(config), pc, pcserv, ch,
		make(chan Message), make(chan KeepAliveMessage), make(chan LockRelayMessage)}
	sendInitialHeartBeat(state)
	go MessageRouter(state.MessageQueue, state.PaxosMessageQueue, state.KeepAliveMessageQueue, state.LockRelayMessageQueue)
	go KeepAliveWorker(state.InterNodeUDPSock, state.Config, state.HeartBeatState, state.KeepAliveMessageQueue)
	return state, nil
}
