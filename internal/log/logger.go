package log

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// This file contains the logger implementation for the application.
const LogFile = "internal/log/logs/vpn-manager.log"

type Log struct {
	Time    string
	Level   string
	Message string
	Source  string
}

func CreateLogFile() error {
	_, err := os.Stat(LogFile)
	if err == nil {
		// The file already exists
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		// Errors Other Than “File Not Found”
		return err
	}

	dirPath := "internal/log/logs"

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// If the file doesn't exist, create it
	if err := os.WriteFile(LogFile, nil, 0644); err != nil {
		return err
	}

	return nil
}

func WriteLog(time, level, message, source string) error {
	data := Log{
		Time:    time,
		Level:   level,
		Message: message,
		Source:  source,
	}

	// Ensure the log file exists before writing
	file, err := os.OpenFile(LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// Write the log entry to the file
	logEntry := fmt.Sprintf("%s [%s] %s (%s)\n", data.Time, data.Level, data.Message, data.Source)

	if _, err := file.WriteString(logEntry); err != nil {
		file.Close()
		return err
	}

	// Close the log file and check for errors
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func ReadLogs() ([]Log, error) {
	file, err := os.ReadFile(LogFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := CreateLogFile(); err != nil {
				return nil, err
			}
			return []Log{}, nil
		}
		return nil, err
	}

	var logs []Log
	lines := strings.Split(string(file), "\n")

	// Parse each line into a Log struct
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Extract log information from each line
		var time, level, message, source string
		_, err := fmt.Sscanf(line, "%s [%s", &time, &level)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse log entry: %w", err)
		}

		level = strings.TrimSuffix(level, "]")

		start := strings.Index(line, "] ")
		end := strings.LastIndex(line, " (")
		if start == -1 || end == -1 || end < start {
			return nil, fmt.Errorf("Invalid log format: %s", line)
		}

		message = line[start+2 : end]
		source = strings.TrimSuffix(line[end+2:], ")")

		logs = append(logs, Log{
			Time:    time,
			Level:   level,
			Message: message,
			Source:  source,
		})
	}

	return logs, nil
}
