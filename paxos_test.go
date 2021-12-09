package paxos

import (
	"fmt"
	"testing"
)

func TestNewPaxosNode(t *testing.T) {

	px := Make(1)
	fmt.Printf("Paxos Node %v \n", px.me)
}

func TestPrepare(t *testing.T) {
	fmt.Printf("Running Prepare test...\n")
	var px []*Paxos
	for i := 0; i < 3; i++ {
		px = append(px, Make(i))
	}
	//Test Prepare
	px[0].Prepare()

}

// func TestProposer(t *testing.T) {
// 	fmt.Printf("Running Propose test....\n")
// 	n := 3
// 	net := MakeNetwork(n)
// 	var px []*Paxos

// 	for i := 0; i < n; i++ {
// 		px = append(px, Make(i, net, n))
// 	}

// 	// TODO: note that we're just manually setting
// 	// node 0 to be leader for the test, should be changed later on

// 	px[0].state = "L"

// 	//leave it outside for now | for testing
// 	px[0].Prepare()
// 	px[0].Propose("cat")

// }

// func TestLearner(t *testing.T) {
// 	fmt.Printf("Running Learner test....\n")
// 	n := 3
// 	net := MakeNetwork(n)
// 	var px []*Paxos

// 	for i := 0; i < n; i++ {
// 		px = append(px, Make(i, net, n))
// 	}

// 	// TODO: note that we're just manually setting
// 	// node 0 to be leader for the test, should be changed later on

// 	px[0].state = "L"

// 	//leave it outside for now | for testing
// 	px[0].Prepare()
// 	px[0].Propose("cat")
// 	time.Sleep(1 * time.Second)

// 	if px[0].acceptedVal != "cat" {
// 		t.Errorf("Learner learn wrong proposal.")

// 	}
// }
