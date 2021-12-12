package paxos

import (
	"math"
) // for testing

//
// struct for messages sent between nodes
//

type ProposalValue struct {
}

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
	AcceptedValue          ProposalValue
	NoMoreAccepted         bool
}

type AcceptMessage struct {
	ProposalNumber     ProposalNumber
	LogEntryIndex      int
	Value              ProposalValue
	FirstUnchosenIndex int
}

type AcceptResponseMessage struct {
	LogEntryIndex      int
	MinProposal        ProposalNumber
	FirstUnchosenIndex int
}

type SuccessMessage struct {
	LogEntryIndex int
	Value         ProposalValue
}

type SuccessResponseMessage struct {
	LogEntryIndex      int
	FirstUnchosenIndex int
}

type Message struct {
	SenderId        int // the actual proposal number is ProposalNumber and SenderId
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
	AcceptedValue          ProposalValue
}

type AcceptorPersistentState struct {
	LastLogIndex       int
	MinProposal        ProposalNumber
	Log                []LogEntry
	FirstUnchosenIndex int
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
}

func MakePaxosNodeState() *PaxosNodeState {
	return &PaxosNodeState{} // TODO: initialize
}

func (p ProposalNumber) GEq(rhs ProposalNumber) bool {
	return p.N > rhs.N || (p.N == rhs.N && p.SenderId > rhs.SenderId)
}

func (state *GlobalState) ProcessPrepareMessage(msg PrepareMessage) PrepareResponseMessage {
	if msg.ProposalNumber.GEq(state.PaxosNodeState.AcceptorPersistentState.MinProposal) {
		state.PaxosNodeState.AcceptorPersistentState.MinProposal = msg.ProposalNumber
	}
	noMoreAccepted := true
	for i := msg.LogEntryIndex + 1; i != len(state.PaxosNodeState.AcceptorPersistentState.Log); i++ {
		if state.PaxosNodeState.AcceptorPersistentState.Log[i].AcceptedProposalNumber.N != 0 {
			noMoreAccepted = false
		}
	}
	return PrepareResponseMessage{
		AcceptedProposalNumber: state.PaxosNodeState.AcceptorPersistentState.Log[msg.LogEntryIndex].AcceptedProposalNumber,
		AcceptedValue:          state.PaxosNodeState.AcceptorPersistentState.Log[msg.LogEntryIndex].AcceptedValue,
		NoMoreAccepted:         noMoreAccepted,
	}
}

func (state *GlobalState) ProcessAcceptMessage(msg AcceptMessage) AcceptResponseMessage {
	if msg.ProposalNumber.GEq(state.PaxosNodeState.AcceptorPersistentState.MinProposal) {
		state.PaxosNodeState.AcceptorPersistentState.Log[msg.LogEntryIndex] = LogEntry{
			msg.ProposalNumber,
			msg.Value,
		}
		state.PaxosNodeState.AcceptorPersistentState.MinProposal = msg.ProposalNumber
	}
	for index := state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex; index < msg.FirstUnchosenIndex; index++ {
		if state.PaxosNodeState.AcceptorPersistentState.Log[index].AcceptedProposalNumber == msg.ProposalNumber {
			state.PaxosNodeState.AcceptorPersistentState.Log[index].AcceptedProposalNumber = math.MaxInt
		}
	}
	return AcceptResponseMessage{
		MinProposal:        state.PaxosNodeState.AcceptorPersistentState.MinProposal,
		FirstUnchosenIndex: state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex,
	}
}
func (state *GlobalState) ProcessSuccessMessage(msg SuccessMessage) SuccessResponseMessage {
	state.PaxosNodeState.AcceptorPersistentState.Log[msg.LogEntryIndex] = LogEntry{
		ProposalNumber{
			N:        math.MaxInt,
			SenderId: 0,
		},
		msg.Value,
	}
	return SuccessResponseMessage{FirstUnchosenIndex: state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex}
}

func (state *GlobalState) SendSuccessMessage(firstUnchosenIndex int, targetId int) {
	used := false
	dispatcher := MakeDispatcher(func(msg Message) bool {
		if used {
			return false
		}
		if msg.SenderId == targetId && msg.SuccessResponse != nil {
			used = true
			return true
		}
		return false
	})
	AddPaxosMessageDispatcher(state, dispatcher)

	sendGenericMessage(state.InterNodeUDPSock, state.Config.PeerAddress[targetId], GenericMessage{Paxos: &Message{
		SenderId: state.Config.SelfId,
		Success: &SuccessMessage{
			LogEntryIndex: firstUnchosenIndex,
			Value:         state.PaxosNodeState.AcceptorPersistentState.Log[firstUnchosenIndex].AcceptedValue,
		},
	}})

	reply := <-dispatcher.ch
	RemovePaxosMessageDispatcher(state, dispatcher)
	if reply.SuccessResponse.FirstUnchosenIndex < state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex {
		state.SendSuccessMessage(reply.SuccessResponse.FirstUnchosenIndex, targetId)
	}
}

func (state *GlobalState) majorityReached(slice []Message) bool {
	if len(slice) > len(state.Config.PeerAddress)/2+1 {
		return true
	} else {
		return false
	}

}

func (state *GlobalState) Broadcast(msg Message) {
	BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), msg)
}
func (state *GlobalState) ProposerAlgorithm(inputValue ProposalValue) bool {

	var index int
	var n int
	var value ProposalValue
	var NumberOfNoMoreAccepted int
	var Majorityqueue []Message

	if state.HeartBeatState.CurrentLeaderId(state.Config) != state.Config.SelfId {
		return false
	}

	if state.PaxosNodeState.ProposerVolatileState.Prepared == true {
		index = state.PaxosNodeState.ProposerVolatileState.NextIndex
		state.PaxosNodeState.ProposerVolatileState.NextIndex += 1
	} else {
		index = state.PaxosNodeState.AcceptorPersistentState.FirstUnchosenIndex
		state.PaxosNodeState.ProposerVolatileState.NextIndex = (index + 1)
		n = state.PaxosNodeState.ProposerPersistentState.MaxRound + 1

		pn := ProposalNumber{N: n, SenderId: state.Config.SelfId}

		PrepareMsg := PrepareMessage{ProposalNumber: &pn, LogEntryIndex: &index}

		Msg := Message{SenderId: state.Config.SelfId, Prepare: &PrepareMsg}

		dispatcher := MakeDispatcher(func(msg Message) bool {
			if msg.PrepareResponse != nil {
				if msg.PrepareResponse.AcceptedProposalNumber == pn && msg.PrepareResponse.LogEntryIndex == index {
					return true
				}
				return false
			}
			return false
		})
		AddPaxosMessageDispatcher(state, dispatcher)

		state.Broadcast(Msg)

		defer RemovePaxosMessageDispatcher(state, dispatcher)

		for {

			NewPaxosMessage := <-dispatcher.ch
			Majorityqueue = append(Majorityqueue, NewPaxosMessage)

			if state.majorityReached(Majorityqueue) == true {
				NumberOfNoMoreAccepted := 0
				maxAcceptedProposal := Majorityqueue[0].PrepareResponse.AcceptedProposalNumber.N
				var maxIndex = -1

				for i := 0; i < len(Majorityqueue); i++ {
					if Majorityqueue[i].PrepareResponse.AcceptedProposalNumber.N > maxAcceptedProposal {
						maxAcceptedProposal = Majorityqueue[i].PrepareResponse.AcceptedProposalNumber.N
						maxIndex = i
					}
					if Majorityqueue[i].PrepareResponse.NoMoreAccepted == true {
						NumberOfNoMoreAccepted += 1
					}
				}
				if maxAcceptedProposal != 0 {
					value = Majorityqueue[maxIndex].PrepareResponse.AcceptedValue
				} else {
					value = inputValue
				}
				if len(Majorityqueue) == NumberOfNoMoreAccepted {
					state.PaxosNodeState.ProposerVolatileState.Prepared = true
				}
				break
			}

		}
		//placeholder

	}

}
