package database

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type CSVWriter interface {
	Write(*LogEntry) error
}

type CSV struct {
	path   string
}

func NewCSV(path string) (*CSV, error) {
	// Default path
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".jn.csv")
	}

	return &CSV{
		path:   path,
	}, nil
}

func (c *CSV) Write(entry *LogEntry) error {

	// Check if file exists to determine if headers are needed
	needsHeader := !fileExists(c.path)

	file, err := os.OpenFile(c.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	return writer.WriteAll([][]string{{strconv.Itoa(int(entry.InitTime)), strconv.Itoa(int(entry.EndTime)), entry.Category, entry.Description}})
}

func (c *CSV) Close() error {
	return nil
}

