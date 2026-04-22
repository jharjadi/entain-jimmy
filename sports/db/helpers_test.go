package db

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// newTestDB returns an in-memory SQLite with the events schema ready for inserts.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE events (
		id INTEGER PRIMARY KEY,
		name TEXT,
		visible INTEGER,
		advertised_start_time DATETIME
	)`)
	require.NoError(t, err)
	return db
}

// insertEvent inserts a single event row.
func insertEvent(t *testing.T, db *sql.DB, id int64, visible int, start time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO events (id, name, visible, advertised_start_time) VALUES (?, ?, ?, ?)`,
		id, fmt.Sprintf("Event %d", id), visible, start.Format(time.RFC3339),
	)
	require.NoError(t, err)
}
