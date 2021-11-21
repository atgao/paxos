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
	acceptedVal 	 int  // ?? 

	// channels to always read from 
	ch 				chan Message // channel to communicate from node --> network ????
	acceptor 		chan Message // channel to read from as acceptor 
	learner  		chan Message // channel as learner
	proposer 		chan Message // channel as proposer/leader


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
func (px *Paxos) runAcceptor(){
	
	for {
		msg := <- px.acceptor

		switch msg.Type {
		case "prepare":
			if msg.ProposalId > px.proposalId {
				px.proposalId = msg.ProposalId // update proposal 

				// TODO: update proposal accepted boolean?? 
				// TODO: check if the fields are right too...
				promiseMessage := Message {
					Type: "promise", 
					ProposalId: msg.ProposalId, 
					AcceptId: px.proposalId, 
					Val: px.acceptedVal, 
				}

				px.net.recvQueue <- promiseMessage
			}


		// case "accept": 
		// 	if msg.ProposalId >= px.ProposalId {
		// 		// TODO: some stuff 
		// 	}


		}
	}

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

func (px *Paxos) runProposer() {
	// var npromised, q int = 0, quorum(px.net.nodes)
	// for {
	// 	msg := <- px.proposer 
	// 	switch msg.Type {
	// 	// case "propose": 
	// 	// 	proposeMsg := {

	// 	// 	}
	// 	case "promise":
	// 		if msg.ProposalId > px.proposalId


	// 	}
	// }

}


// TODO: need to figure this out...
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
	px.me = me // my index in the sendQueue array 
	px.net = net
	px.ch = px.net.sendQueue[me] 

	// init the paxos node 
	px.state = "A" // for acceptor ??? may need to fix...
	px.Log = make([]LogEntry, 0) // TODO: is this needed...
	px.proposalId = -1 // this way when we start its 0
	px.proposalAccepted = false 
	px.acceptedVal = -1 
	px.acceptor = make(chan Message)
	px.learner = make(chan Message)
	px.proposer = make(chan Message)


	go px.run()

	return px
}

// this go routine keeps on running in the background
func (px *Paxos) run() {

	go px.runProposer()
	go px.runAcceptor()
	// go px.runLearner()

	// loop to listen to messages and 
	// forward them along to proper channels
	for {
		// select {
		msg := <- px.ch
			fmt.Printf("from %v to %v, %v\n", msg.From, px.me, msg)

		// }

		switch msg.Type {
		case "prepare": // proposer --> acceptor 
			px.acceptor <- msg 
		case "accept": // proposer --> acceptor 
			px.acceptor <- msg 
		
		// // do some stuff here for proposer
		case "promise": 
			px.proposer <- msg
		// case "propose": I DONT THINK THIS ACTUALLY GOES HERE
		// 	px.proposer <- msg 

		case "accepted":
			px.learner <- msg			
		}
		
	}

}



// TODO: function for later to restart election
// for when the current leader dies
func (px *Paxos) startElection() {

}


