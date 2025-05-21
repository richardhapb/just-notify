package database

import (
	"os"
	"path/filepath"
	"encoding/csv"
	"fmt"
	"just-notify/config"
)

func LogData(data [][]string) error {
	config := config.LoadConfig()

	filePath := config["CSV_PATH"]

	// Default path
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filePath = filepath.Join(home, ".jn.csv")
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
	return writer.WriteAll(data)
}


func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
