package paxos

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type GenericMessage struct {
	Paxos     *Message
	KeepAlive *KeepAliveMessage
	LockRelay *LockRelayMessage
}

func sendOneAddr(conn *net.UDPConn, address *net.UDPAddr, buffer bytes.Buffer) error {
	_, err := conn.WriteTo(buffer.Bytes(), address)
	if err != nil {
		return err
	}
	return nil
}

func sendOne(conn *net.UDPConn, address string, buffer bytes.Buffer) error {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	return sendOneAddr(conn, raddr, buffer)
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
	gmsg := GenericMessage{&msg, nil, nil}
	log.Warn(fmt.Sprintf("paxos msg %+v", *gmsg.Paxos))
	broadcastGenericMsg(conn, addresses, gmsg)
}

func BroadcastKeepAliveMessage(conn *net.UDPConn, addresses []string, msg KeepAliveMessage) {
	log.Info("Broadcasting keep alive message")
	broadcastGenericMsg(conn, addresses, GenericMessage{nil, &msg, nil})
}

func SendLockRelayMessage(conn *net.UDPConn, address string, msg LockRelayMessage) {
	log.Info("Sending lock relay message")
	sendGenericMessage(conn, address, GenericMessage{LockRelay: &msg})
}

func UDPServeGenericMessage(conn *net.UDPConn, ch chan GenericMessage) {
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
			ch <- msg
		}
	}()
}
