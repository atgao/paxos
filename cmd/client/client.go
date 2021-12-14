package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atgao/paxos"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type command interface {
	execute(clientID **string, msgUUID uuid.UUID)
}

type SleepCommand struct {
	time time.Duration
}

type LockCommand struct {
	lockId int
	server string
}

type UnlockCommand struct {
	lockId int
	server string
}

type SetIdCommand struct {
	id string
}

func (comm SleepCommand) execute(clientID **string, msgUUID uuid.UUID) {
	time.Sleep(comm.time)
}

func (comm LockCommand) execute(clientID **string, msgUUID uuid.UUID) {
	if *clientID == nil {
		log.Warn("ClientID is nil, please run setid <yourid>")
		return
	}
	res := paxos.RequestLockServer(comm.server, true, comm.lockId, **clientID, msgUUID)
	log.Info(fmt.Sprintf("Lock command execution result: %s", paxos.FormatLockResult(res)))
}

func (comm UnlockCommand) execute(clientID **string, msgUUID uuid.UUID) {
	if *clientID == nil {
		log.Warn("ClientID is nil, please run setid <yourid>")
		return
	}
	res := paxos.RequestLockServer(comm.server, false, comm.lockId, **clientID, msgUUID)
	log.Info(fmt.Sprintf("Unlock command execution result: %v", paxos.FormatLockResult(res)))
}

func (comm SetIdCommand) execute(clientID **string, msgUUID uuid.UUID) {
	*clientID = &comm.id
	log.Info(fmt.Sprintf("Id set to: %s", comm.id))
}

func parser(line string) (command, error) {
	splitted := strings.Split(line, " ")
	filtered := []string{}
	for _, s := range splitted {
		if len(s) != 0 {
			filtered = append(filtered, s)
		}
	}
	switch {
	case filtered[0] == "sleep":
		if len(filtered) != 2 {
			return nil, errors.New("failed to parse sleep command")
		}
		s1, err := time.ParseDuration(filtered[1])
		if err != nil {
			return nil, errors.New("failed to parse sleep duration")
		}
		return SleepCommand{s1}, nil
	case filtered[0] == "lock":
		if len(filtered) != 3 {
			return nil, errors.New("failed to parse lock command")
		}
		lockId, err := strconv.Atoi(filtered[1])
		if err != nil {
			return nil, errors.New("failed to parse lock id")
		}
		return LockCommand{lockId, filtered[2]}, nil
	case filtered[0] == "unlock":
		if len(filtered) != 3 {
			return nil, errors.New("failed to parse unlock command")
		}
		lockId, err := strconv.Atoi(filtered[1])
		if err != nil {
			return nil, errors.New("failed to parse lock id")
		}
		return UnlockCommand{lockId, filtered[2]}, nil
	case filtered[0] == "setid":
		if len(filtered) != 2 {
			return nil, errors.New("failed to parse setid command")
		}
		return SetIdCommand{filtered[1]}, nil
	default:
		return nil, errors.New("failed to parse command")
	}
}

func main() {

	/*
		clientAddr := flag.String("address", "", "Client address")
		flag.Parse()
		if *clientAddr == "" {
			panic("No client address specified")
		}
	*/

	/*
		addr, err := net.ResolveUDPAddr("udp", *clientAddr)
		if err != nil {
			panic("Failed to resolve client address: " + err.Error())
		}
		clientSock, err := net.ListenUDP("udp", addr)
		if err != nil {
			panic("Failed to open client socket: " + err.Error())
		}
		log.Info(fmt.Sprintf("Started client on %s", *clientAddr))
	*/

	var clientID *string = nil
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		comm, err := parser(scanner.Text())
		if err != nil {
			fmt.Println(err)
		} else {
			//for i := 0; i != 10; i++ {
			func() {
				msgUUID := uuid.New()
				comm.execute(&clientID, msgUUID)
			}()
			// }
		}
	}
}
