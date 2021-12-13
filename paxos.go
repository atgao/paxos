package paxos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"os"
	"sync"

	"github.com/google/uuid"
) // for testing

//
// struct for messages sent between nodes
//

type ProposalNumber struct {
	N        int
	SenderId int
}

type PrepareMessage struct {
	ProposalNumber ProposalNumber
	LogEntryIndex  int
}

type PrepareResponseMessage struct {
	LogEntryIndex          int
	AcceptedProposalNumber ProposalNumber
	AcceptedValue          *LockRelayMessage
	NoMoreAccepted         bool
}

type AcceptMessage struct {
	ProposalNumber     ProposalNumber
	LogEntryIndex      int
	Value              LockRelayMessage
	FirstUnchosenIndex int
}

type AcceptResponseMessage struct {
	LogEntryIndex      int
	MinProposal        ProposalNumber
	FirstUnchosenIndex int
}

type SuccessMessage struct {
	LogEntryIndex int
	Value         LockRelayMessage
}

type SuccessResponseMessage struct {
	LogEntryIndex      int
	FirstUnchosenIndex int
}

type Message struct {
	SenderId        int       // the actual proposal number is ProposalNumber and SenderId
	Uuid            uuid.UUID // matching the message and response
	Prepare         *PrepareMessage
	PrepareResponse *PrepareResponseMessage
	Accept          *AcceptMessage
	AcceptResponse  *AcceptResponseMessage
	Success         *SuccessMessage
	SuccessResponse *SuccessResponseMessage
	/*
		ProposalId         int    // id proposed
		CurrentVal         int    // value for currrent round of proposal
		AcceptId           int    // id accepted
		HighestAcceptedVal string // value from the highest proposer ID acceptor
		Type               string // prepare, propose, accept, accepted etc
		DecidedVal         string // value from the consensus
	*/
}

type LogEntry struct {
	AcceptedProposalNumber ProposalNumber
	AcceptedValue          *LockRelayMessage
}

type AcceptorPersistentState struct {
	mu                  sync.Mutex
	LastLogIndex        int
	MinProposal         ProposalNumber
	Log                 []LogEntry
	FirstUnchosenIndex_ int
}

func (st *AcceptorPersistentState) FirstUnchosenIndex() int {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.FirstUnchosenIndex_
}
func (st *AcceptorPersistentState) SetFirstUnchosenIndex(v int) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.FirstUnchosenIndex_ = v
}

func (st *AcceptorPersistentState) getLog(i int) LogEntry {
	st.mu.Lock()
	defer st.mu.Unlock()
	for i >= len(st.Log) {
		st.Log = append(st.Log, LogEntry{
			AcceptedProposalNumber: ProposalNumber{
				N:        0,
				SenderId: 0,
			},
			AcceptedValue: nil,
		})
	}
	return st.Log[i]
}

func (st *AcceptorPersistentState) logSize() int {
	st.mu.Lock()
	defer st.mu.Unlock()
	return len(st.Log)
}

func (st *AcceptorPersistentState) setLogProposalNumber(i int, pn ProposalNumber) {
	st.getLog(i)
	st.mu.Lock()
	defer st.mu.Unlock()
	st.Log[i].AcceptedProposalNumber = pn
	if pn.N == math.MaxInt {
		log.Fatal("Please call setChooseProposal")
		/*
			for st.firstUnchosenIndex < len(st.log) && st.log[st.firstUnchosenIndex].AcceptedProposalNumber.N == math.MaxInt {
				st.firstUnchosenIndex++
			}
		*/
	}
	if pn.N != 0 && i > st.LastLogIndex {
		st.LastLogIndex = i
	}
}

func (st *AcceptorPersistentState) setLogProposalValue(i int, v LockRelayMessage) {
	st.getLog(i)
	st.mu.Lock()
	defer st.mu.Unlock()
	st.Log[i].AcceptedValue = &v
}

type ProposerPersistentState struct {
	MaxRound int
}

type ProposerVolatileState struct {
	NextIndex int
	Prepared  bool
}

type PaxosNodeState struct {
	AcceptorPersistentState AcceptorPersistentState
	ProposerPersistentState ProposerPersistentState
	ProposerVolatileState   ProposerVolatileState
	LogChan                 chan LockRelayMessage
}

func (state *GlobalState) setChooseProposal(i int, msg LockRelayMessage) {
	st := state.PaxosNodeState
	log.Info(fmt.Sprintf("Lock entry %d chosen as %+v", i, msg))
	st.AcceptorPersistentState.getLog(i)
	st.AcceptorPersistentState.mu.Lock()
	defer st.AcceptorPersistentState.mu.Unlock()
	st.AcceptorPersistentState.Log[i].AcceptedProposalNumber = ProposalNumber{
		N:        math.MaxInt,
		SenderId: 0,
	}
	st.AcceptorPersistentState.Log[i].AcceptedValue = &msg
	for st.AcceptorPersistentState.FirstUnchosenIndex_ < len(st.AcceptorPersistentState.Log) &&
		st.AcceptorPersistentState.Log[st.AcceptorPersistentState.FirstUnchosenIndex_].AcceptedProposalNumber.N == math.MaxInt {
		st.LogChan <- *st.AcceptorPersistentState.Log[st.AcceptorPersistentState.FirstUnchosenIndex_].AcceptedValue
		st.AcceptorPersistentState.FirstUnchosenIndex_++
	}
	if i > st.AcceptorPersistentState.LastLogIndex {
		st.AcceptorPersistentState.LastLogIndex = i
	}
	state.PaxosNodeState.persistentToFile(state.Config.StateFile)
}

func MakePaxosNodeState() *PaxosNodeState {
	return &PaxosNodeState{
		ProposerPersistentState: ProposerPersistentState{MaxRound: 0},
		ProposerVolatileState:   ProposerVolatileState{0, false},
		AcceptorPersistentState: AcceptorPersistentState{
			LastLogIndex:        -1,
			MinProposal:         ProposalNumber{},
			Log:                 nil,
			FirstUnchosenIndex_: 0,
		},
		LogChan: make(chan LockRelayMessage),
	}
}

type PaxosNodePersistentState struct {
	Proposer ProposerPersistentState
	Acceptor AcceptorPersistentState
}

func (st *PaxosNodeState) persistentToFile(filename string) {
	data, err := json.Marshal(PaxosNodePersistentState{
		Proposer: st.ProposerPersistentState,
		Acceptor: AcceptorPersistentState{
			LastLogIndex:        st.AcceptorPersistentState.LastLogIndex,
			MinProposal:         st.AcceptorPersistentState.MinProposal,
			Log:                 st.AcceptorPersistentState.Log,
			FirstUnchosenIndex_: st.AcceptorPersistentState.FirstUnchosenIndex_,
		},
	})
	if err != nil {
		log.Fatal("Failed to persistent PaxosState: " + err.Error())
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Fatal("Failed to persistent PaxosState: " + err.Error())
	}
}

func MakePaxosNodeStateFromPersistentFile(filename string) *PaxosNodeState {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warn("Failed to read persistent PaxosState: " + err.Error())
		return MakePaxosNodeState()
	}
	var state PaxosNodePersistentState
	err = json.Unmarshal(data, &state)
	if err != nil {
		log.Warn("Failed to read persistent PaxosState: " + err.Error())
		return MakePaxosNodeState()
	}
	return &PaxosNodeState{
		ProposerPersistentState: state.Proposer,
		AcceptorPersistentState: AcceptorPersistentState{
			LastLogIndex:        state.Acceptor.LastLogIndex,
			MinProposal:         state.Acceptor.MinProposal,
			Log:                 state.Acceptor.Log,
			FirstUnchosenIndex_: state.Acceptor.FirstUnchosenIndex_,
		},
		ProposerVolatileState: ProposerVolatileState{0, false},
		LogChan:               make(chan LockRelayMessage),
	}
}

func (p ProposalNumber) GE(rhs ProposalNumber) bool {
	return p.N > rhs.N || (p.N == rhs.N && p.SenderId > rhs.SenderId)
}

func (p ProposalNumber) GEq(rhs ProposalNumber) bool {
	return p.GE(rhs) || p == rhs
}

func (state *GlobalState) ProcessPrepareMessage(msg PrepareMessage) PrepareResponseMessage {
	if msg.ProposalNumber.GEq(state.PaxosNodeState.AcceptorPersistentState.MinProposal) {
		state.PaxosNodeState.AcceptorPersistentState.MinProposal = msg.ProposalNumber
	}
	noMoreAccepted := true
	for i := msg.LogEntryIndex + 1; i != state.PaxosNodeState.AcceptorPersistentState.logSize(); i++ {
		if state.PaxosNodeState.AcceptorPersistentState.getLog(i).AcceptedProposalNumber.N != 0 {
			noMoreAccepted = false
		}
	}
	state.PaxosNodeState.persistentToFile(state.Config.StateFile)
	return PrepareResponseMessage{
		AcceptedProposalNumber: state.PaxosNodeState.AcceptorPersistentState.getLog(msg.LogEntryIndex).AcceptedProposalNumber,
		AcceptedValue:          state.PaxosNodeState.AcceptorPersistentState.getLog(msg.LogEntryIndex).AcceptedValue,
		NoMoreAccepted:         noMoreAccepted,
	}
}

func (state *GlobalState) ProcessAcceptMessage(msg AcceptMessage) AcceptResponseMessage {
	if msg.ProposalNumber.GEq(state.PaxosNodeState.AcceptorPersistentState.MinProposal) {
		logEntry := state.PaxosNodeState.AcceptorPersistentState.getLog(msg.LogEntryIndex)
		if logEntry.AcceptedProposalNumber.N == math.MaxInt {
			if logEntry.AcceptedValue.MsgUUID != msg.Value.MsgUUID {
				log.Fatal("Trying to replace an chosen proposal")
			}
		} else {
			state.PaxosNodeState.AcceptorPersistentState.setLogProposalNumber(msg.LogEntryIndex, msg.ProposalNumber)
			state.PaxosNodeState.AcceptorPersistentState.setLogProposalValue(msg.LogEntryIndex, msg.Value)
			log.Info(fmt.Sprintf("Log %d updated with accept message %+v", msg.LogEntryIndex,
				state.PaxosNodeState.AcceptorPersistentState.getLog(msg.LogEntryIndex)))
			state.PaxosNodeState.AcceptorPersistentState.MinProposal = msg.ProposalNumber
		}
	}
	firstUnchosenIndex := state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex()
	for index := firstUnchosenIndex; index < msg.FirstUnchosenIndex; index++ {
		if state.PaxosNodeState.AcceptorPersistentState.getLog(index).AcceptedProposalNumber == msg.ProposalNumber {
			state.setChooseProposal(index,
				*state.PaxosNodeState.AcceptorPersistentState.getLog(index).AcceptedValue)
			// state.PaxosNodeState.AcceptorPersistentState.setLogProposalNumber(index, ProposalNumber{N: math.MaxInt})
			log.Info(fmt.Sprintf("Log %d updated with accept message %+v", index,
				state.PaxosNodeState.AcceptorPersistentState.getLog(index)))
		}
	}
	state.PaxosNodeState.persistentToFile(state.Config.StateFile)
	return AcceptResponseMessage{
		MinProposal:        state.PaxosNodeState.AcceptorPersistentState.MinProposal,
		FirstUnchosenIndex: firstUnchosenIndex,
	}
}
func (state *GlobalState) ProcessSuccessMessage(msg SuccessMessage) SuccessResponseMessage {
	logEntry := state.PaxosNodeState.AcceptorPersistentState.getLog(msg.LogEntryIndex)
	if logEntry.AcceptedProposalNumber.N == math.MaxInt {
		if logEntry.AcceptedValue.MsgUUID != msg.Value.MsgUUID {
			log.Fatal("Trying to replace an chosen proposal")
		}
	} else {
		state.setChooseProposal(msg.LogEntryIndex, msg.Value)
		// state.PaxosNodeState.AcceptorPersistentState.setLogProposalNumber(msg.LogEntryIndex, ProposalNumber{N: math.MaxInt})
		// state.PaxosNodeState.AcceptorPersistentState.setLogProposalValue(msg.LogEntryIndex, msg.Value)
		log.Info(fmt.Sprintf("Log %d updated with success message %+v", msg.LogEntryIndex,
			state.PaxosNodeState.AcceptorPersistentState.getLog(msg.LogEntryIndex)))
	}
	state.PaxosNodeState.persistentToFile(state.Config.StateFile)
	return SuccessResponseMessage{FirstUnchosenIndex: state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex()}
}

func (state *GlobalState) SendSuccessMessage(firstUnchosenIndex int, targetId int) {
	successUUID := uuid.New()

	dispatcher := MakeDispatcher(func(msg Message) bool {
		return msg.Uuid == successUUID
	})
	AddPaxosMessageDispatcher(state, dispatcher)

	sendGenericMessage(state.InterNodeUDPSock, state.Config.PeerAddress[targetId], GenericMessage{Paxos: &Message{
		SenderId: state.Config.SelfId,
		Uuid:     successUUID,
		Success: &SuccessMessage{
			LogEntryIndex: firstUnchosenIndex,
			Value:         *state.PaxosNodeState.AcceptorPersistentState.getLog(firstUnchosenIndex).AcceptedValue,
		},
	}})

	reply := <-dispatcher.ch
	RemovePaxosMessageDispatcher(state, dispatcher)
	state.PaxosNodeState.persistentToFile(state.Config.StateFile)
	if reply.SuccessResponse.FirstUnchosenIndex < state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex() {
		state.SendSuccessMessage(reply.SuccessResponse.FirstUnchosenIndex, targetId)
	}
}

func (state *GlobalState) majorityReached(slice []Message) bool {
	return len(slice) >= (len(state.Config.PeerAddress))/2
}

func (state *GlobalState) Broadcast(msg Message) {
	BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), msg)
}
func (state *GlobalState) ProposerAlgorithm(inputValue LockRelayMessage) bool {
	log.Info("Enter proposer algorithm")
	if state.HeartBeatState.CurrentLeaderId(state.Config) != state.Config.SelfId {
		log.Warn("Not leader, exit proposer algorithm")
		return false
	}

	for {
		var index int
		var value LockRelayMessage
		var NumberOfNoMoreAccepted int
		var pn ProposalNumber
		var Majorityqueue []Message

		if state.PaxosNodeState.ProposerVolatileState.Prepared == true {
			index = state.PaxosNodeState.ProposerVolatileState.NextIndex
			state.PaxosNodeState.ProposerVolatileState.NextIndex += 1
			pn = ProposalNumber{state.PaxosNodeState.ProposerPersistentState.MaxRound, state.Config.SelfId}
			value = inputValue
			log.Info("Already prepared, fast path, proposal number: %v", pn)
		} else {
			index = state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex()
			state.PaxosNodeState.ProposerVolatileState.NextIndex = index + 1
			state.PaxosNodeState.ProposerPersistentState.MaxRound++
			state.PaxosNodeState.persistentToFile(state.Config.StateFile)
			pn = ProposalNumber{N: state.PaxosNodeState.ProposerPersistentState.MaxRound, SenderId: state.Config.SelfId}
			log.Info("Not prepared, slow path, proposal number: %v", pn)

			prepareUUID := uuid.New()
			PrepareMsg := PrepareMessage{ProposalNumber: pn, LogEntryIndex: index}

			Msg := Message{SenderId: state.Config.SelfId, Uuid: prepareUUID, Prepare: &PrepareMsg}

			dispatcher := MakeUUIDDispatcher(prepareUUID)
			AddPaxosMessageDispatcher(state, dispatcher)

			state.Broadcast(Msg)

			for {

				NewPaxosMessage := <-dispatcher.ch
				Majorityqueue = append(Majorityqueue, NewPaxosMessage)

				if state.majorityReached(Majorityqueue) == true {
					NumberOfNoMoreAccepted = 0
					maxAcceptedProposal := Majorityqueue[0].PrepareResponse.AcceptedProposalNumber
					var maxIndex = -1

					for i := 0; i < len(Majorityqueue); i++ {
						if Majorityqueue[i].PrepareResponse.AcceptedProposalNumber.GEq(maxAcceptedProposal) {
							maxAcceptedProposal = Majorityqueue[i].PrepareResponse.AcceptedProposalNumber
							maxIndex = i
						}
						if Majorityqueue[i].PrepareResponse.NoMoreAccepted == true {

							NumberOfNoMoreAccepted += 1
						}
					}
					if maxAcceptedProposal.N != 0 {
						value = *Majorityqueue[maxIndex].PrepareResponse.AcceptedValue
					} else {
						value = inputValue
					}
					if len(Majorityqueue) == NumberOfNoMoreAccepted {
						state.PaxosNodeState.ProposerVolatileState.Prepared = true
					}
					RemovePaxosMessageDispatcher(state, dispatcher)
					break
				}

			}
		}

		acceptUUID := uuid.New()
		dispatcher := MakeUUIDDispatcher(acceptUUID)
		AddPaxosMessageDispatcher(state, dispatcher)

		acceptMessage := &AcceptMessage{
			ProposalNumber:     pn,
			LogEntryIndex:      index,
			Value:              value,
			FirstUnchosenIndex: state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex(),
		}
		state.Broadcast(Message{
			SenderId: state.Config.SelfId, Accept: acceptMessage,
			Uuid: acceptUUID,
		})

		goodToContinue := make(chan bool)
		p := func() {
			sent := false
			for num := 0; num != len(state.Config.PeerAddress); num++ {
				send := false
				var res bool
				if num == len(state.Config.PeerAddress)/2 {
					send = true
					res = true
				}
				reply := <-dispatcher.ch
				acceptReply := reply.AcceptResponse
				if acceptReply.MinProposal.GE(pn) {
					state.PaxosNodeState.ProposerPersistentState.MaxRound = acceptReply.MinProposal.N
					state.PaxosNodeState.persistentToFile(state.Config.StateFile)
					state.PaxosNodeState.ProposerVolatileState.Prepared = false
					send = true
					res = false
				}
				if acceptReply.FirstUnchosenIndex <= state.PaxosNodeState.AcceptorPersistentState.LastLogIndex &&
					state.PaxosNodeState.AcceptorPersistentState.getLog(acceptReply.FirstUnchosenIndex).AcceptedProposalNumber.N == math.MaxInt {
					state.SendSuccessMessage(acceptReply.FirstUnchosenIndex, reply.SenderId)
				}
				if send && !sent {
					sent = true
					goodToContinue <- res
					close(goodToContinue)
				}
			}
			RemovePaxosMessageDispatcher(state, dispatcher)
		}
		go p()

		if <-goodToContinue {
			state.setChooseProposal(index, value)
			/*
				state.PaxosNodeState.AcceptorPersistentState.setLogProposalNumber(index, ProposalNumber{
					N:        math.MaxInt,
					SenderId: 0,
				})
				state.PaxosNodeState.AcceptorPersistentState.setLogProposalValue(index, value)
			*/
		} else {
			continue
		}

		if value.MsgUUID == inputValue.MsgUUID {
			state.PaxosNodeState.persistentToFile(state.Config.StateFile)
			return true
		}
	}
}
