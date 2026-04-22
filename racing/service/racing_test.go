package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
)

// mockRacesRepo is a hand-rolled test double for db.RacesRepo. Records the last
// ListRacesOptions it received so tests can assert on service → repo translation.
type mockRacesRepo struct {
	lastOpts db.ListRacesOptions
	races    []*racing.Race
	listErr  error
}

func (m *mockRacesRepo) Init() error { return nil }

func (m *mockRacesRepo) List(ctx context.Context, opts db.ListRacesOptions) ([]*racing.Race, error) {
	m.lastOpts = opts
	return m.races, m.listErr
}

func TestListRaces_PassesFilterFieldsThroughToRepo(t *testing.T) {
	repo := &mockRacesRepo{races: []*racing.Race{{Id: 1}}}
	svc := NewRacingService(repo)

	_, err := svc.ListRaces(context.Background(), &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			MeetingIds:     []int64{7, 8},
			VisibleOnly:    true,
			SortBy:         "id",
			SortDirection:  racing.SortDirection_DESC,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []int64{7, 8}, repo.lastOpts.MeetingIDs)
	assert.True(t, repo.lastOpts.VisibleOnly)
	assert.Equal(t, "id", repo.lastOpts.SortBy)
	assert.Equal(t, "DESC", repo.lastOpts.SortDirection)
}

func TestListRaces_RejectsInvalidSortBy(t *testing.T) {
	// SQL-injection attempt should never reach the repo.
	repo := &mockRacesRepo{}
	svc := NewRacingService(repo)

	_, err := svc.ListRaces(context.Background(), &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{SortBy: "DROP TABLE races"},
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "expected a gRPC status error")
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, db.ListRacesOptions{}, repo.lastOpts, "repo must not be called on invalid input")
}

func TestListRaces_WrapsRepoErrorAsInternal(t *testing.T) {
	repo := &mockRacesRepo{listErr: errors.New("connection refused")}
	svc := NewRacingService(repo)

	_, err := svc.ListRaces(context.Background(), &racing.ListRacesRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}
