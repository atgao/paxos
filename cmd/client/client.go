package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)
import "github.com/atgao/paxos"

type command interface {
	execute()
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

func (comm SleepCommand) execute() {
	time.Sleep(comm.time)
}

func (comm LockCommand) execute() {
	paxos.RequestLockServer(comm.server, true)
}

func (comm UnlockCommand) execute() {
	paxos.RequestLockServer(comm.server, false)
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
	default:
		return nil, errors.New("failed to parse command")
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		comm, err := parser(scanner.Text())
		if err != nil {
			fmt.Println(err)
		} else {
			comm.execute()
		}
	}
}
