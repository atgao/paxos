package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
	server string
}

type UnlockCommand struct {
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
	res := paxos.RequestLockServer(comm.server, true, **clientID, msgUUID)
	log.Info(fmt.Sprintf("Lock command execution result: %v", res))
}

func (comm UnlockCommand) execute(clientID **string, msgUUID uuid.UUID) {
	if *clientID == nil {
		log.Warn("ClientID is nil, please run setid <yourid>")
		return
	}
	res := paxos.RequestLockServer(comm.server, false, **clientID, msgUUID)
	log.Info(fmt.Sprintf("Unlock command execution result: %v", res))
}

func (comm SetIdCommand) execute(clientID **string, msgUUID uuid.UUID) {
	*clientID = &comm.id
}

func parser(line string) (command, error) {
	splitted := strings.Split(line, " ")
	filtered := []string{}
	for _, s := range splitted {
		if len(s) != 0 {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) != 2 {
		return nil, errors.New("please enter a command")
	}
	switch {
	case filtered[0] == "sleep":
		s1, err := time.ParseDuration(filtered[1])
		if err != nil {
			return nil, errors.New("failed to parse sleep duration")
		}
		return SleepCommand{s1}, nil
	case filtered[0] == "lock":
		return LockCommand{filtered[1]}, nil
	case filtered[0] == "unlock":
		return UnlockCommand{filtered[1]}, nil
	case filtered[0] == "setid":
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
			for i := 0; i != 10; i++ {
				go func() {
					msgUUID := uuid.New()
					comm.execute(&clientID, msgUUID)
				}()
			}
		}
	}
}
