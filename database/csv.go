package database

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"io"
)

type CSVWriter interface {
	Write(*LogEntry) error
}

type CSV struct {
	path string
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
		path: path,
	}, nil
}

func (c *CSV) Write(entry *LogEntry) error {
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(c.path, flags, 0644)
	if err != nil {
		return fmt.Errorf("opening CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if new file
	if stat, err := file.Stat(); err == nil && stat.Size() == 0 {
		headers := []string{"init_time_ms", "end_time_ms", "category", "description"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("writing CSV headers: %w", err)
		}
	}

	record := []string{
		strconv.FormatInt(entry.InitTime, 10),
		strconv.FormatInt(entry.EndTime, 10),
		entry.Category,
		entry.Description,
	}
	
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("writing CSV record: %w", err)
	}
	return nil
}

func (c *CSV) Exists(entry *LogEntry) (bool, error) {
	record, err := c.findRecord(entry)
	if err != nil {
		return false, fmt.Errorf("checking existence: %w", err)
	}

	return record != nil, nil
}

func (c *CSV) IsFinished(entry *LogEntry) (bool, error) {
	record, err := c.findRecord(entry)
	if err != nil {
		return false, fmt.Errorf("checking finished status: %w", err)
	}

	if record == nil {
		return false, nil
	}

	endTime, err := strconv.ParseInt(record[1], 10, 64)
	if err != nil {
		return false, fmt.Errorf("parsing end time: %w", err)
	}

	return endTime > 0, nil
}

func (c *CSV) Close() error {
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (c *CSV) findRecord(entry *LogEntry) ([]string, error) {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil, nil
	}

	file, err := os.OpenFile(c.path, os.O_RDONLY, 0444)
	if err != nil {
		return nil, fmt.Errorf("opening CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header row
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV record: %w", err)
		}

		initTime, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing init time: %w", err)
		}

		if initTime == entry.InitTime && record[2] == entry.Category {
			return record, nil
		}
	}

	return nil, nil
}

