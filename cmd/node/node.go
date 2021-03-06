package main

import (
	"flag"
	"github.com/atgao/paxos"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

// var state *paxos.GlobalState

func main() {
	// log.SetReportCaller(true)
	configPath := flag.String("config", "", "Config file path")
	verbose := flag.Int("verbose", 0, "Verbose output")
	flag.Parse()
	if *configPath == "" {
		panic("Config file path is empty")
	}
	if *verbose == 1 {
		log.SetLevel(log.DebugLevel)
	} else if *verbose >= 2 {
		log.SetLevel(log.TraceLevel)
	}

	config, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	paxos.GlobalInitialize([]byte(config))
	// time.Sleep(4 * time.Second)

	/*
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

	*/

	select {}
}

// func TestLearner(state *paxos.GlobalState) {
// 	px1 := paxos.Make(1, state)
// 	prepareMessage := px1.Prepare(state)
// 	paxos.BroadcastPaxosMessage(state.InterNodeUDPSock, state.Config.AllPeerAddresses(), prepareMessage)
// 	px1.Propose(100)
// 	time.Sleep(1 * time.Second)
// }
