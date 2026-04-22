package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
)

// allowedSortFields is the allowlist of sort_by values accepted from ListRaces callers.
// SQL drivers cannot parameter-bind column identifiers, so ORDER BY values get
// concatenated directly into the query — the allowlist prevents SQL injection.
var allowedSortFields = map[string]struct{}{
	"id":                    {},
	"meeting_id":            {},
	"name":                  {},
	"number":                {},
	"advertised_start_time": {},
}

type Racing interface {
	// ListRaces will return a collection of races.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
}

// racingService implements the Racing interface.
type racingService struct {
	racesRepo db.RacesRepo
}

// NewRacingService instantiates and returns a new racingService.
func NewRacingService(racesRepo db.RacesRepo) Racing {
	return &racingService{racesRepo}
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	opts, err := listRacesOptionsFromRequest(in)
	if err != nil {
		return nil, err
	}

	races, err := s.racesRepo.List(ctx, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &racing.ListRacesResponse{Races: races}, nil
}

// listRacesOptionsFromRequest validates the request and translates it into the
// transport-agnostic ListRacesOptions that the repo consumes.
func listRacesOptionsFromRequest(in *racing.ListRacesRequest) (db.ListRacesOptions, error) {
	var opts db.ListRacesOptions
	if in.Filter == nil {
		return opts, nil
	}

	if in.Filter.SortBy != "" {
		if _, ok := allowedSortFields[in.Filter.SortBy]; !ok {
			return opts, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid sort_by %q", in.Filter.SortBy))
		}
	}

	opts.MeetingIDs = in.Filter.MeetingIds
	opts.VisibleOnly = in.Filter.VisibleOnly
	opts.SortBy = in.Filter.SortBy
	if in.Filter.SortDirection == racing.SortDirection_DESC {
		opts.SortDirection = "DESC"
	}
	return opts, nil
}
