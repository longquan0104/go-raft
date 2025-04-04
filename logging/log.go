package logging

import (
	"errors"
	"go-raft/utils"
	"strconv"
	"strings"
)

// Register server
const registryFileName string = "file_registry"

func RegisterServer(name, port string) error {
	if err := utils.CreateFileIfNotExists(registryFileName); err != nil {
		return err
	}
	registerLog := name + "_" + port + "\n"
	return utils.Write(registryFileName, registerLog)
}

func ListRegisterServers() (map[string]int, error) {
	registerdLines, err := utils.Read(registryFileName)
	if err != nil {
		return nil, err
	}

	servers := map[string]int{}
	for _, line := range registerdLines {
		splits := strings.Split(line, "_")
		name := splits[0]
		port, err := strconv.Atoi(splits[1])
		if err != nil {
			return nil, err
		}
		servers[name] = port
	}
	return servers, nil
}

// Server state operators
const serverStateFileName string = "file_server-state"

// Save all state log to a storage (file)
func PersistServerState(serverStateLog string) error {
	if err := utils.CreateFileIfNotExists(serverStateFileName); err != nil {
		return err
	}
	return utils.Write(serverStateFileName, serverStateLog+"\n")
}

// Get latest statelog of a server from a storage (file)
func GetLatestServerState(serverName string) (string, error) {
	stateLogs, err := utils.Read(serverStateFileName)
	if err != nil {
		return "", err
	}
	stateLog := ""
	for _, line := range stateLogs {
		splits := strings.Split(line, "_")
		if splits[0] == serverName {
			stateLog = line
		}
	}
	if stateLog == "" {
		return "", errors.New("no existing server state found")
	}
	return stateLog, nil
}
