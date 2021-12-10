package paxos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

type LockMessage struct {
	Lock bool
}

type LockRelayMessage struct {
	Lock           bool
	OriginServerId int
	ClientAddr     *net.UDPAddr
}

func UDPServeLockMessage(selfId int, conn *net.UDPConn, ch chan GenericMessage) {
	buf := make([]byte, 1024)
	go func() {
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			newbuf := append(make([]byte, 0), buf[:n]...)
			if err != nil {
				log.Warn(fmt.Sprintf("Error read from UDP: " + err.Error()))
				continue
			}
			log.Info(fmt.Sprintf("Received %d bytes from %v", n, addr))
			var msg = LockMessage{}
			if err := json.Unmarshal(newbuf, &msg); err != nil {
				log.Warn(string(newbuf))
				log.Warn(fmt.Sprintf("Error decoding message: " + err.Error()))
			}
			ch <- GenericMessage{LockRelay: &LockRelayMessage{msg.Lock, selfId, addr}}
		}
	}()
}
