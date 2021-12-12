
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
// func (px *Paxos) run(state *GlobalState) {
// 	// If no consensus were reached
// 	for {
// 		NewPaxosMessage := <-state.PaxosMessageQueue
// 		log.Info(fmt.Sprintf("Received paxos message to run: %v", NewPaxosMessage))
// 		switch NewPaxosMessage.Type {
// 		case "prepare": // proposer --> acceptor
// 			log.Info(fmt.Sprintf("Proposer start run... val:", NewPaxosMessage.ProposalId))
// 			px.runAcceptor(NewPaxosMessage)

// 			//Proposor send prepare message to acceptor to reach majority consensus.
// 		case "promise":
// 			px.runPromisor(NewPaxosMessage)

// 		case "accept": // proposer --> acceptor
// 			log.Info(fmt.Sprintf("acceptor start run... val:", NewPaxosMessage.CurrentVal))

// 			px.runAcceptor(NewPaxosMessage)

// 		case "accepted":
// 			// px.runLearner(NewPaxosMessage)
// 		}

// 	}

// 	// }
// }
// func (px *Paxos) runPromisor(msg Message) {
// 	if msg.ProposalId > px.minProposal {
// 		px.minProposal = msg.ProposalId
// 		log.Info(fmt.Sprintf("promise received\n"))
// 	}

// }

// func (px *Paxos) ReturnValue() interface{} {
// 	return px.acceptedVal
// }

// func (px *Paxos) Prepare(state *GlobalState) {

// 	//creating an array of messages, each message have different To target
// 	//right now we just send the message to every acceptor (m=3))
// 	log.Info(fmt.Sprintf("Prepare is called...\n"))
// 	msg := Message{
// 		Type:       "prepare",
// 		ProposalId: px.minProposal + 1,
// 		CurrentVal: 10,
// 	}
// 	BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), msg)

// 	// send to network so can send to others
// }

// // // function for learner
// // func (px *Paxos) runLearner(msg Message) string {

// // 	for {
// // 		if msg.Val == "" {
// // 			continue
// // 		}
// // 		if msg.ProposalId > px.proposalId {
// // 			px.proposalId = msg.ProposalId
// // 		}
// // 		learnMsg, islearn := px.choseMajority()
// // 		if islearn == false {
// // 			continue
// // 		}
// // 		px.acceptedVal = learnMsg.Val
// // 		return learnMsg.Val
// // 	}
// // }

// // func (px *Paxos) choseMajority() (Message, bool) {
// // 	//need to loop through all accepted message
// // 	CountResult := make(map[int]int)
// // 	MessageResult := make(map[int]Message)

// // 	for _, Msg := range px.acceptedMessages {
// // 		ProposalID := Msg.ProposalId
// // 		CountResult[ProposalID] += 1
// // 		MessageResult[ProposalID] = Msg
// // 	}

// // 	for ChosenID, ChosenMsg := range MessageResult {
// // 		fmt.Printf("Proposal[%v] Message[%s]\n", ChosenID, ChosenMsg.Val)
// // 		if CountResult[ChosenID] > px.quorum() {
// // 			return ChosenMsg, true
// // 		}
// // 	}
// // 	return Message{}, false
// // }

// // function for acceptor
// func (px *Paxos) runAcceptor(msg Message) {

// 	switch msg.Type {
// 	case "prepare": // phase 1
// 		log.Info(fmt.Sprintf("[proposer:%d] phase 1, prepareMsg:%v", px.minProposal, msg.ProposalId))

// 		if msg.ProposalId > px.minProposal {
// 			px.minProposal = msg.ProposalId
// 			promiseMessage := Message{
// 				Type:       "promise",
// 				ProposalId: msg.ProposalId,
// 				AcceptId:   px.minProposal,
// 				CurrentVal: px.acceptedVal,
// 			}
// 			log.Info(fmt.Sprintf("promise returned"))
// 			BroadcastPaxosMessage(px.globalstate.InterNodeUDPSock, px.globalstate.Config.AllPeerAddresses(), promiseMessage)
// 		}
// 		//otherwise reject

// 	case "accept": // phase 2
// 		log.Info(fmt.Sprintf("[proposer:%d] phase 2, acceptMsg current val:%v", px.me, msg.CurrentVal))

// 		if msg.ProposalId >= px.minProposal {
// 			px.minProposal = msg.ProposalId
// 			px.acceptedVal = msg.CurrentVal
// 			log.Info(fmt.Sprintf("[proposer:%d] Value accepted, current val:%v", px.me, px.ReturnValue()))

// 			acceptedMsg := Message{
// 				Type:       "accepted",
// 				ProposalId: msg.ProposalId,
// 				AcceptId:   msg.ProposalId,
// 				CurrentVal: msg.CurrentVal,
// 			}
// 			BroadcastPaxosMessage(px.globalstate.InterNodeUDPSock, px.globalstate.Config.AllPeerAddresses(), acceptedMsg)
// 		}
// 	}
// }

// // // function for determining if we reach majority
// // func (px *Paxos) MajorityReach() bool {
// // 	return true
// // }

// // functions for proposer/leader
// // TODO / NOTE: only call prepare when starting election

// // TODO / NOTE: only the leader should be calling this
// func (px *Paxos) Propose(s int) {

// 	// if px.role != "L" {
// 	// 	return
// 	// }

// 	px.minProposal += 1
// 	// message type is accept bc we want the acceptors
// 	// to accept
// 	msg := Message{
// 		Type:       "accept",
// 		ProposalId: px.minProposal,
// 		CurrentVal: s,
// 	}
// 	log.Info(fmt.Sprintf("accept sended %+v\n", msg))
// 	BroadcastPaxosMessage(px.globalstate.InterNodeUDPSock, px.globalstate.Config.AllPeerAddresses(), msg)

// }

// // function for majority threshold
// func quorum(state *GlobalState) int {
// 	return len(state.HeartBeatState.AllAlivePeerAddresses(state.Config))/2 + 1
// }

// //function for clean the accepted value
// func (px *Paxos) clean() {
// 	// px.proposalId = 0
// 	// px.proposalAccepted = false
// 	// px.acceptedMessages = nil
// 	// px.acceptedVal = nil
// }
