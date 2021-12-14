package paxos

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

const LockWaitDuration = 1000 * time.Second

func RequestLockServer(server string, lock bool, lockId int, clientID string, msgUUID uuid.UUID) LockResult {
	msg := LockMessage{lock, lockId, clientID, msgUUID}
	buffer, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Failed to encode lock message")
	}
	readBuffer := make([]byte, 1024)

	for {
		clientSock, err := net.ListenUDP("udp", nil)
		if err != nil {
			log.Fatal("Failed to resolve the server")
		}

		log.Info(fmt.Sprintf("Sending lock message %+v to %s", msg, server))
		addr, err := net.ResolveUDPAddr("udp", server)
		if err != nil {
			return FAIL
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
				ch <- LockReplyMessage{Result: FAIL}
				return
			}
			var b LockReplyMessage
			if err := json.Unmarshal(readBufferResized, &b); err != nil {
				log.Warn("Bad response from server " + err.Error())
				ch <- LockReplyMessage{Result: FAIL}
			} else {
				ch <- b
			}
		}()

		select {
		case <-time.After(LockWaitDuration):
			log.Warn("Lock timeout")
			return TIMEOUT
		case r := <-ch:
			switch r.Result {
			case RETRY:
				server = *r.RetryAddr
			default:
				return r.Result
			}
		}
	}
}
