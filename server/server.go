package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"go-raft/db"
	"go-raft/logging"
	"go-raft/model"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

type serverRole int

const (
	Follower serverRole = iota
	Candidate
	Leader
	BroadcastPeriod    = 3000
	ElectionMinTimeout = 3001
	ElectionMaxTimeout = 10000
)

var (
	serverName = flag.String("server-name", "", "the server's name")
	port       = flag.String("port", "", "the server's port")
)

type Server struct {
	port           int
	db             *db.Database
	logs           []string
	serverState    *model.ServerState
	role           serverRole
	leaderNodeName string
	peerData       *model.PeerData
	electionModule *model.ElectionModule
}

// 1/9 on initialization do
func main() {
	// Init server
	// Get server (name, port)
	parseFlags()
	// init database
	db, err := db.NewDatabase()
	if err != nil {
		panic(err)
	}
	// Init election module
	electionTimeOutInterval := rand.Intn(ElectionMaxTimeout-ElectionMinTimeout) + ElectionMinTimeout
	electionTimeOutModule := model.NewElectionModule(electionTimeOutInterval)

	// 1/9 on recovery from crash do
	state := model.GetOrCreateServerState(*serverName)
	// Register server to cluster
	if err := logging.RegisterServer(*serverName, *port); err != nil {
		fmt.Println(err)
		panic(err)
	}
	portNum, _ := strconv.Atoi(*port)
	s := &Server{
		port:           portNum,
		db:             db,
		logs:           db.RebuildLogIfExists(*serverName),
		serverState:    state,
		role:           Follower,
		peerData:       model.NewPeerData(),
		electionModule: electionTimeOutModule,
	}
	s.serverState.LogServerPersistedState()

	l, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	defer l.Close()
	go s.electionTimer()
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		go s.handleConnection(c)
	}

}

// 2/9 handling vote request
// on receiving (VoteRequest,cId,cTerm,cLogLength,cLogTerm)
func (s *Server) handleVoteRequest(request string) string {
	voteRequest, err := model.ParseVoteRequest(request)
	if err != nil {
		fmt.Println("handling vote request fail in parsing", err.Error())
	}
	if voteRequest.CandidateTerm > s.serverState.CurrentTerm {
		s.serverState.CurrentTerm = voteRequest.CandidateTerm
		s.role = Follower
		s.serverState.VotedFor = ""
		s.electionModule.ResetElectionTimer <- struct{}{}
	}
	var lastTerm int
	if len(s.logs) > 0 {
		lastTerm = parseLogTerm(s.logs[(len(s.logs) - 1)])
	}
	// ! supposed to be voteRequest.CandidateLogLength > len(s.logs)
	logOk := voteRequest.CandidateLogTerm > lastTerm ||
		(voteRequest.CandidateLogTerm == lastTerm && voteRequest.CandidateLogLength >= len(s.logs))
	if voteRequest.CandidateTerm == s.serverState.CurrentTerm && logOk && (s.serverState.VotedFor == voteRequest.CandidateId || s.serverState.VotedFor == "") {
		s.serverState.VotedFor = voteRequest.CandidateId
		// Send message to every node
		s.serverState.LogServerPersistedState()
		return model.NewVoteResponse(s.serverState.Name, s.serverState.CurrentTerm, true).String()
	} else {
		return model.NewVoteResponse(s.serverState.Name, s.serverState.CurrentTerm, false).String()
	}
}

// 3/9 Collecting Vote
// on receiving (VoteResponse,voterId,term,granted) at nodeId do
func (s *Server) handleVoteResponse(request string) {
	voteResponse, err := model.ParseVoteResponse(request)
	if err != nil {
		fmt.Println("handling vote response fail in parsing ", err.Error())
	}
	// Convert to follower if term is higher than current
	if s.serverState.CurrentTerm < voteResponse.CurrentTerm {
		if s.role != Leader {
			s.electionModule.ResetElectionTimer <- struct{}{}
		}
		s.role = Follower
		s.serverState.CurrentTerm = voteResponse.CurrentTerm
		s.serverState.VotedFor = ""
	}

	// If vote success -> join check result
	if s.role == Candidate && s.serverState.CurrentTerm == voteResponse.CurrentTerm && voteResponse.VoteSuccess {
		s.peerData.VotesReceived[voteResponse.VoteForNodeId] = true
		s.checkElectionResult()
	}
}

// 5/9 replicating from leader to followers
func (s *Server) replicateLog(followerName string, followerPort int) {
	// TODO: Check this
	if followerName == s.serverState.Name {
		go s.commitLogEntries()
		return
	}

	prefixLength := s.peerData.SentLength[followerName]
	suffix := s.logs[prefixLength:]
	var prefixTerm int
	if prefixLength > 0 {
		prefixTerm = parseLogTerm(s.logs[prefixLength-1])
	}
	logRequest := model.NewLogRequest(s.serverState.Name, s.serverState.CurrentTerm, prefixLength, prefixTerm, s.serverState.CommitLength, suffix)
	s.sendMessageToFollowers(logRequest.String(), followerPort)
}

// 6/9 Followers receiving messages
// on receiving (LogRequest,leaderId,term,prefixLen,prefixTerm, leaderCommit,suffix) at node nodeId do
func (s *Server) handleLogRequest(request string) string {
	s.electionModule.ResetElectionTimer <- struct{}{}
	logRequest, err := model.ParseLogRequest(request)
	if err != nil {
		return fmt.Sprintf("Handling Log Request meet error: %s", err.Error())
	}
	if logRequest.CurrentTerm > s.serverState.CurrentTerm {
		s.serverState.CurrentTerm = logRequest.CurrentTerm
		s.serverState.VotedFor = ""
	}
	if s.serverState.CurrentTerm == logRequest.CurrentTerm {
		if s.role == Leader {
			go s.electionTimer()
		}
		s.role = Follower
		s.leaderNodeName = logRequest.LeaderId
	}
	logOk := len(s.logs) >= logRequest.PrefixLength &&
		(logRequest.PrefixLength == 0 || parseLogTerm(s.logs[logRequest.PrefixLength-1]) == logRequest.PrefixTerm)
	if logRequest.CurrentTerm == s.serverState.CurrentTerm && logOk {
		s.appendEntries(logRequest.PrefixLength, logRequest.CommitLength, logRequest.Suffix)
		ack := logRequest.PrefixLength + len(logRequest.Suffix)
		logResposne := model.NewLogResponse(s.serverState.Name, s.port, s.serverState.CurrentTerm, ack, true)
		return logResposne.String()
	} else {
		logResposne := model.NewLogResponse(s.serverState.Name, s.port, s.serverState.CurrentTerm, 0, false)
		return logResposne.String()
	}
}

// 7/9 updating followers’ logs
func (s *Server) appendEntries(prefixLength int, leaderCommitLength int, suffix []string) {
	if len(suffix) > 0 && len(s.logs) > prefixLength {
		index := min(len(s.logs), prefixLength+len(suffix)) - 1
		if parseLogTerm(s.logs[index]) != parseLogTerm(suffix[index-prefixLength]) {
			s.logs = s.logs[:prefixLength]
		}
	}
	if prefixLength+len(suffix) > len(s.logs) {
		for i := len(s.logs) - prefixLength; i < len(suffix); i++ {
			// Add log to server logs.
			s.logs = append(s.logs, suffix[i])
			err := s.db.LogCommand(suffix[i], s.serverState.Name)
			if err != nil {
				fmt.Println("Error on logging command" + err.Error())
			}
		}
	}
	if leaderCommitLength > s.serverState.CommitLength {
		for i := s.serverState.CommitLength; i < leaderCommitLength; i++ {
			s.db.ExecuteQuery(db.ExtractQuery(s.logs[i]))
		}
		s.serverState.CommitLength = leaderCommitLength
		s.serverState.LogServerPersistedState()
	}
}

// 8/9 leader receiving log acknowledgements
// on receiving (LogResponse,follower,term,ack,success) at nodeId do
func (s *Server) handleLogResponse(request string) string {
	logResponse, err := model.ParseLogResponse(request)
	if err != nil {
		return fmt.Sprintf("Handling Log Response has error on parsing: %s", err.Error())
	}
	if logResponse.CurrentTerm == s.serverState.CurrentTerm && s.role == Leader {
		if logResponse.ReplicationSuccessful && logResponse.AckLength >= s.peerData.AckedLength[logResponse.FollowerName] {
			s.peerData.SentLength[logResponse.FollowerName] = logResponse.AckLength
			s.peerData.AckedLength[logResponse.FollowerName] = logResponse.AckLength
			s.commitLogEntries()
		} else if s.peerData.SentLength[logResponse.FollowerName] > 0 {
			s.peerData.SentLength[logResponse.FollowerName] -= 1
			s.replicateLog(logResponse.FollowerName, logResponse.FollowerPort)
		}
	} else if logResponse.CurrentTerm > s.serverState.CurrentTerm {
		s.serverState.CurrentTerm = logResponse.CurrentTerm
		s.role = Follower
		s.serverState.VotedFor = ""
		go s.electionTimer()
	}
	return "Handling Log Response successful"
}

// 9/9 leader committing log entries
func (s *Server) commitLogEntries() {
	allNodes, err := logging.ListRegisterServers()
	if err != nil {
		fmt.Printf("commitLogEntry has error in getting all server: %s", err.Error())
	}
	legitNodeCount := len(allNodes) - len(s.peerData.SuspectedNodes)
	for i := s.serverState.CommitLength; i < len(s.logs); i++ {
		ackCount := 0
		for node := range allNodes {
			if s.peerData.AckedLength[node] > s.serverState.CommitLength {
				ackCount++
			}
		}
		if ackCount >= (legitNodeCount+1)/2 || legitNodeCount == 1 {
			log := s.logs[i]
			query := db.ExtractQuery(log)
			_, err = s.db.ExecuteQuery(query)
			if err != nil {
				fmt.Printf("Error on executing query of commit log entry of server %s is %s", s.serverState.Name, err.Error())
			}
			s.serverState.CommitLength++
			err = s.serverState.LogServerPersistedState()
			if err != nil {
				fmt.Printf("Error on log persisted state of server %s is %s", s.serverState.Name, err.Error())
			}
		} else {
			break
		}
	}
}

// Used by all the nodes with all states
func (s *Server) handleConnection(c net.Conn) {
	defer c.Close()
	for {
		// Read request
		data, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Println("error in reading request ", err.Error())
				panic(err)
			}
			continue
		}
		request := strings.TrimSpace(data)
		if request == "invalid command" || request == "Handling Log Response successful" {
			continue
		}
		fmt.Printf("> Message: %s \n", request)
		// Handle request
		// Log (Request and Response)
		var response string
		if strings.HasPrefix(request, "LogRequest") {
			fmt.Println("Handling LogRequest")
			response = s.handleLogRequest(request)
		} else if strings.HasPrefix(request, "LogResponse") {
			fmt.Println("Handling LogResponse")
			response = s.handleLogResponse(request)
		}
		// Vote (Request and Response)
		if strings.HasPrefix(request, "VoteRequest") {
			fmt.Println("Handling VoteRequest")
			response = s.handleVoteRequest(request)
		} else if strings.HasPrefix(request, "VoteResponse") {
			fmt.Println("Handling VoteResponse")
			s.handleVoteResponse(request)
		}
		// 4/9 Broadcasting messages
		// response is empty means this is not log request/response neither vote request/response
		if err := db.ValidateQuery(request); err == nil && s.role == Leader {
			// Execute Query in Message Get to get data from database
			// Get values from the leader.
			if strings.HasPrefix(request, "GET") {
				if err := db.ValidateQuery(request); err != nil {
					response = fmt.Sprintf("Validate query got error %s ", err.Error())
				} else {
					response, err = s.db.ExecuteQuery(request)
					if err != nil {
						response = fmt.Sprintf("Execute query got error %s ", err.Error())
					}
				}
			} else {
				// on request to broadcast msg at node nodeId do
				logMessage := createLogMessage(request, s.serverState.CurrentTerm)
				s.peerData.AckedLength[s.serverState.Name] = len(s.logs)
				s.logs = append(s.logs, logMessage)
				if err := s.db.LogCommand(logMessage, s.serverState.Name); err != nil {
					response = "Logging command meets error"
				}
				allNodes, err := logging.ListRegisterServers()
				if err != nil {
					response = fmt.Sprintf("Meet error when list all servers %s", err.Error())
				} else {
					for serverName, serverPort := range allNodes {
						s.replicateLog(serverName, serverPort)
					}
					// periodically at node nodeId do
					for s.serverState.CommitLength < len(s.logs) {
						fmt.Println("Waiting for consensus")
					}
					response = "operation sucessful"
				}

			}
		}
		// write response
		if response != "" {
			c.Write([]byte(response + "\n"))
		}
	}
}
