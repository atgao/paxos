package paxos

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
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
	MinProposal        ProposalNumber
	FirstUnchosenIndex int
}

type SuccessMessage struct {
	LogEntryIndex int
	Value         ProposalValue
}

type SuccessResponseMessage struct {
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

/*
// Go object implementing single Paxos node
type Paxos struct {
	me int // id in the peers array

	// proposalId is also the highest one seen so far
	acceptedProposal string          // the number of the last proposal the server has accepted for this entry
	acceptedVal      int             // the value in the last proposal the server accepted for this entry
	acceptedMessages map[int]Message //not sure it's still needed
	lastLogIndex     int             //the largest entry for which this server has accepted a proposal
	minProposal      int             //the number of the smallest proposal this server will accept for any log entry
	globalstate      *GlobalState
	role             string // leader (proposer) or acceptor
}

func Make(me int, state *GlobalState) *Paxos {
	px := &Paxos{}      // init the paxos node
	px.me = me          // my index in the sendQueue array
	px.role = "A"       // for acceptor ??? may need to fix...
	px.minProposal = -1 // this way when we start its 0 (Kaiming: bookmark for now)
	px.acceptedVal = -1
	px.acceptedMessages = make(map[int]Message)
	px.globalstate = state
	go px.run(state)
	return px
}
*/

// this go routine keeps on running in the background
func (px *Paxos) run(state *GlobalState) {
	// If no consensus were reached
	for {
		NewPaxosMessage := <-state.PaxosMessageQueue
		log.Info(fmt.Sprintf("Received paxos message to run: %v", NewPaxosMessage))
		switch NewPaxosMessage.Type {
		case "prepare": // proposer --> acceptor
			log.Info(fmt.Sprintf("Proposer start run... val:", NewPaxosMessage.ProposalId))
			px.runAcceptor(NewPaxosMessage)

			//Proposor send prepare message to acceptor to reach majority consensus.
		case "promise":
			px.runPromisor(NewPaxosMessage)

		case "accept": // proposer --> acceptor
			log.Info(fmt.Sprintf("acceptor start run... val:", NewPaxosMessage.CurrentVal))

			px.runAcceptor(NewPaxosMessage)

		case "accepted":
			// px.runLearner(NewPaxosMessage)
		}

	}

	// }
}
func (px *Paxos) runPromisor(msg Message) {
	if msg.ProposalId > px.minProposal {
		px.minProposal = msg.ProposalId
		log.Info(fmt.Sprintf("promise received\n"))
	}

}

func (px *Paxos) ReturnValue() interface{} {
	return px.acceptedVal
}

func (px *Paxos) Prepare(state *GlobalState) {

	//creating an array of messages, each message have different To target
	//right now we just send the message to every acceptor (m=3))
	log.Info(fmt.Sprintf("Prepare is called...\n"))
	msg := Message{
		Type:       "prepare",
		ProposalId: px.minProposal + 1,
		CurrentVal: 10,
	}
	BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), msg)

	// send to network so can send to others
}

// // function for learner
// func (px *Paxos) runLearner(msg Message) string {

// 	for {
// 		if msg.Val == "" {
// 			continue
// 		}
// 		if msg.ProposalId > px.proposalId {
// 			px.proposalId = msg.ProposalId
// 		}
// 		learnMsg, islearn := px.choseMajority()
// 		if islearn == false {
// 			continue
// 		}
// 		px.acceptedVal = learnMsg.Val
// 		return learnMsg.Val
// 	}
// }

// func (px *Paxos) choseMajority() (Message, bool) {
// 	//need to loop through all accepted message
// 	CountResult := make(map[int]int)
// 	MessageResult := make(map[int]Message)

// 	for _, Msg := range px.acceptedMessages {
// 		ProposalID := Msg.ProposalId
// 		CountResult[ProposalID] += 1
// 		MessageResult[ProposalID] = Msg
// 	}

// 	for ChosenID, ChosenMsg := range MessageResult {
// 		fmt.Printf("Proposal[%v] Message[%s]\n", ChosenID, ChosenMsg.Val)
// 		if CountResult[ChosenID] > px.quorum() {
// 			return ChosenMsg, true
// 		}
// 	}
// 	return Message{}, false
// }

// function for acceptor
func (px *Paxos) runAcceptor(msg Message) {

	switch msg.Type {
	case "prepare": // phase 1
		log.Info(fmt.Sprintf("[proposer:%d] phase 1, prepareMsg:%v", px.minProposal, msg.ProposalId))

		if msg.ProposalId > px.minProposal {
			px.minProposal = msg.ProposalId
			promiseMessage := Message{
				Type:       "promise",
				ProposalId: msg.ProposalId,
				AcceptId:   px.minProposal,
				CurrentVal: px.acceptedVal,
			}
			log.Info(fmt.Sprintf("promise returned"))
			BroadcastPaxosMessage(px.globalstate.InterNodeUDPSock, px.globalstate.Config.AllPeerAddresses(), promiseMessage)
		}
		//otherwise reject

	case "accept": // phase 2
		log.Info(fmt.Sprintf("[proposer:%d] phase 2, acceptMsg current val:%v", px.me, msg.CurrentVal))

		if msg.ProposalId >= px.minProposal {
			px.minProposal = msg.ProposalId
			px.acceptedVal = msg.CurrentVal
			log.Info(fmt.Sprintf("[proposer:%d] Value accepted, current val:%v", px.me, px.ReturnValue()))

			acceptedMsg := Message{
				Type:       "accepted",
				ProposalId: msg.ProposalId,
				AcceptId:   msg.ProposalId,
				CurrentVal: msg.CurrentVal,
			}
			BroadcastPaxosMessage(px.globalstate.InterNodeUDPSock, px.globalstate.Config.AllPeerAddresses(), acceptedMsg)
		}
	}
}

// // function for determining if we reach majority
// func (px *Paxos) MajorityReach() bool {
// 	return true
// }

// functions for proposer/leader
// TODO / NOTE: only call prepare when starting election

// TODO / NOTE: only the leader should be calling this
func (px *Paxos) Propose(s int) {

	// if px.role != "L" {
	// 	return
	// }

	px.minProposal += 1
	// message type is accept bc we want the acceptors
	// to accept
	msg := Message{
		Type:       "accept",
		ProposalId: px.minProposal,
		CurrentVal: s,
	}
	log.Info(fmt.Sprintf("accept sended %+v\n", msg))
	BroadcastPaxosMessage(px.globalstate.InterNodeUDPSock, px.globalstate.Config.AllPeerAddresses(), msg)

}

// function for majority threshold
func quorum(state *GlobalState) int {
	return len(state.HeartBeatState.AllAlivePeerAddresses(state.Config))/2 + 1
}

//function for clean the accepted value
func (px *Paxos) clean() {
	// px.proposalId = 0
	// px.proposalAccepted = false
	// px.acceptedMessages = nil
	// px.acceptedVal = nil
}
