package utils

import (
	"bufio"
	"os"
)

// CreateFileIfNotExists creates a file with given fileName if it doesn't exists
func CreateFileIfNotExists(fileName string) error {
	var _, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		var file, err = os.Create(fileName)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

// Write append message to a file
func Write(fName string, message string) error {
	file, err := os.OpenFile(fName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = file.WriteString(message)
	if err != nil {
		return err
	}
	return file.Close()
}

// Read get messages from a file.
func Read(fName string) ([]string, error) {
	file, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}
