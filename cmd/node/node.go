package main

import (
	"bufio"
	"flag"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/atgao/paxos"
)

// var state *paxos.GlobalState

func main() {
	log.SetReportCaller(true)
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

	if state.Config.SelfId == 3 || state.Config.SelfId == 4 {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		state.ProposerAlgorithm(paxos.ProposalValue{
			Lock: true,
			Addr: net.UDPAddr{
				Port: state.Config.SelfId,
			},
			UUID: uuid.New(),
		})

		state.ProposerAlgorithm(paxos.ProposalValue{
			Lock: true,
			Addr: net.UDPAddr{
				Port: state.Config.SelfId,
			},
			UUID: uuid.New(),
		})
	}

	select {}
}

// func TestLearner(state *paxos.GlobalState) {
// 	px1 := paxos.Make(1, state)
// 	prepareMessage := px1.Prepare(state)
// 	paxos.BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), prepareMessage)
// 	px1.Propose(100)
// 	time.Sleep(1 * time.Second)
// }
