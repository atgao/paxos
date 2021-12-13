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

func formatPaxosMessage(paxos *Message) string {
	header := fmt.Sprintf("PaxosMessage{SenderId: %d, Uuid: %v, ", paxos.SenderId, paxos.Uuid)
	trailer := "}"
	if paxos.Prepare != nil {
		return header + fmt.Sprintf("%+v", *paxos.Prepare) + trailer
	} else if paxos.PrepareResponse != nil {
		return header + fmt.Sprintf("%+v", *paxos.PrepareResponse) + trailer
	} else if paxos.Accept != nil {
		return header + fmt.Sprintf("%+v", *paxos.Accept) + trailer
	} else if paxos.AcceptResponse != nil {
		return header + fmt.Sprintf("%+v", *paxos.AcceptResponse) + trailer
	} else if paxos.Success != nil {
		return header + fmt.Sprintf("%+v", *paxos.Success) + trailer
	} else if paxos.SuccessResponse != nil {
		return header + fmt.Sprintf("%+v", *paxos.SuccessResponse) + trailer
	} else {
		return "Bad Paxos Message"
	}
}

func formatGenericMessage(message *GenericMessage) string {
	if message.LockRelay != nil {
		return fmt.Sprintf("%+v", *message.LockRelay)
	} else if message.Paxos != nil {
		return formatPaxosMessage(message.Paxos)
	} else if message.KeepAlive != nil {
		return fmt.Sprintf("%+v", *message.KeepAlive)
	} else {
		return "Unknown message type"
	}
}

func sendGenericMessage(conn *net.UDPConn, address string, msg GenericMessage) error {
	buffer, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Failed to encode messsage: " + err.Error())
	}
	log.Trace(fmt.Sprintf("SENDING message to %s: %s", address, formatGenericMessage(&msg)))
	return sendOne(conn, address, buffer)
}

func broadcastGenericMsg(conn *net.UDPConn, addresses []string, msg GenericMessage) {
	buffer, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Failed to encode messsage: " + err.Error())
	}
	log.Trace(fmt.Sprintf("BROADCASTING message: %s", formatGenericMessage(&msg)))
	broadcast(conn, addresses, buffer)
}

func BroadcastPaxosMessage(conn *net.UDPConn, addresses []string, msg Message) {
	gmsg := GenericMessage{&msg, nil, nil}
	broadcastGenericMsg(conn, addresses, gmsg)
}

func BroadcastKeepAliveMessage(conn *net.UDPConn, addresses []string, msg KeepAliveMessage) {
	broadcastGenericMsg(conn, addresses, GenericMessage{nil, &msg, nil})
}

func SendLockRelayMessage(conn *net.UDPConn, address string, msg LockRelayMessage) {
	sendGenericMessage(conn, address, GenericMessage{LockRelay: &msg})
}

func UDPServeGenericMessage(conn *net.UDPConn, ch chan GenericMessage) {
	buf := make([]byte, 1024)
	go func() {
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			newbuf := append(make([]byte, 0), buf[:n]...)
			if err != nil {
				log.Warn(fmt.Sprintf("Error read from UDP: " + err.Error()))
				continue
			}

			var msg = GenericMessage{}
			if err := json.Unmarshal(newbuf, &msg); err != nil {
				log.Warn(string(newbuf))
				log.Warn(fmt.Sprintf("Error decoding message: " + err.Error()))
				continue
			}
			log.Trace(fmt.Sprintf("RECEIVING message from %+v: %s", addr, formatGenericMessage(&msg)))
			ch <- msg
		}
	}()
}
