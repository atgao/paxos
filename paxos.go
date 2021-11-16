package paxos 

import network

// 
// struct for messages sent between nodes
// 
type Message struct {
	type 			string // prepare, propose, accept, etc 
	proposalId 		int    // id proposed
	acceptId		int    // id accepted
	val 			int    // value proposed or accepted or promised
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
type Paxos {
	// add clients or?? 
	peers 	[]*network.ClientEnd 
	me 		int // id in the peers array

	proposalId int // 
	maxId 	   int // max id seen so far
	proposalAccepted bool // already accepted proposal ?? 

	// TODO: create recv channel always reading from..


	state 	string // leader or acceptor 
	Log 	[]LogEntry // logs for paxos 
}

// function for learner
func (px *Paxos) runLearner() {

}

// function for acceptor
func (px *Paxos) runAcceptor(){
	switch msg.type {

	case msg.type == "propose":
		// TODO: check if ID is largest so far 
		// note that proposal was accepted
		// save proposal number, save proposal data
	
		if msg.proposalId == px.proposalId {
			accpetedMsg := Message {
				type: "accepted", 
				acceptId: msg.proposalId, 
				val: msg.val, 
			}
		}

		// TODO: Send accepted message out to proposer 
		// and all learners

	case msg.type == "prepare": 
		if msg.proposalId > px.proposalId {
			px.proposalId = msg.proposalId 

			promiseMsg := Message {
				type: "promise",
			}

			// TODO: configure to send promise back to sender
		}


	}

}

// functions for proposer/leader 
// TODO: should only be called by leader??
func (px *Paxos) Propose(proposalId int, val int){
	msg := Message {
		type: 	"propose", 
		proposalId: proposalId, 
		val: val 
	}

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
func Make(peers []*network.ClientEnd, me int) *Paxos {
	px := &Paxos{}
	px.me = me 
	px.peers = peers

	// init the paxos node 
	px.state = "A" // for acceptor 


	return px
}


// TODO: will only be called on first node
func (px *Paxos) startElection() {

}


// this go routine keeps on running in the background
func (px *Paxos) run() {

	go px.runAcceptor()

}