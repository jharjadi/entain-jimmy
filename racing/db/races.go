package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// ErrRaceNotFound is returned by Get when no race matches the requested ID.
// Callers MUST use errors.Is to distinguish this sentinel from infrastructure errors;
// the service layer maps ErrRaceNotFound -> codes.NotFound and everything else -> codes.Internal.
var ErrRaceNotFound = errors.New("race not found")

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(ctx context.Context, opts ListRacesOptions) ([]*racing.Race, error)

	// Get returns a single race by ID, or ErrRaceNotFound if it does not exist.
	Get(ctx context.Context, id int64) (*racing.Race, error)
}

// ListRacesOptions is the pre-validated, transport-agnostic query input for List.
// The service layer is responsible for validating SortBy against an allowlist; the
// repo trusts SortBy and SortDirection and concatenates them into the ORDER BY
// clause directly.
type ListRacesOptions struct {
	MeetingIDs    []int64
	VisibleOnly   bool
	SortBy        string // must be pre-validated; empty defaults to advertised_start_time
	SortDirection string // "ASC" or "DESC"; empty defaults to ASC
}

const (
	defaultSortField     = "advertised_start_time"
	defaultSortDirection = "ASC"
)

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(ctx context.Context, opts ListRacesOptions) ([]*racing.Race, error) {
	query := getRaceQueries()[racesList]
	query, args := r.applyFilter(query, opts)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	races, err := r.scanRaces(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return races, nil
}

func (r *racesRepo) Get(ctx context.Context, id int64) (*racing.Race, error) {
	rows, err := r.db.QueryContext(ctx, getRaceQueries()[racesGet], id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	races, err := r.scanRaces(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(races) == 0 {
		return nil, ErrRaceNotFound
	}
	return races[0], nil
}

func (r *racesRepo) applyFilter(query string, opts ListRacesOptions) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if len(opts.MeetingIDs) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(opts.MeetingIDs)-1)+"?)")

		for _, meetingID := range opts.MeetingIDs {
			args = append(args, meetingID)
		}
	}

	if opts.VisibleOnly {
		clauses = append(clauses, "visible = 1")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	field := defaultSortField
	if opts.SortBy != "" {
		field = opts.SortBy
	}
	direction := defaultSortDirection
	if opts.SortDirection == "DESC" {
		direction = "DESC"
	}
	query += " ORDER BY " + field + " " + direction

	return query, args
}

func (r *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = timestamppb.New(advertisedStart)
		if advertisedStart.Before(time.Now()) {
			race.Status = racing.RaceStatus_CLOSED
		}

		races = append(races, &race)
	}

	return races, nil
}
