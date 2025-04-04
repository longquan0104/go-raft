package model

import (
	"fmt"
	"go-raft/logging"
	"strconv"
	"strings"
)

type ServerState struct {
	Name         string
	CurrentTerm  int
	VotedFor     string
	CommitLength int
}

func NewServerState(serverName string) *ServerState {
	return &ServerState{
		Name: serverName,
	}
}

func parseServerState(log string) *ServerState {
	splits := strings.Split(log, "_")
	if len(splits) < 4 {
		return nil
	}
	name := splits[0]
	currentTerm, _ := strconv.Atoi(splits[1])
	votedFor := splits[2]
	commitLength, _ := strconv.Atoi(splits[3])
	return &ServerState{
		Name:         name,
		CurrentTerm:  currentTerm,
		VotedFor:     votedFor,
		CommitLength: commitLength,
	}
}

func GetOrCreateServerState(serverName string) *ServerState {
	log, err := logging.GetLatestServerState(serverName)
	if err != nil {
		return NewServerState(serverName)
	}
	if state := parseServerState(log); state == nil {
		fmt.Println("fail parsing server state " + log)
		return NewServerState(serverName)
	} else {
		return state
	}
}

func (state *ServerState) LogServerPersistedState() error {
	serverStateLog := state.Name + "_" + strconv.Itoa(state.CurrentTerm) + "_" + state.VotedFor + "_" + strconv.Itoa(state.CommitLength)
	return logging.PersistServerState(serverStateLog)
}
