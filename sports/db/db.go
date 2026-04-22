package db

import (
	"time"

	"syreclabs.com/go/faker"
)

// seed populates the events table with 50 dummy rows for local dev and demo.
// Idempotent via INSERT OR IGNORE.
func (e *eventsRepo) seed() error {
	statement, err := e.db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, name TEXT, visible INTEGER, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 50; i++ {
		statement, err = e.db.Prepare(`INSERT OR IGNORE INTO events(id, name, visible, advertised_start_time) VALUES (?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Team().Name()+" vs "+faker.Team().Name(),
				faker.Number().Between(0, 1),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			)
		}
	}

	return err
}
