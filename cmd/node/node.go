package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/atgao/paxos"
)

// var state *paxos.GlobalState

func main() {
	configPath := flag.String("config", "", "Config file path")
	flag.Parse()
	if *configPath == "" {
		panic("Config file path is empty")
	}

	config, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	state, err := paxos.GlobalInitialize([]byte(config))
	time.Sleep(4 * time.Second)

	// TestNewPaxosNode(state)
	TestSingleProposer(state)
	select {}
}

func TestNewPaxosNode(state *paxos.GlobalState) {

	var px []*paxos.Paxos
	for i := 0; i < 3; i++ {
		px = append(px, paxos.Make(i, state))
	}
	px[0].Prepare(state)
}

func TestSingleProposer(state *paxos.GlobalState) {
	var px []*paxos.Paxos
	for i := 0; i < 3; i++ {
		px = append(px, paxos.Make(i, state))
	}
	px[0].Prepare(state)
	time.Sleep(2 * time.Second)

	px[0].Propose("100")
	time.Sleep(2 * time.Second)

	for i := 0; i < 3; i++ {
		if px[i].ReturnValue() != "100" {
			fmt.Printf("wrong decided value, want value:%v decided value:%v", 100, px[i].ReturnValue())
		}
	}

}

// func TestLearner(state *paxos.GlobalState) {
// 	px1 := paxos.Make(1, state)
// 	prepareMessage := px1.Prepare(state)
// 	paxos.BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), prepareMessage)
// 	px1.Propose(100)
// 	time.Sleep(1 * time.Second)
// }
