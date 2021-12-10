package main

import (
	"flag"
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

	t := time.NewTimer(8 * time.Second)
	select {
	case <-t.C:
	}

	px1 := paxos.Make(1, state)
	// px2 := paxos.Make(2,state)
	prepareMessage := px1.Prepare()
	paxos.BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), prepareMessage)

	select {}
}
