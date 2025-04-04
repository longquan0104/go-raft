package model

import "time"

type ElectionModule struct {
	ElectionTimeout         *time.Ticker
	ResetElectionTimer      chan struct{}
	ElectionTimeoutInterval int
}

func NewElectionModule(electionTimeOutInterval int) *ElectionModule {
	return &ElectionModule{
		ElectionTimeout:         time.NewTicker(time.Duration(electionTimeOutInterval) * time.Millisecond),
		ResetElectionTimer:      make(chan struct{}),
		ElectionTimeoutInterval: electionTimeOutInterval,
	}
}
