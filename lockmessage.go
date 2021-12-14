package paxos

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
)

type LockMessage struct {
	Lock     bool
	LockId   int
	ClientID string
	MsgUUID  uuid.UUID
}

type LockRelayMessage struct {
	Lock           bool
	LockId         int
	OriginServerId int
	ClientAddr     *net.UDPAddr
	ClientID       string
	MsgUUID        uuid.UUID
}

type LockResult int

const (
	SUCCESS      LockResult = 0
	FAIL         LockResult = 1
	RETRY        LockResult = 2
	DEADLOCK     LockResult = 3
	NO_SUCH_LOCK LockResult = 4
	TIMEOUT      LockResult = 5
	ALREADY_HELD LockResult = 6
)

type LockReplyMessage struct {
	Result    LockResult
	RetryAddr *string
}

func FormatLockResult(v LockResult) string {
	switch v {
	case SUCCESS:
		return "Success"
	case FAIL:
		return "Failed"
	case RETRY:
		return "Retry"
	case DEADLOCK:
		return "Deadlock detected"
	case NO_SUCH_LOCK:
		return "No such lock"
	case TIMEOUT:
		return "Timeout"
	case ALREADY_HELD:
		return "Already held"
	default:
		return "Unknown result"
	}
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
			log.Debug(fmt.Sprintf("Received %d bytes from %v", n, addr))
			var msg = LockMessage{}
			if err := json.Unmarshal(newbuf, &msg); err != nil {
				log.Warn(string(newbuf))
				log.Warn(fmt.Sprintf("Error decoding message: " + err.Error()))
			}
			ch <- GenericMessage{LockRelay: &LockRelayMessage{msg.Lock, msg.LockId, selfId, addr, msg.ClientID, msg.MsgUUID}}
		}
	}()
}
