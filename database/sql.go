package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type dbHandler struct {
	db *sql.DB
}

type DB interface {
	Insert(*LogEntry) error
	Exists(*LogEntry)(bool, error) 
	IsFinished(*LogEntry)(bool, error) 
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

////// POSTGRES ///////

func (l *PgHandler) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id SERIAL PRIMARY KEY,
		init_time_ms BIGINT NOT NULL,
		end_time_ms BIGINT,
		category TEXT NOT NULL,
		description TEXT,
	    constraint unique_task unique (init_time_ms, category)
	);`

	_, err := l.db.Exec(schema)
	return err
}

func (l *PgHandler) Insert(data *LogEntry) error {
	stmt := `
	INSERT INTO logs (init_time_ms, end_time_ms, category, description)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT ON CONSTRAINT unique_task 
	DO UPDATE SET end_time_ms = EXCLUDED.end_time_ms`

	_, err := l.db.Exec(stmt, data.InitTime, data.EndTime, data.Category, data.Description)
	return err
}

func (l *PgHandler) Exists(entry *LogEntry) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT 1 FROM logs
		WHERE init_time_ms = $1 AND category = $2
	)`

	var exists bool
	err := l.db.QueryRow(query, entry.InitTime, entry.Category).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking existence: %w", err)
	}
	return exists, nil
}

func (l *PgHandler) IsFinished(entry *LogEntry) (bool, error) {
	query := `
	SELECT COALESCE(end_time_ms <> 0, false)
	FROM logs
	WHERE init_time_ms = $1 AND category = $2`

	var finished bool
	err := l.db.QueryRow(query, entry.InitTime, entry.Category).Scan(&finished)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking finished status: %w", err)
	}
	return finished, nil
}

////// SQLITE ///////

func (l *SqliteHandler) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		init_time_ms BIGINT NOT NULL,
		end_time_ms BIGINT,
		category TEXT NOT NULL,
		description TEXT
	);
	CREATE UNIQUE INDEX IF NOT EXISTS unique_task ON logs(init_time_ms, category);`

	_, err := l.db.Exec(schema)
	return err
}

func (l *SqliteHandler) Insert(data *LogEntry) error {
	stmt := `
	INSERT OR REPLACE INTO logs (init_time_ms, end_time_ms, category, description)
	VALUES (?, ?, ?, ?)` 

	_, err := l.db.Exec(stmt, data.InitTime, data.EndTime, data.Category, data.Description)
	return err
}

func (l *SqliteHandler) Exists(entry *LogEntry) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT 1 FROM logs
		WHERE init_time_ms = ? AND category = ?
	)`

	var exists bool
	err := l.db.QueryRow(query, entry.InitTime, entry.Category).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking existence: %w", err)
	}
	return exists, nil
}

func (l *SqliteHandler) IsFinished(entry *LogEntry) (bool, error) {
	query := `
	SELECT COALESCE(end_time_ms <> 0, false)
	FROM logs
	WHERE init_time_ms = ? AND category = ?`

	var finished bool
	err := l.db.QueryRow(query, entry.InitTime, entry.Category).Scan(&finished)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking finished status: %w", err)
	}
	return finished, nil
}


