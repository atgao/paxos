package paxos

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

const LockWaitDuration = 10 * time.Second

func RequestLockServer(clientSock *net.UDPConn, server string, lock bool, UUID uuid.UUID) bool {
	msg := LockMessage{lock, UUID}
	buffer, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Failed to encode lock message")
	}
	readBuffer := make([]byte, 1024)

	for {
		log.Info(fmt.Sprintf("Sending lock message %+v to %s", msg, server))
		addr, err := net.ResolveUDPAddr("udp", server)
		if err != nil {
			return false
		}

		if _, err := clientSock.WriteTo(buffer, addr); err != nil {
			log.Fatal("Failed to send lock message")
		}

		ch := make(chan LockReplyMessage)
		go func() {
			n, _, err := clientSock.ReadFromUDP(readBuffer)
			readBufferResized := append(make([]byte, 0), readBuffer[:n]...)
			if err != nil {
				log.Warn("Failed to read server response: " + err.Error())
				return
			}
			var b LockReplyMessage
			if err := json.Unmarshal(readBufferResized, &b); err != nil {
				log.Warn("Bad response from server " + err.Error())
				ch <- LockReplyMessage{Success: false}
			} else {
				ch <- b
			}
		}()

		select {
		case <-time.After(LockWaitDuration):
			log.Warn("Lock timeout")
			return false
		case r := <-ch:
			if r.Success {
				return true
			} else if r.RetryAddr != nil {
				server = *r.RetryAddr
			} else {
				return false
			}
		}
	}
}
