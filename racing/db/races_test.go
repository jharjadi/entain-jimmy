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
