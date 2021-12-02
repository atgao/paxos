package paxos 

import (
	"fmt"
	) // for testing

// 
// struct for messages sent between nodes
// 
type Message struct {
	Type 			string // prepare, propose, accept, etc 
	ProposalId 		int    // id proposed
	AcceptId		int    // id accepted
	Val 			string    // value proposed or accepted or promised 
	From 			int    // index of the sending node
	To				int 	// index of the receiving node
}

//
// struct for log entries ??
//
type LogEntry struct {
	term 	int // term this long belongs to
	command interface{} // command to be executed
}

//
// Go object implementing single Paxos node 
//
type Paxos struct {
	me 			int // id in the peers array
	net 		*Network // network node belongs to

	// proposalId is also the highest one seen so far
	proposalId 		 int  // NOTE THAT THIS SHOULD ALWAYS BE INCREASING
	proposalAccepted bool // already accepted proposal ?? 
	acceptedVal 	 string  // I think this

	// channels to always read from 
	ch 				chan Message // channel to communicate from node --> network ????
	state 			string // leader (proposer) or acceptor
	Log 			[]LogEntry // TODO: fix this... 
}

func quorum(n int) int{
	return n/2 +1
}

// function for learner
func (px *Paxos) runLearner() {
}

// function for acceptor
func (px *Paxos) runAcceptor(msg Message){
	
		switch msg.Type {
			case "prepare": // phase 1
				if msg.ProposalId > px.proposalId {
					px.proposalId = msg.ProposalId // update proposal 

					// TODO: update proposal accepted boolean?? 
					// TODO: check if the fields are right too...
					// TODO: may need to fix the accept id

					promiseMessage := Message {
						Type: "promise", 
						ProposalId: msg.ProposalId, 
						AcceptId: px.proposalId, 
						Val: px.acceptedVal, 
						From: px.me, 
						To: 
					}
					px.net.recvQueue <- promiseMessage
				}
			case "accept": // phase 2
				if msg.ProposalId >= px.proposalId {
					px.proposalId = msg.ProposalId
					px.acceptedVal = msg.Val 

					acceptedMsg := Message{
						Type: "accepted", 
						From: px.me, 
						ProposalId: msg.ProposalId, 
						AcceptId: msg.ProposalId, 
						Val: msg.Val,
						To:
					}
					px.net.recvQueue <- acceptedMsg
				}

		}


}

// functions for proposer/leader 
// TODO / NOTE: only call prepare when starting election
func (px *Paxos) Prepare() {

	//creating an array of messages, each message have different To target

	//right now we just send the message to every acceptor (which is every Px Node) 
	for acepId, _ := range px {
		px.proposalId += 1
		msg := Message {
			Type: "prepare", 
			ProposalId: px.proposalId, 
			From: px.me,
			To: 
		}
	// send to network so can send to others


	px.net.recvQueue <- msg
}

// TODO / NOTE: only the leader should be calling this
func (px *Paxos) Propose(val string) {
	if px.state != "L" {
		return 
	}

	px.proposalId += 1
	// message type is accept bc we want the acceptors 
	// to accept 
	msg := Message {
		Type: "accept", 
		ProposalId: px.proposalId, 
		From: px.me,
		Val: val, 
	}

	px.net.recvQueue <- msg
}


func (px *Paxos) kill() {
	// TODO:
}

// 
// make new paxos node
// 
func Make(me int, net *Network) *Paxos {
	px := &Paxos{}
	px.me = me // my index in the sendQueue array 
	px.net = net
	px.ch = px.net.sendQueue[me] 

	// init the paxos node 
	px.state = "A" // for acceptor ??? may need to fix...
	px.Log = make([]LogEntry, 0) // TODO: is this needed...
	px.proposalId = -1 // this way when we start its 0
	px.proposalAccepted = false 
	px.acceptedVal = "Test"

	go px.run()
	return px
}

// this go routine keeps on running in the background
func (px *Paxos) run() {

	// loop to listen to messages and 
	// forward them along to proper channels
	for {
		msg := <- px.ch
			fmt.Printf("from %v to %v, %v\n", msg.From, px.me, msg)

		switch msg.Type {
		case "prepare": // proposer --> acceptor 

			fmt.Println("Proposer start run... val:", msg.ProposalId)
			//Proposor send prepare message to acceptor to reach majority consensus.
			
			//need a condition to check if majority reached



			
		case "accept": // proposer --> acceptor 
			px.runAcceptor(msg)
		
		// // do some stuff here for proposer
		case "promise": 

		// case "propose": I DONT THINK THIS ACTUALLY GOES HERE
		// 	px.proposer <- msg 

		case "accepted":
		}
		
	}

}

// TODO: function for later to restart election
// for when the current leader dies
func (px *Paxos) startElection() {

}


