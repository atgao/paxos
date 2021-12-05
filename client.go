package paxos

import (
	"bytes"
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

const LockWaitDuration = 10 * time.Second

func RequestLockServer(server string, lock bool) bool {
	addr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return false
	}
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
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(LockMessage{lock}); err != nil {
		log.Fatal("Failed to encode lock message")
	}
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		log.Fatal("Failed to send lock message")
	}

	ch := make(chan bool)
	go func() {
		_, err := conn.Read(buffer.Bytes())
		if err != nil {
			log.Warn("Failed to read server response: " + err.Error())
			return
		}
		dec := gob.NewDecoder(&buffer)
		var b bool
		if err := dec.Decode(&b); err != nil {
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
