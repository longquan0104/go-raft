package model

import (
	"strconv"
	"strings"
)

type VoteResponse struct {
	VoteForNodeId string
	CurrentTerm   int
	VoteSuccess   bool
}

func (vr *VoteResponse) String() string {
	return "VoteResponse" + "|" + vr.VoteForNodeId + "|" + strconv.Itoa(vr.CurrentTerm) + "|" + strconv.FormatBool(vr.VoteSuccess)
}

func ParseVoteResponse(message string) (*VoteResponse, error) {
	splits := strings.Split(message, "|")
	var err error
	_, err = strconv.Atoi(splits[2])
	if err != nil {
		return nil, err
	}
	currentTerm, _ := strconv.Atoi(splits[2])
	_, err = strconv.ParseBool(splits[3])
	if err != nil {
		return nil, err
	}
	voteInFavor, _ := strconv.ParseBool(splits[3])
	return NewVoteResponse(splits[1], currentTerm, voteInFavor), nil
}

func NewVoteResponse(voteForNodeId string, currentTerm int, voteSuccess bool) *VoteResponse {
	return &VoteResponse{
		VoteForNodeId: voteForNodeId,
		CurrentTerm:   currentTerm,
		VoteSuccess:   voteSuccess,
	}
}
