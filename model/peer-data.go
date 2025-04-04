package model

// PeerData hold the information of followers or when acting as a follower
type PeerData struct {
	VotesReceived  map[string]bool // Votes received from other servers.
	AckedLength    map[string]int  // Log that has acknowledgements from other server.
	SentLength     map[string]int  // Length of the log that sent to followers
	SuspectedNodes map[int]bool    // Node that suspect are being unable to connect to using port is easier to be defined here rather than servername
}

func NewPeerData() *PeerData {
	return &PeerData{
		VotesReceived:  make(map[string]bool),
		AckedLength:    make(map[string]int),
		SentLength:     make(map[string]int),
		SuspectedNodes: make(map[int]bool),
	}
}
