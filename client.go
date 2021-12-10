package paxos

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

const LockWaitDuration = 10 * time.Second

func RequestLockServer(clientSock *net.UDPConn, server string, lock bool) bool {
	addr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return false
	}
	/*
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return false
		}
		defer func() {
			err = conn.Close()
			if err != nil {
				return
			}
		}()
	*/
	buffer, err := json.Marshal(LockMessage{lock})
	if err != nil {
		log.Fatal("Failed to encode lock message")
	}

	if _, err := clientSock.WriteTo(buffer, addr); err != nil {
		log.Fatal("Failed to send lock message")
	}

	readBuffer := make([]byte, 1024)
	ch := make(chan bool)
	go func() {
		n, _, err := clientSock.ReadFromUDP(readBuffer)
		readBufferResized := append(make([]byte, 0), readBuffer[:n]...)
		if err != nil {
			log.Warn("Failed to read server response: " + err.Error())
			return
		}
		var b bool
		if err := json.Unmarshal(readBufferResized, &b); err != nil {
			log.Warn("Bad response from server " + err.Error())
			ch <- false
		} else {
			ch <- b
		}
	}()

	select {
	case <-time.After(LockWaitDuration):
		log.Warn("Lock timeout")
		return false
	case r := <-ch:
		return r
	}
}
