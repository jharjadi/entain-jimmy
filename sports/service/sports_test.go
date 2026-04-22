package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// mockEventsRepo is a hand-rolled test double for db.EventsRepo.
type mockEventsRepo struct {
	lastFilter *sports.ListEventsRequestFilter
	events     []*sports.Event
	listErr    error
}

func (m *mockEventsRepo) Init() error { return nil }

func (m *mockEventsRepo) List(ctx context.Context, filter *sports.ListEventsRequestFilter) ([]*sports.Event, error) {
	m.lastFilter = filter
	return m.events, m.listErr
}

func TestListEvents_PassesFilterThroughToRepo(t *testing.T) {
	repo := &mockEventsRepo{events: []*sports.Event{{Id: 1}}}
	svc := NewSportsService(repo)

	_, err := svc.ListEvents(context.Background(), &sports.ListEventsRequest{
		Filter: &sports.ListEventsRequestFilter{VisibleOnly: true},
	})
	require.NoError(t, err)
	require.NotNil(t, repo.lastFilter)
	assert.True(t, repo.lastFilter.VisibleOnly)
}

func TestListEvents_WrapsRepoErrorAsInternal(t *testing.T) {
	repo := &mockEventsRepo{listErr: errors.New("connection refused")}
	svc := NewSportsService(repo)

	_, err := svc.ListEvents(context.Background(), &sports.ListEventsRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}
