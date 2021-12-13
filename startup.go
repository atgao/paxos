package paxos

import (
	"fmt"
	"github.com/google/uuid"
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
	return &Dispatcher{filter, make(chan Message, 1024)}
}

func MakeUUIDDispatcher(uuid uuid.UUID) *Dispatcher {
	return MakeDispatcher(func(msg Message) bool {
		return msg.Uuid == uuid
	})
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
	close(dispatcher.ch)
	for {
		v, more := <-dispatcher.ch
		if more {
			state.PaxosMessageQueue <- v
		} else {
			break
		}
	}
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
				log.Debug(fmt.Sprintf("Received paxos message: %+v", *genericMessage.Paxos))
				paxosMessageQueue <- *genericMessage.Paxos
			} else if genericMessage.KeepAlive != nil {
				log.Debug(fmt.Sprintf("Received keep alive message: %+v", *genericMessage.KeepAlive))
				keepAliveMessageQueue <- *genericMessage.KeepAlive
			} else if genericMessage.LockRelay != nil {
				log.Debug(fmt.Sprintf("Received lock message: %+v", *genericMessage.LockRelay))
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
	log.Info("Started message router")
	go KeepAliveWorker(state.InterNodeUDPSock, state.Config, state.HeartBeatState, state.KeepAliveMessageQueue)
	log.Info("Started keep alive worker")
	go LockRelay(state)
	log.Info("Started lock relay")
	go DispatchPaxosMessage(state)
	log.Info("Started paxos message dispatcher")

	prepareDispatcher := MakeDispatcher(func(msg Message) bool {
		return msg.Prepare != nil
	})
	acceptDispatcher := MakeDispatcher(func(msg Message) bool {
		return msg.Accept != nil
	})
	successDispatcher := MakeDispatcher(func(msg Message) bool {
		return msg.Success != nil
	})
	AddPaxosMessageDispatcher(state, prepareDispatcher)
	AddPaxosMessageDispatcher(state, acceptDispatcher)
	AddPaxosMessageDispatcher(state, successDispatcher)

	go func() {
		for {
			select {
			case msg := <-prepareDispatcher.ch:
				log.Info(fmt.Sprintf("Processing prepare %+v message from %d", msg.Uuid, msg.SenderId))
				res := state.ProcessPrepareMessage(*(msg.Prepare))
				if err := sendGenericMessage(state.InterNodeUDPSock, state.Config.PeerAddress[msg.SenderId],
					GenericMessage{Paxos: &Message{
						SenderId:        state.Config.SelfId,
						Uuid:            msg.Uuid,
						PrepareResponse: &res,
					}}); err != nil {
					log.Warn("Failed to send prepare response: " + err.Error())
				}
			case msg := <-acceptDispatcher.ch:
				log.Info(fmt.Sprintf("Processing accept %+v message from %d", msg.Uuid, msg.SenderId))
				res := state.ProcessAcceptMessage(*(msg.Accept))
				if err := sendGenericMessage(state.InterNodeUDPSock, state.Config.PeerAddress[msg.SenderId],
					GenericMessage{Paxos: &Message{
						SenderId:       state.Config.SelfId,
						Uuid:           msg.Uuid,
						AcceptResponse: &res,
					}}); err != nil {
					log.Warn("Failed to send prepare response: " + err.Error())
				}
			case msg := <-successDispatcher.ch:
				log.Info(fmt.Sprintf("Processing success %+v message from %d", msg.Uuid, msg.SenderId))
				res := state.ProcessSuccessMessage(*(msg.Success))
				if err := sendGenericMessage(state.InterNodeUDPSock, state.Config.PeerAddress[msg.SenderId],
					GenericMessage{Paxos: &Message{
						SenderId:        state.Config.SelfId,
						Uuid:            msg.Uuid,
						SuccessResponse: &res,
					}}); err != nil {
					log.Warn("Failed to send prepare response: " + err.Error())
				}
			}
		}
	}()

	return state, nil
}
