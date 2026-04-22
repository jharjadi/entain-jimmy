package service_test

import (
	"context"
	"database/sql"
	"net"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"git.neds.sh/matty/entain/racing/service"
)

// newBufconnServer boots a real gRPC server over an in-process bufconn pipe
// with a real service + real repo + in-memory SQLite. Returns a dialled client
// and the underlying *sql.DB so tests can seed rows.
//
// This verifies end-to-end that gRPC status codes survive the wire: a unit test
// would prove service.GetRace returns codes.NotFound in Go, but couldn't prove
// the code round-trips through gRPC serialisation without loss.
func newBufconnServer(t *testing.T) (racing.RacingClient, *sql.DB) {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	_, err = sqlDB.Exec(`CREATE TABLE races (
		id INTEGER PRIMARY KEY,
		meeting_id INTEGER,
		name TEXT,
		number INTEGER,
		visible INTEGER,
		advertised_start_time DATETIME
	)`)
	require.NoError(t, err)

	repo := db.NewRacesRepo(sqlDB)
	svc := service.NewRacingService(repo)

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	racing.RegisterRacingServer(grpcServer, svc)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return racing.NewRacingClient(conn), sqlDB
}

func TestIntegration_GetRace_MissingIDReturnsNotFound(t *testing.T) {
	client, _ := newBufconnServer(t)

	_, err := client.GetRace(context.Background(), &racing.GetRaceRequest{Id: 99999})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "expected a gRPC status error")
	assert.Equal(t, codes.NotFound, st.Code(), "missing race must be codes.NotFound, not Unknown/Internal")
}

func TestIntegration_GetRace_FoundReturnsRace(t *testing.T) {
	client, sqlDB := newBufconnServer(t)

	_, err := sqlDB.Exec(
		`INSERT INTO races (id, meeting_id, name, number, visible, advertised_start_time) VALUES (?, ?, ?, ?, ?, ?)`,
		42, 1, "Test Race", 1, 1, "2099-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	got, err := client.GetRace(context.Background(), &racing.GetRaceRequest{Id: 42})
	require.NoError(t, err)
	assert.Equal(t, int64(42), got.Id)
	assert.Equal(t, "Test Race", got.Name)
}
