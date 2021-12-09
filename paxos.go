package paxos

import (
	"fmt"
) // for testing

//
// struct for messages sent between nodes
//

type Message struct {
	Type               string      // prepare, propose, accept, accepted etc
	ProposalId         int         // id proposed
	currentVal         interface{} // value for currrent round of proposal
	AcceptId           int         // id accepted
	highestAcceptedVal interface{} // value from the highest proposer ID acceptor
	decidedVal         interface{} // value from the consensus
}

// Go object implementing single Paxos node
type Paxos struct {
	me int // id in the peers array

	// proposalId is also the highest one seen so far
	proposalId       int             // NOTE THAT THIS SHOULD ALWAYS BE INCREASING
	proposalAccepted bool            // already accepted proposal ??
	acceptedVal      interface{}     //value we accept
	acceptedMessages map[int]Message //not sure it's still needed

	role string // leader (proposer) or acceptor
}

func Make(me int, state *GlobalState) *Paxos {
	px := &Paxos{}     // init the paxos node
	px.me = me         // my index in the sendQueue array
	px.role = "A"      // for acceptor ??? may need to fix...
	px.proposalId = -1 // this way when we start its 0 (Kaiming: bookmark for now)
	px.proposalAccepted = false
	px.acceptedVal = nil
	px.acceptedMessages = make(map[int]Message)
	go px.run(state)
	return px
}

// this go routine keeps on running in the background
func (px *Paxos) run(state *GlobalState) {
	// If no consensus were reached
	for {
		NewPaxosMessage := <-state.PaxosMessageQueue
		fmt.Printf("Received paxos message: %v", NewPaxosMessage)
	}
	// switch msg.Type {
	// case "prepare": // proposer --> acceptor
	// 	fmt.Println("Proposer start run... val:", msg.ProposalId)
	// 	//Proposor send prepare message to acceptor to reach majority consensus.
	// 	//need a condition to check if majority reached

	// case "accept": // proposer --> acceptor
	// 	px.runAcceptor(msg)

	// // // do some stuff here for proposer
	// case "promise":
	// // 	px.proposer <- msg

	// case "accepted":
	// 	px.runLearner(msg)
	// }

	// }
}
func (px *Paxos) Prepare() Message {

	//creating an array of messages, each message have different To target
	//right now we just send the message to every acceptor (m=3))
	fmt.Printf("Prepare is called...\n")
	px.proposalId += 1
	msg := Message{
		Type:       "prepare",
		ProposalId: px.proposalId,
	}
	return msg
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

// // function for acceptorgo bui
// func (px *Paxos) runAcceptor(msg Message) {

// 	switch msg.Type {
// 	case "prepare": // phase 1
// 		if msg.ProposalId > px.proposalId {
// 			px.proposalId = msg.ProposalId // update proposal

// 			// TODO: update proposal accepted boolean??
// 			// TODO: check if the fields are right too...
// 			// TODO: may need to fix the accept id

// 			promiseMessage := Message{
// 				Type:       "promise",
// 				ProposalId: msg.ProposalId,
// 				AcceptId:   px.proposalId,
// 				Val:        px.acceptedVal,
// 				From:       px.me,
// 				To:         msg.To,
// 			}
// 			px.net.recvQueue <- promiseMessage
// 		}
// 	case "accept": // phase 2
// 		if msg.ProposalId >= px.proposalId {
// 			px.proposalId = msg.ProposalId
// 			px.acceptedVal = msg.Val

// 			acceptedMsg := Message{
// 				Type:       "accepted",
// 				From:       px.me,
// 				ProposalId: msg.ProposalId,
// 				AcceptId:   msg.ProposalId,
// 				Val:        msg.Val,
// 				To:         msg.To,
// 			}
// 			px.net.recvQueue <- acceptedMsg
// 		}

// 	}

// }

// // function for determining if we reach majority
// func (px *Paxos) MajorityReach() bool {
// 	return true
// }

// functions for proposer/leader
// TODO / NOTE: only call prepare when starting election

// // TODO / NOTE: only the leader should be calling this
// func (px *Paxos) Propose(s string) {

// 	if px.state != "L" {
// 		return
// 	}

// 	px.proposalId += 1
// 	// message type is accept bc we want the acceptors
// 	// to accept
// 	msg := Message{
// 		Type:       "accept",
// 		ProposalId: px.proposalId,
// 		From:       px.me,
// 		Val:        s,
// 	}
// 	px.net.recvQueue <- msg
// }

// function for majority threshold
func quorum(state *GlobalState) int {
	return len(state.HeartBeatState.AllAlivePeerAddresses(state.Config))/2 + 1
}

//function for clean the accepted value
func (px *Paxos) clean() {
	px.proposalId = 0
	px.proposalAccepted = false
	px.acceptedMessages = nil
	px.acceptedVal = nil
}
