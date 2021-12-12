package paxos

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type GlobalState struct {
	Config                  *Config
	HeartBeatState          *HeartBeatState
	InterNodeUDPSock        *net.UDPConn
	ClientFacingUDPSock     *net.UDPConn
	MessageQueue            chan GenericMessage
	PaxosMessageQueue       chan Message
	KeepAliveMessageQueue   chan KeepAliveMessage
	LockRelayMessageQueue   chan LockRelayMessage
	LockState               *LockState
	PaxosNodeState          *PaxosNodeState
	PaxosMessageDispatchers map[*Dispatcher]bool
}

type Dispatcher struct {
	Filter func(Message) bool
	ch     chan Message
}

func MakeDispatcher(filter func(Message) bool) *Dispatcher {
	return &Dispatcher{filter, make(chan Message)}
}

func DispatchPaxosMessage(state *GlobalState) {
	for {
		select {
		case msg := <-state.PaxosMessageQueue:
			for dispatcher, _ := range state.PaxosMessageDispatchers {
				if dispatcher.Filter(msg) {
					dispatcher.ch <- msg
					break
				}
			}
		}
	}
}

func AddPaxosMessageDispatcher(state *GlobalState, dispatcher *Dispatcher) {
	state.PaxosMessageDispatchers[dispatcher] = true
}

func RemovePaxosMessageDispatcher(state *GlobalState, dispatcher *Dispatcher) {
	delete(state.PaxosMessageDispatchers, dispatcher)
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
				log.Info(fmt.Sprintf("Received paxos message: %+v", *genericMessage.Paxos))
				paxosMessageQueue <- *genericMessage.Paxos
			} else if genericMessage.KeepAlive != nil {
				log.Info(fmt.Sprintf("Received keep alive message: %+v", *genericMessage.KeepAlive))
				keepAliveMessageQueue <- *genericMessage.KeepAlive
			} else if genericMessage.LockRelay != nil {
				log.Info(fmt.Sprintf("Received lock message: %+v", *genericMessage.LockRelay))
				lockRelayMessageQueue <- *genericMessage.LockRelay
			} else {
				log.Warn("Unrecognized message")
			}
		}
	}
}

func LockRelay(state *GlobalState) {
	for {
		select {
		case msg := <-state.LockRelayMessageQueue:
			if state.Config.SelfId == state.HeartBeatState.CurrentLeaderId(state.Config) {
				log.Warn("Should process lock relay message here")
			} else {
				SendLockRelayMessage(state.InterNodeUDPSock, state.HeartBeatState.CurrentLeaderAddress(state.Config), msg)
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
		make(chan Message), make(chan KeepAliveMessage), make(chan LockRelayMessage),
		&LockState{}, MakePaxosNodeState(), make(map[*Dispatcher]bool)}
	sendInitialHeartBeat(state)
	go MessageRouter(state.MessageQueue, state.PaxosMessageQueue, state.KeepAliveMessageQueue, state.LockRelayMessageQueue)
	go KeepAliveWorker(state.InterNodeUDPSock, state.Config, state.HeartBeatState, state.KeepAliveMessageQueue)
	go LockRelay(state)
	go DispatchPaxosMessage(state)
	return state, nil
}
