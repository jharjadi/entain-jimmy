package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

func TestList_VisibleOnlyReturnsOnlyVisible(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	insertRace(t, db, 1, 1, now.Add(time.Hour))
	insertRace(t, db, 2, 0, now.Add(time.Hour))
	insertRace(t, db, 3, 1, now.Add(time.Hour))

	repo := &racesRepo{db: db}

	all, err := repo.List(context.Background(), ListRacesOptions{})
	require.NoError(t, err)
	assert.Len(t, all, 3, "no filter should return all races")

	visible, err := repo.List(context.Background(), ListRacesOptions{VisibleOnly: true})
	require.NoError(t, err)
	assert.Len(t, visible, 2, "visible_only=true should return only visible races")
	for _, r := range visible {
		assert.True(t, r.Visible)
	}
}

func TestList_DefaultsToAdvertisedStartTimeAsc(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	insertRace(t, db, 1, 1, now.Add(2*time.Hour))
	insertRace(t, db, 2, 1, now.Add(1*time.Hour))
	insertRace(t, db, 3, 1, now.Add(3*time.Hour))

	repo := &racesRepo{db: db}
	races, err := repo.List(context.Background(), ListRacesOptions{})
	require.NoError(t, err)
	require.Len(t, races, 3)
	assert.Equal(t, int64(2), races[0].Id, "earliest start should come first")
	assert.Equal(t, int64(1), races[1].Id)
	assert.Equal(t, int64(3), races[2].Id, "latest start should come last")
}

func TestList_DerivesStatusFromStartTime(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	insertRace(t, db, 1, 1, now.Add(-time.Hour)) // past -> CLOSED
	insertRace(t, db, 2, 1, now.Add(time.Hour))  // future -> OPEN

	repo := &racesRepo{db: db}
	races, err := repo.List(context.Background(), ListRacesOptions{})
	require.NoError(t, err)
	require.Len(t, races, 2)

	// Default order is advertised_start_time ASC, so the past race is first.
	assert.Equal(t, int64(1), races[0].Id)
	assert.Equal(t, racing.RaceStatus_CLOSED, races[0].Status)
	assert.Equal(t, int64(2), races[1].Id)
	assert.Equal(t, racing.RaceStatus_OPEN, races[1].Status)
}

func TestList_CustomSortByIdDesc(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	insertRace(t, db, 1, 1, now.Add(time.Hour))
	insertRace(t, db, 2, 1, now.Add(time.Hour))
	insertRace(t, db, 3, 1, now.Add(time.Hour))

	repo := &racesRepo{db: db}
	races, err := repo.List(context.Background(), ListRacesOptions{SortBy: "id", SortDirection: "DESC"})
	require.NoError(t, err)
	require.Len(t, races, 3)
	assert.Equal(t, int64(3), races[0].Id)
	assert.Equal(t, int64(2), races[1].Id)
	assert.Equal(t, int64(1), races[2].Id)
}
