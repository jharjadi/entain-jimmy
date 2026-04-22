package db

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// newTestDB returns an in-memory SQLite with the races schema ready for inserts.
// Shared by tests across this package.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE races (
		id INTEGER PRIMARY KEY,
		meeting_id INTEGER,
		name TEXT,
		number INTEGER,
		visible INTEGER,
		advertised_start_time DATETIME
	)`)
	require.NoError(t, err)
	return db
}

// insertRace inserts a single race row. Shared by tests across this package.
func insertRace(t *testing.T, db *sql.DB, id int64, visible int, start time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO races (id, meeting_id, name, number, visible, advertised_start_time) VALUES (?, ?, ?, ?, ?, ?)`,
		id, 1, fmt.Sprintf("Race %d", id), 1, visible, start.Format(time.RFC3339),
	)
	require.NoError(t, err)
}
