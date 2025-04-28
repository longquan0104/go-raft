package main

import (
	"bufio"
	"fmt"
	"go-raft/utils"
	"io"
	"net"
	"os"
	"strings"
)

const leaderFileName = "file_leader"

func main() {
	CONNECT, err := getInitialConnection(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := establishConnection(CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	handleClient(c)
}

func getInitialConnection(arguments []string) (string, error) {
	if len(arguments) < 2 {
		fmt.Println("Host:port required for client")
		return getConnectionFromLeaderFile()
	}
	return arguments[1], nil
}

func establishConnection(address string) (net.Conn, error) {
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", address, err)
	}
	return c, nil
}

func handleClient(c net.Conn) {
	defer c.Close()
	for {
		text := readInput()
		sendMessage(c, text)

		message, err := receiveMessage(c)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by server.")
				newConnection, err := resetConnection()
				if err != nil {
					fmt.Println("Connection cannot be reset:", err)
					return
				}
				c = newConnection
				continue
			}
			fmt.Println("Error reading from server:", err)
			return
		}

		fmt.Print("> " + message)
		if strings.TrimSpace(text) == "STOP" {
			fmt.Println("TCP client exiting.")
			return
		}
	}
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">")
	text, _ := reader.ReadString('\n')
	return text
}

func sendMessage(c net.Conn, message string) {
	fmt.Fprintf(c, message+"\n")
}

func receiveMessage(c net.Conn) (string, error) {
	return bufio.NewReader(c).ReadString('\n')
}

func resetConnection() (net.Conn, error) {
	CONNECT, err := getConnectionFromLeaderFile()
	if err != nil {
		return nil, err
	}
	return establishConnection(CONNECT)
}

func getConnectionFromLeaderFile() (string, error) {
	leaderFileContent, err := utils.Read(leaderFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read leader file: %v", err)
	}
	return leaderFileContent[0], nil
}
