package paxos

// import "encoding/gob"
// import "bytes"
// import "reflect"
// import "sync"
// import "log"
// import "strings"
// import "math/rand"
// import "time"
// import "fmt"


type Network struct {
	nodes 	  int // number of Paxos nodes
	sendQueue map[int] chan Message
	recvQueue chan Message
}

func (net *Network) recv() {
	for {
		msg := <- net.recvQueue
		for i, sendQueue := range net.sendQueue {
			
			if i == msg.From {
				continue
			}
			// fmt.Printf("sending to %v\n", i)
			// go func(sendQueue chan Message, msg Message) {
			// 	sendQueue <- msg
			// }(sendQueue, msg)

			// TODO: need to make sure this always sends...
			sendQueue <- msg
			
		} 
	}

}

func MakeNetwork(nodes int) *Network {
	net := &Network{}
	net.nodes = nodes

	net.sendQueue = make(map[int] chan Message, 0)
	net.recvQueue = make(chan Message)

	//TODO: fill out other fields here...
	for i := 0; i < nodes; i++ {
		net.sendQueue[i] = make(chan Message, 1) // arbitrary to avoid blocking
	}

	go net.recv()

	return net
}

