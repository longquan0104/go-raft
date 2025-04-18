package model

import (
	"errors"
	"strconv"
	"strings"
)

type VoteRequest struct {
	CandidateId        string
	CandidateTerm      int
	CandidateLogTerm   int
	CandidateLogLength int
}

func (r *VoteRequest) String() string {
	return "VoteRequest" + "|" + r.CandidateId + "|" + strconv.Itoa(r.CandidateTerm) + "|" + strconv.Itoa(r.CandidateLogLength) + "|" + strconv.Itoa(r.CandidateLogTerm)
}

func ParseVoteRequest(message string) (*VoteRequest, error) {
	splits := strings.Split(message, "|")
	if len(splits) < 5 {
		return nil, errors.New("invalid VoteRequest message")
	}
	candidateId := splits[1]
	candidateTerm, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, errors.New("invalid VoteRequest message")
	}
	candidateLogTerm, err := strconv.Atoi(splits[3])
	if err != nil {
		return nil, errors.New("invalid VoteRequest message")
	}
	candidateLogLength, err := strconv.Atoi(splits[4])
	if err != nil {
		return nil, errors.New("invalid VoteRequest message")
	}
	return NewVoteRequest(candidateId, candidateTerm, candidateLogTerm, candidateLogLength), nil
}

func NewVoteRequest(
	candidateId string,
	candidateTerm int,
	candidateLogTerm int,
	candidateLogLength int,
) *VoteRequest {
	return &VoteRequest{
		CandidateId:        candidateId,
		CandidateTerm:      candidateTerm,
		CandidateLogTerm:   candidateLogTerm,
		CandidateLogLength: candidateLogLength,
	}
}
