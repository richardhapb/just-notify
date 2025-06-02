package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

type dbHandler struct {
	db *sql.DB
}

type DB interface {
	Insert(*LogEntry) error
	initSchema() error
	Close() error
	Logger
}

func (h *dbHandler) Close() error {
	return h.db.Close()
}

type PgHandler struct {
	dbHandler
}

type SqliteHandler struct {
	dbHandler
}

func openDB(dsn string) (DB, error) {
	var handler, connStr string

	switch {
	case strings.HasPrefix(dsn, "sqlite://"):
		handler = "sqlite3"
		connStr = strings.TrimPrefix(dsn, "sqlite://")
	case strings.HasPrefix(dsn, "postgresql://") || strings.HasPrefix(dsn, "postgres://"):
		handler = "postgres"
		connStr = dsn
	default:
		return nil, fmt.Errorf("unsupported driver in DSN: %s", dsn)
	}

	conn, err := sql.Open(handler, connStr)
	if err != nil {
		return nil, err
	}

	var db DB
	switch handler {
	case "sqlite3":
		db = &SqliteHandler{dbHandler: dbHandler{db: conn}}
	case "postgres":
		db = &PgHandler{dbHandler: dbHandler{db: conn}}
	}

	return db, db.initSchema()
}

func (l *PgHandler) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id SERIAL PRIMARY KEY,
		init_time_ms BIGINT NOT NULL,
		end_time_ms BIGINT NOT NULL,
		category TEXT NOT NULL,
		description TEXT
	);`

	_, err := l.db.Exec(schema)
	return err
}

func (l *SqliteHandler) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		init_time_ms BIGINT NOT NULL,
		end_time_ms BIGINT NOT NULL,
		category TEXT NOT NULL,
		description TEXT
	);`

	_, err := l.db.Exec(schema)
	return err
}

func (l *PgHandler) Insert(data *LogEntry) error {
	stmt := `
			INSERT INTO logs (init_time_ms, end_time_ms, category, description)
			VALUES ($1, $2, $3, $4)`

	_, err := l.db.Exec(stmt, data.InitTime, data.EndTime, data.Category, data.Description)
	return err
}

func (l *SqliteHandler) Insert(data *LogEntry) error {
	stmt := `
	INSERT INTO logs (init_time_ms, end_time_ms, category, description)
	VALUES (?, ?, ?, ?)`

	_, err := l.db.Exec(stmt, data.InitTime, data.EndTime, data.Category, data.Description)
	return err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

