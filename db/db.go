package db

import (
	"errors"
	"fmt"
	"go-raft/utils"
	"strconv"
	"strings"
)

// Database is a mock of kv-storage

type Database struct {
	db map[string]int
}

// Init
func NewDatabase() (*Database, error) {
	return &Database{
		db: make(map[string]int),
	}, nil
}

// get
func (d *Database) get(key string) (int, error) {
	if values, ok := d.db[key]; !ok {
		return -1, errors.New("key " + key + " not found")
	} else {
		return values, nil
	}
}

// set
func (d *Database) set(key string, val int) error {
	d.db[key] = val
	return nil
}

// delete
func (d *Database) delete(key string) error {
	if _, ok := d.db[key]; !ok {
		return errors.New("key " + key + " not found")
	}
	delete(d.db, key)
	return nil
}

// ValidateQuery
func ValidateQuery(query string) error {
	splits := strings.Split(query, " ")
	operations := splits[0]
	switch operations {
	case "GET", "DELETE":
		if len(splits) != 2 {
			return errors.New("input missing key or has a malformed format")
		}
	case "SET":
		if len(splits) != 3 {
			return errors.New("input missing key or value or has a malformed format")
		}
	default:
		return errors.New("invalid command " + operations)
	}
	return nil
}

// ExecuteQuery
func (d *Database) ExecuteQuery(query string) (string, error) {
	if err := ValidateQuery(query); err != nil {
		return "", err
	}
	var resp string
	splits := strings.Split(query, " ")
	operations, key := splits[0], splits[1]
	switch operations {
	case "GET":
		val, err := d.get(key)
		if err != nil {
			return resp, err
		}
		resp = "Value for key (" + key + ") is: " + strconv.Itoa(val)
		return resp, err
	case "SET":
		val := splits[2]
		values, err := strconv.Atoi(val)
		if err != nil {
			return resp, err
		}
		resp = fmt.Sprintf("Key %s is set with value %s", key, val)
		return resp, d.set(key, values)
	case "DELETE":
		resp = fmt.Sprintf("Key %s is deleted from database", key)
		return resp, d.delete(key)
	default:
		return resp, nil
	}
}

func ExtractQuery(message string) string {
	return strings.Split(message, "#")[0]
}

// LogDbCommand logs the database command to log file
func (d *Database) LogCommand(command string, serverName string) error {
	fileName := "file_" + serverName
	var err = utils.CreateFileIfNotExists(fileName)
	if err != nil {
		return err
	}
	err = utils.Write(fileName, serverName+","+command+"\n")
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) RebuildLogIfExists(serverName string) []string {
	logs := make([]string, 0)
	fileName := "file_" + serverName
	utils.CreateFileIfNotExists(fileName)
	lines, _ := utils.Read(fileName)
	for _, line := range lines {
		splits := strings.Split(line, ",")
		logs = append(logs, splits[1])
	}
	return logs
}
