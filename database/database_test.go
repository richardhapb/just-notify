package database

import (
	"database/sql"
	"os"
	"testing"
	"time"
	"sync"
	"fmt"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		conn        string
		isDatabase  bool
		shouldError bool
	}{
		{"SQLite Valid", "sqlite://test.db", true, false},
		{"PostgreSQL Invalid", "postgresql://invalid", true, true},
		{"CSV Default Path", "", false, false},
		{"CSV Custom Path", "/tmp/test.csv", false, false},
		{"Invalid DSN", "invalid://test", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.conn, tt.isDatabase)
			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && logger == nil {
				t.Error("logger is nil but no error returned")
			}

			// Cleanup SQLite database file
			if tt.conn == "sqlite://test.db" {
				os.Remove("test.db")
			}
		})
	}
}

func TestCSVLogger(t *testing.T) {
	tempFile := "test.csv"
	defer os.Remove(tempFile)

	logger, err := NewLogger(tempFile, false)
	if err != nil {
		t.Fatalf("failed to create CSV logger: %v", err)
	}

	entry := &LogEntry{
		InitTime:    time.Now().UnixMilli(),
		EndTime:     time.Now().Add(time.Second).UnixMilli(),
		Category:    "test",
		Description: "test description",
	}

	if err := logger.Log(entry); err != nil {
		t.Errorf("failed to log entry: %v", err)
	}

	// Verify file exists and is not empty
	info, err := os.Stat(tempFile)
	if err != nil {
		t.Errorf("failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("CSV file is empty")
	}
}

func TestSQLiteLogger(t *testing.T) {
	t.Run("creates schema and logs entries", func(t *testing.T) {
		dbFile := "test.db"
		defer os.Remove(dbFile)

		logger, err := NewLogger("sqlite://"+dbFile, true)
		if err != nil {
			t.Fatalf("failed to create SQLite logger: %v", err)
		}
		defer logger.Close()

		want := &LogEntry{
			InitTime:    time.Now().UnixMilli(),
			EndTime:     time.Now().Add(time.Second).UnixMilli(),
			Category:    "test",
			Description: "test description",
		}

		if err := logger.Log(want); err != nil {
			t.Fatalf("failed to log entry: %v", err)
		}

		// Open database to verify contents
		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		// Query the logged entry
		var got LogEntry
		err = db.QueryRow(`
			SELECT init_time_ms, end_time_ms, category, description 
			FROM logs 
			ORDER BY init_time_ms DESC 
			LIMIT 1
		`).Scan(&got.InitTime, &got.EndTime, &got.Category, &got.Description)
		if err != nil {
			t.Fatalf("failed to query log entry: %v", err)
		}

		// Compare logged data
		if got.Category != want.Category {
			t.Errorf("category = %q, want %q", got.Category, want.Category)
		}
		if got.Description != want.Description {
			t.Errorf("description = %q, want %q", got.Description, want.Description)
		}
		if got.InitTime != want.InitTime {
			t.Errorf("init_time_ms = %d, want %d", got.InitTime, want.InitTime)
		}
		if got.EndTime != want.EndTime {
			t.Errorf("end_time_ms = %d, want %d", got.EndTime, want.EndTime)
		}
	})
}

func TestConcurrentLogging(t *testing.T) {
	tempFile := "concurrent_test.csv"
	defer os.Remove(tempFile)

	logger, err := NewLogger(tempFile, false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	// Start goroutines
	for i := range numGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			entry := &LogEntry{
				InitTime:    time.Now().UnixMilli(),
				EndTime:     time.Now().Add(time.Second).UnixMilli(),
				Category:    fmt.Sprintf("concurrent-%d", i),
				Description: "test concurrent logging",
			}
			if err := logger.Log(entry); err != nil {
				errCh <- fmt.Errorf("goroutine %d: %w", i, err)
			}
		}(i)
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success case
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for concurrent operations")
	}

	// Check for any errors
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent logging error: %v", err)
	}
}
