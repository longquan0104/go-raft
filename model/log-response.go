package model

import (
	"strconv"
	"strings"
)

type LogResponse struct {
	FollowerName          string
	FollowerPort          int
	CurrentTerm           int
	AckLength             int
	ReplicationSuccessful bool
}

func (l *LogResponse) String() string {
	return "LogResponse" + "|" + l.FollowerName + "|" + strconv.Itoa(l.FollowerPort) + "|" + strconv.Itoa(l.CurrentTerm) + "|" + strconv.Itoa(l.AckLength) + "|" + strconv.FormatBool(l.ReplicationSuccessful)
}

func ParseLogResponse(message string) (*LogResponse, error) {
	splits := strings.Split(message, "|")
	var err error
	_, err = strconv.Atoi(splits[2])
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(splits[2])
	_, err = strconv.Atoi(splits[3])
	if err != nil {
		return nil, err
	}
	currentTerm, _ := strconv.Atoi(splits[3])
	_, err = strconv.Atoi(splits[4])
	if err != nil {
		return nil, err
	}
	ackLength, _ := strconv.Atoi(splits[4])
	_, err = strconv.ParseBool(splits[5])
	if err != nil {
		return nil, err
	}
	replicationSuccessful, _ := strconv.ParseBool(splits[5])
	return NewLogResponse(splits[1], port, currentTerm, ackLength, replicationSuccessful), nil
}

func NewLogResponse(nodeId string, port int, currentTerm int, ackLength int, replicationSuccessful bool) *LogResponse {
	return &LogResponse{
		FollowerName:          nodeId,
		FollowerPort:          port,
		CurrentTerm:           currentTerm,
		AckLength:             ackLength,
		ReplicationSuccessful: replicationSuccessful,
	}
}
