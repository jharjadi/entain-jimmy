package main

import (
	"database/sql"
	"flag"
	"log"
	"net"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"git.neds.sh/matty/entain/sports/service"
)

var (
	grpcEndpoint = flag.String("grpc-endpoint", ":9001", "gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc server: %s\n", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", *grpcEndpoint)
	if err != nil {
		return err
	}

	sportsDB, err := sql.Open("sqlite3", "./db/sports.db")
	if err != nil {
		return err
	}

	eventsRepo := db.NewEventsRepo(sportsDB)
	if err := eventsRepo.Init(); err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	sports.RegisterSportsServer(grpcServer, service.NewSportsService(eventsRepo))

	log.Printf("gRPC server listening on: %s\n", *grpcEndpoint)
	return grpcServer.Serve(conn)
}
