package main

import (
	"flag"
	"fmt"
	"github.com/atgao/paxos"
	"io/ioutil"
)

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

	t := paxos.ConfigFromJSON([]byte(`{"id": 1, "peers": [{"id": 2, "address": "a"}]}`))
	fmt.Printf("%v\n", t)

	t = paxos.ConfigFromJSON([]byte(config))
	fmt.Printf("%v\n", t)
}
