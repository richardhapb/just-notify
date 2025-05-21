package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
	"strings"
)

type LogDB struct {
	db     *sql.DB
	driver string
}


func OpenDB(dsn string) (*LogDB, error) {
	var driver, connStr string

	switch {
	case strings.HasPrefix(dsn, "sqlite://"):
		driver = "sqlite3"
		connStr = strings.TrimPrefix(dsn, "sqlite://")
	case strings.HasPrefix(dsn, "postgresql://"):
		driver = "postgres"
		connStr = dsn
	default:
		return nil, fmt.Errorf("unsupported driver in DSN: %s", dsn)
	}

	conn, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}

	return &LogDB{conn, driver}, nil
}

func (l *LogDB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id SERIAL PRIMARY KEY,
		init_time_ms BIGINT NOT NULL,
		end_time_ms BIGINT NOT NULL,
		task_name TEXT NOT NULL
	);`

	_, err := l.db.Exec(schema)
	return err
}

func (l *LogDB) Insert(data *LogEntry) error {
	stmt := ""
	switch l.driver {
	case "sqlite3":
		stmt = `
		INSERT INTO logs (init_time_ms, end_time_ms, task_name)
		VALUES (?, ?, ?)`
	case "postgres":
		stmt = `
			INSERT INTO logs (init_time_ms, end_time_ms, task_name)
			VALUES ($1, $2, $3)`
	default:
		return fmt.Errorf("unknown driver in connection %v", l)
	}

	_, err := l.db.Exec(stmt, data.InitTime, data.EndTime, data.TaskName)
	return err
}

func (l *LogDB) Close() error {
	return l.db.Close()
}
