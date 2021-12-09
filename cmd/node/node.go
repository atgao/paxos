package main

import (
	"flag"
	"io/ioutil"

	"github.com/atgao/paxos"
)

var state *paxos.GlobalState

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

	paxos.GlobalInitialize([]byte(config))

	select {}
}
