package database

type Logger interface {
	Log(*LogEntry) error
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
		return err
	}

	return nil
}

func (h *SqliteHandler) Log(entry *LogEntry) error {
	if err := h.Insert(entry); err != nil {
		return err
	}

	return nil
}

func (l *CSV) Log(entry *LogEntry) error {
	return l.Write(entry)
}
