package database

import (
	"fmt"
	"just-notify/config"
	"log"
)

type Logger interface {
	Log(*LogEntry) error
	Exists(*LogEntry) (bool, error)
	IsFinished(*LogEntry) (bool, error)
	Close() error
}

type LogEntry struct {
	InitTime    int64
	EndTime     int64
	Category    string
	Description string
}

func NewLogger(conn string, database bool) (Logger, error) {
	if database {
		return openDB(conn)
	}

	return NewCSV(conn)
}

func Log(logger Logger, entry *LogEntry) error {
	return logger.Log(entry)
}

func (h *PgHandler) Log(entry *LogEntry) error {
	if err := h.Insert(entry); err != nil {
		log.Printf("Error inserting to database: %s", err)
		if err := csvFallback(entry); err != nil {
			return err
		}
	}

	return nil
}

func (h *SqliteHandler) Log(entry *LogEntry) error {
	if err := h.Insert(entry); err != nil {
		log.Printf("Error inserting to database: %s", err)
		if err := csvFallback(entry); err != nil {
			return err
		}
	}

	return nil
}

func (l *CSV) Log(entry *LogEntry) error {
	return l.Write(entry)
}

func csvFallback(entry *LogEntry) error {

	cfg := config.LoadConfig()
	args, err := config.ParseArgs(cfg)

	if err != nil {
		return err
	}

	csvHandler, err := NewCSV(args.CsvPath)

	if err != nil {
		return fmt.Errorf("Error creating csv handler: %s", err)
	}

	if err = csvHandler.Write(entry); err != nil {
		return fmt.Errorf("Error inserting data to csv: %s", err)
	}

	return nil
}
