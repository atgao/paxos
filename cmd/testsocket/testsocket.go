package main

import (
	"github.com/atgao/paxos"
)

func main() {
	paxos.BroadcastPaxosMessage(
		[]string{"127.0.0.1:1027"}, paxos.Message{
			Type:       "ty",
			ProposalId: 0,
			AcceptId:   0,
			Val:        0,
			From:       0,
		})
}
