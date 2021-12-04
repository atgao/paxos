package paxos

import (
	"encoding/json"
	"sync"
)

type Config struct {
	mu          sync.Mutex
	SelfId      int
	PaxosAddr   string
	ServerAddr  string
	PeerAddress map[int]string
}

func MkConfig(selfId int, paxosAddr string, serverAddr string, peerAddress map[int]string) *Config {
	config := &Config{}
	config.SelfId = selfId
	config.PaxosAddr = paxosAddr
	config.ServerAddr = serverAddr
	config.PeerAddress = peerAddress
	return config
}

func ConfigFromJSON(data []byte) *Config {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		panic(err)
	}
	selfId := int(result["id"].(float64))
	addresses := result["address"].(map[string]interface{})
	paxosAddr := addresses["paxos"].(string)
	serverAddr := addresses["server"].(string)
	// var peerAddress map[int]string
	peers := result["peers"].([]interface{})
	peerAddress := make(map[int]string)
	for _, p := range peers {
		p1 := p.(map[string]interface{})
		peerAddress[int(p1["id"].(float64))] = p1["address"].(string)
	}

	return MkConfig(selfId, paxosAddr, serverAddr, peerAddress)
}

func (config *Config) AllPeerAddresses() []string {
	addresses := make([]string, len(config.PeerAddress))
	idx := 0
	for _, addr := range config.PeerAddress {
		addresses[idx] = addr
		idx++
	}
	return addresses
}
