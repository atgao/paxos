package paxos

import "net"

type LockMessage struct {
	Lock bool
}

type LockRelayMessage struct {
	Lock       bool
	ClientAddr net.TCPAddr
}
