package paxos

import (
	"bytes"
	"encoding/gob"
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
			if err != nil {
				log.Warn(fmt.Sprintf("Error read from UDP: " + err.Error()))
				continue
			}
			log.Info(fmt.Sprintf("Received %d bytes from %v", n, addr))
			r := bytes.NewBuffer(buf)
			dec := gob.NewDecoder(r)
			var msg = LockMessage{}
			err = dec.Decode(&msg)
			if err != nil {
				log.Warn(fmt.Sprintf("Error decoding message: " + err.Error()))
			}
			ch <- GenericMessage{LockRelay: &LockRelayMessage{msg.Lock, selfId, addr}}
		}
	}()
}
