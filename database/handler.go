package database

import (
	"encoding/csv"
	"fmt"
	"just-notify/config"
	"os"
	"path/filepath"
	"strconv"
)

type LogEntry struct {
	InitTime int64
	EndTime  int64
	TaskName string
}

func LogData(data LogEntry) error {
	config := config.LoadConfig()

	filePath := config["CSV_PATH"]
	connStr := config["CONN"]

	// Default path
	if connStr == "" && filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filePath = filepath.Join(home, ".jn.csv")
	}

	if connStr != "" {
		conn, err := OpenDB(connStr)

		if err != nil {
			return err
		}

		defer conn.Close()

		if err := conn.InitSchema(); err != nil {
			return err
		}

		if err := conn.Insert(&data); err != nil {
			return err
		}

		return nil
	}

	// Check if file exists to determine if headers are needed
	needsHeader := !fileExists(filePath)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if new file
	if needsHeader {
		headers := []string{"init_time_ms", "end_time_ms", "task_name"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}

	// Write the actual data
	return writer.WriteAll([][]string{{strconv.Itoa(int(data.InitTime)), strconv.Itoa(int(data.EndTime)), data.TaskName}})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
