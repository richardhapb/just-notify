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
	InitTime    int64
	EndTime     int64
	Category    string
	Description string
}

func LogData(data LogEntry, useDatabase bool, connString ...string) error {
	config := config.LoadConfig()

	var connStr string

	filePath := config["CSV_PATH"]
	if len(connString) > 0 && connString[0] != "" {
		connStr = connString[0]
	} else {
		connStr = config["CONN"]
	}

	if useDatabase && connStr == "" {
		return fmt.Errorf("The database is enabled, but there is no connection string provided.")
	}

	// Default path
	if connStr == "" && filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filePath = filepath.Join(home, ".jn.csv")
	}

	if connStr != "" && useDatabase {
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
		headers := []string{"init_time_ms", "end_time_ms", "category", "description"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}

	// Write the actual data
	return writer.WriteAll([][]string{{strconv.Itoa(int(data.InitTime)), strconv.Itoa(int(data.EndTime)), data.Category, data.Description}})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
