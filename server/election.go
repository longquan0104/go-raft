package main

import (
	"fmt"
	"go-raft/logging"
	"go-raft/model"
	"go-raft/utils"
	"strconv"
	"time"
)

const leaderFileName = "file_leader"

// 1/9 on node nodeId suspects leader has failed, or on election timeout do
func (s *Server) electionTimer() {
	for {
		select {
		case <-s.electionModule.ElectionTimeout.C:
			fmt.Println("Election Timeout")
			if s.role == Follower {
				s.startElection()
			} else {
				s.role = Follower
				s.electionModule.ResetElectionTimer <- struct{}{}
			}
		case <-s.electionModule.ResetElectionTimer:
			fmt.Println("Reset election timer")
			s.electionModule.ElectionTimeout.Reset(time.Duration(s.electionModule.ElectionTimeoutInterval) * time.Millisecond)
		}
	}
}

func (s *Server) startElection() {
	s.serverState.CurrentTerm++
	s.role = Candidate
	s.serverState.VotedFor = s.serverState.Name
	s.peerData.VotesReceived = map[string]bool{s.serverState.Name: true}
	var lastTerm int
	if len(s.logs) > 0 {
		lastTerm = parseLogTerm(s.logs[len(s.logs)-1])
	}
	voteRequest := model.NewVoteRequest(s.serverState.Name, s.serverState.CurrentTerm, lastTerm, len(s.logs))
	servers, err := logging.ListRegisterServers()
	if err != nil {
		fmt.Println("error in startElection()", err.Error())
	}
	for k, v := range servers {
		if k == s.serverState.Name {
			continue
		}
		s.sendMessageToFollowers(voteRequest.String(), v)
	}
	s.checkElectionResult()
}

// 3/9 collecting vote
// on receiving (VoteResponse,voterId,term,granted) at nodeId do
func (s *Server) checkElectionResult() {
	if s.role == Leader {
		return
	}
	votesCount := 0
	allServers, err := logging.ListRegisterServers()
	if err != nil {
		fmt.Println("errors:", err.Error())
	}
	for sv := range allServers {
		if v, ok := s.peerData.VotesReceived[sv]; ok && v {
			votesCount++
		}
	}

	// quorum must be an odd number. eg 5 server has approved of 3 servers for at least
	if votesCount >= (len(allServers)+1)/2 {
		// Got major vote accept
		fmt.Printf("Server %s win election with votes %d from %d servers \n", s.serverState.Name, votesCount, len(allServers))
		s.role = Leader
		s.leaderNodeName = s.serverState.Name
		s.peerData.VotesReceived = make(map[string]bool)
		s.electionModule.ElectionTimeout.Stop()
		utils.ReWrite(leaderFileName, "localhost:"+strconv.Itoa(s.port))
		s.syncUp()
	}
}

// Sync up and Sending HeartBeat
func (s *Server) syncUp() {
	// if s.role != Leader {
	// 	return
	// }
	ticker := time.NewTicker(BroadcastPeriod * time.Millisecond)
	for t := range ticker.C {
		allNodes, _ := logging.ListRegisterServers()
		fmt.Println("Sending heartbeat at ", t)
		for name, port := range allNodes {
			if name == s.serverState.Name {
				continue
			}
			fmt.Printf("Sending heartbeat to server %s at %v \n", name, t)
			s.replicateLog(name, port)
		}
	}
}
