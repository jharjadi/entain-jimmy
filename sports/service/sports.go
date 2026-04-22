package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
)

// Sports defines the sports gRPC service surface.
type Sports interface {
	// ListEvents will return a collection of events.
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
}

// sportsService implements the Sports interface.
type sportsService struct {
	eventsRepo db.EventsRepo
}

// NewSportsService instantiates and returns a new sportsService.
func NewSportsService(r db.EventsRepo) Sports {
	return &sportsService{eventsRepo: r}
}

func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	events, err := s.eventsRepo.List(ctx, in.Filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &sports.ListEventsResponse{Events: events}, nil
}
