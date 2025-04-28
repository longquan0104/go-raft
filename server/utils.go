package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func parseLogTerm(message string) int {
	// message has the format: log#term
	split := strings.Split(message, "#")
	if len(split) < 2 {
		fmt.Println(message, "error in parsing log term")
		panic(fmt.Errorf(message, "error in parsing log term"))
	}
	pTerm, err := strconv.Atoi(split[1])
	if err != nil {
		fmt.Println(message, "error in parsing log term", err.Error())
	}
	return pTerm
}

func parseFlags() {
	flag.Parse()

	if *serverName == "" {
		log.Fatalf("Must provide serverName for the server")
	}

	if *port == "" {
		log.Fatalf("Must provide a port number for server to run")
	}
}

func createLogMessage(message string, term int) string {
	return fmt.Sprintf("%s#%d", message, term)
}

func (s *Server) sendMessageToFollowers(message string, port int) {
	c, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error in sendMessageToFollowers", err)
		s.peerData.SuspectedNodes[port] = true
		return
	}
	delete(s.peerData.SuspectedNodes, port)
	fmt.Fprintf(c, message+"\n")
	go s.handleConnection(c)
}
