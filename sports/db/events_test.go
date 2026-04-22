package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

func TestList_VisibleOnlyReturnsOnlyVisible(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	insertEvent(t, db, 1, 1, now.Add(time.Hour))
	insertEvent(t, db, 2, 0, now.Add(time.Hour))
	insertEvent(t, db, 3, 1, now.Add(time.Hour))

	repo := &eventsRepo{db: db}

	all, err := repo.List(context.Background(), &sports.ListEventsRequestFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 3, "no filter should return all events")

	visible, err := repo.List(context.Background(), &sports.ListEventsRequestFilter{VisibleOnly: true})
	require.NoError(t, err)
	assert.Len(t, visible, 2, "visible_only=true should return only visible events")
	for _, e := range visible {
		assert.True(t, e.Visible)
	}
}

func TestList_NilFilterReturnsAll(t *testing.T) {
	db := newTestDB(t)
	insertEvent(t, db, 1, 1, time.Now().Add(time.Hour))
	insertEvent(t, db, 2, 0, time.Now().Add(time.Hour))

	repo := &eventsRepo{db: db}
	events, err := repo.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, events, 2)
}
