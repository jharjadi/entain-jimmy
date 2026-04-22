package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"git.neds.sh/matty/entain/api/proto/racing"
	"git.neds.sh/matty/entain/api/proto/sports"
)

var (
	apiEndpoint    = flag.String("api-endpoint", "localhost:8000", "API endpoint")
	grpcEndpoint   = flag.String("grpc-endpoint", "localhost:9000", "Racing gRPC server endpoint")
	sportsEndpoint = flag.String("sports-endpoint", "localhost:9001", "Sports gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running api server: %s\n", err)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := racing.RegisterRacingHandlerFromEndpoint(ctx, mux, *grpcEndpoint, dialOpts); err != nil {
		return err
	}
	if err := sports.RegisterSportsHandlerFromEndpoint(ctx, mux, *sportsEndpoint, dialOpts); err != nil {
		return err
	}

	log.Printf("API server listening on: %s\n", *apiEndpoint)

	return http.ListenAndServe(*apiEndpoint, mux)
}
