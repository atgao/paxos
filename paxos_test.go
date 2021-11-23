package paxos 

import (
	"fmt"
	"testing"
)

func TestNewPaxosNode(t *testing.T) {
	net := MakeNetwork(1)
	px := Make(0, net)
	fmt.Printf("Paxos Node %v \n", px.me)
}

func TestPrepare(t *testing.T) {
	fmt.Printf("Running Prepare test...\n")
	n := 3
	net := MakeNetwork(n)
	var px []*Paxos

	for i := 0; i < n; i++ {
		px = append(px, Make(i, net))
	}

	px[0].Prepare()

}

func TestProposer(t *testing.T) {
	fmt.Printf("Running Propose test....\n")
	n :=3 
	net := MakeNetwork(n)
	var px []*Paxos 

	for i := 0; i < n; i++ {
		px = append(px, Make(i, net))
	}

	// TODO: note that we're just manually setting
	// node 0 to be leader for the test, should be changed later on

	px[0].state = "L"
	px[0].Propose(0)
}
