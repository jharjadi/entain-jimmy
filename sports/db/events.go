package db

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// EventsRepo provides repository access to events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// List will return a list of events matching the given filter.
	List(ctx context.Context, filter *sports.ListEventsRequestFilter) ([]*sports.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository with seeded dummy data.
func (e *eventsRepo) Init() error {
	var err error
	e.init.Do(func() {
		err = e.seed()
	})
	return err
}

func (e *eventsRepo) List(ctx context.Context, filter *sports.ListEventsRequestFilter) ([]*sports.Event, error) {
	query := getEventQueries()[eventsList]
	query, args := applyFilter(query, filter)

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events, err := scanEvents(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if filter.VisibleOnly {
		clauses = append(clauses, "visible = 1")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	return query, args
}

func scanEvents(rows *sql.Rows) ([]*sports.Event, error) {
	var events []*sports.Event
	for rows.Next() {
		var event sports.Event
		var advertisedStart time.Time

		if err := rows.Scan(&event.Id, &event.Name, &event.Visible, &advertisedStart); err != nil {
			return nil, err
		}
		event.AdvertisedStartTime = timestamppb.New(advertisedStart)

		events = append(events, &event)
	}
	return events, nil
}
