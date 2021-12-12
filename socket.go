package paxos

import (
	"encoding/json"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type GenericMessage struct {
	Paxos     *Message
	KeepAlive *KeepAliveMessage
	LockRelay *LockRelayMessage
}

func sendOneAddr(conn *net.UDPConn, address *net.UDPAddr, buffer []byte) error {
	_, err := conn.WriteTo(buffer, address)
	if err != nil {
		return err
	}
	return nil
}

func sendOne(conn *net.UDPConn, address string, buffer []byte) error {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	return sendOneAddr(conn, raddr, buffer)
}

func broadcast(conn *net.UDPConn, addresses []string, buffer []byte) {
	for _, address := range addresses {
		if err := sendOne(conn, address, buffer); err != nil {
			log.Fatal("Failed to send msaage to %s: err: %s", address, err)
		}
	}
}

func sendGenericMessage(conn *net.UDPConn, address string, msg GenericMessage) error {
	buffer, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Failed to encode messsage: " + err.Error())
	}
	log.Warn("Sending generic message %s", string(buffer))
	return sendOne(conn, address, buffer)
}

func broadcastGenericMsg(conn *net.UDPConn, addresses []string, msg GenericMessage) {
	buffer, err := json.Marshal(msg)
	log.Warn("---wtf---" + string(buffer))
	if err != nil {
		log.Fatal("Failed to encode messsage: " + err.Error())
	}
	broadcast(conn, addresses, buffer)
}

func BroadcastPaxosMessage(conn *net.UDPConn, addresses []string, msg Message) {
	log.Info(fmt.Sprintf("Broadcasting paxos message: %+v", msg))
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
			newbuf := append(make([]byte, 0), buf[:n]...)
			log.Warn("Received message: " + string(newbuf))
			if err != nil {
				log.Warn(fmt.Sprintf("Error read from UDP: " + err.Error()))
				continue
			}
			log.Info(fmt.Sprintf("Received %d bytes from %v", n, addr))

			var msg = GenericMessage{}
			if err := json.Unmarshal(newbuf, &msg); err != nil {
				log.Warn(string(newbuf))
				log.Warn(fmt.Sprintf("Error decoding message: " + err.Error()))
			}
			ch <- msg
		}
	}()
}
