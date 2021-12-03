package paxos

import (
	"bytes"
	"encoding/gob"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

type GenericMessage struct {
	Paxos     *Message
	KeepAlive *KeepAliveMessage
}

func sendOne(conn *net.UDPConn, address string, buffer bytes.Buffer) error {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	/*
		conn, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			return err
		}
		defer conn.Close()
	*/

	_, err = conn.WriteTo(buffer.Bytes(), raddr)
	if err != nil {
		return err
	}
	return nil
}

func broadcast(conn *net.UDPConn, addresses []string, buffer bytes.Buffer) {
	for _, address := range addresses {
		if err := sendOne(conn, address, buffer); err != nil {
			log.Fatal("Failed to send msaage to %s: err: %s", address, err)
		}
	}
}

func sendGenericMessage(conn *net.UDPConn, address string, msg GenericMessage) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(msg); err != nil {
		log.Fatal("Failed to encode messsage: " + err.Error())
	}
	return sendOne(conn, address, buffer)
}

func broadcastGenericMsg(conn *net.UDPConn, addresses []string, msg GenericMessage) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(msg); err != nil {
		log.Fatal("Failed to encode messsage: " + err.Error())
	}
	broadcast(conn, addresses, buffer)
}

func BroadcastPaxosMessage(conn *net.UDPConn, addresses []string, msg Message) {
	log.Info("Broadcasting paxos message")
	broadcastGenericMsg(conn, addresses, GenericMessage{&msg, nil})
}

func BroadcastKeepAliveMessage(conn *net.UDPConn, addresses []string, msg KeepAliveMessage) {
	log.Info("Broadcasting keep alive message")
	broadcastGenericMsg(conn, addresses, GenericMessage{nil, &msg})
}

func MkUDPMessageQueue(conn *net.UDPConn) (chan GenericMessage, error) {
	ret := make(chan GenericMessage)
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
			var msg = GenericMessage{}
			err = dec.Decode(&msg)
			if err != nil {
				log.Warn(fmt.Sprintf("Error decoding message: " + err.Error()))
			}
			ret <- msg
		}
	}()
	return ret, nil
}
