package paxos 

import "fmt" // for testing

// 
// struct for messages sent between nodes
// 
type Message struct {
	Type 			string // prepare, propose, accept, etc 
	ProposalId 		int    // id proposed
	AcceptId		int    // id accepted
	Val 			int    // value proposed or accepted or promised
	From 			int    // index of the sending node
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

	// channels to always read from 
	ch 				chan Message // channel to communicate from node --> network
	acceptor 		chan Message // channel to read from as acceptor 
	learner  		chan Message // channel as learner


	state 			string // leader (proposer) or acceptor
	Log 			[]LogEntry // TODO: fix this... 
}

// function for learner
func (px *Paxos) runLearner() {

}

// function for acceptor
func (px *Paxos) runAcceptor(){

}

// functions for proposer/leader 
// TODO / NOTE: only call prepare when starting election
func (px *Paxos) Prepare() {
	px.proposalId += 1
	msg := Message {
		Type: "prepare", 
		ProposalId: px.proposalId, 
		From: px.me,
	}

	// send to network so can send to others
	px.net.recvQueue <- msg

}

//
// kills this paxos node
//
func (px *Paxos) kill() {
	// TODO:
}

// 
// make new paxos node
// 
func Make(me int, net *Network) *Paxos {
	px := &Paxos{}
	px.me = me 
	px.net = net
	px.ch = px.net.sendQueue[me] // SHOULD ALWAYS BE THIS INDEX... 

	// init the paxos node 
	px.state = "A" // for acceptor 
	px.Log = make([]LogEntry, 0) // TODO: is this needed...
	px.proposalId = -1 // this way when we start its 0


	go px.run()

	return px
}

// this go routine keeps on running in the background
func (px *Paxos) run() {

	// go px.runAcceptor()
	// go px.runLearner()

	// loop to listen to messages and 
	// forward them along to proper channels
	for {
		// select {
		msg := <- px.ch
			fmt.Printf("from %v to %v, %v\n", msg.From, px.me, msg)

		// }
	}

}



// TODO: function for later to restart election
// for when the current leader dies
func (px *Paxos) startElection() {

}


