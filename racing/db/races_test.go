package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// newTestDB returns an in-memory SQLite with the races schema ready for inserts.
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

func insertRace(t *testing.T, db *sql.DB, id int64, visible int, start time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO races (id, meeting_id, name, number, visible, advertised_start_time) VALUES (?, ?, ?, ?, ?, ?)`,
		id, 1, fmt.Sprintf("Race %d", id), 1, visible, start.Format(time.RFC3339),
	)
	require.NoError(t, err)
}

func TestList_VisibleOnlyReturnsOnlyVisible(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	insertRace(t, db, 1, 1, now.Add(time.Hour))
	insertRace(t, db, 2, 0, now.Add(time.Hour))
	insertRace(t, db, 3, 1, now.Add(time.Hour))

	repo := &racesRepo{db: db}

	all, err := repo.List(context.Background(), &racing.ListRacesRequestFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 3, "no filter should return all races")

	visible, err := repo.List(context.Background(), &racing.ListRacesRequestFilter{VisibleOnly: true})
	require.NoError(t, err)
	assert.Len(t, visible, 2, "visible_only=true should return only visible races")
	for _, r := range visible {
		assert.True(t, r.Visible)
	}
}
