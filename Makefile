.PHONY: build run run-racing run-sports run-api proto proto-racing proto-sports proto-api test lint clean kill install-tools

export PATH := $(PATH):$(shell go env GOPATH)/bin

install-tools:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

build:
	cd racing && go build -o racing .
	cd sports && go build -o sports .
	cd api && go build -o api .

run-racing:
	cd racing && go run main.go &

run-sports:
	cd sports && go run main.go &

run-api:
	cd api && go run main.go &

run: run-racing run-sports run-api

proto-racing:
	cd racing && go generate ./proto/...

proto-sports:
	cd sports && go generate ./proto/...

proto-api:
	cd api && go generate ./proto/...

proto: proto-racing proto-sports proto-api

test:
	cd racing && go test ./... -race -cover
	cd sports && go test ./... -race -cover
	cd api && go test ./... -race -cover

lint:
	cd racing && go vet ./...
	cd sports && go vet ./...
	cd api && go vet ./...

clean:
	rm -f racing/racing sports/sports api/api

kill:
	-lsof -ti :9000 | xargs kill -9 2>/dev/null
	-lsof -ti :9001 | xargs kill -9 2>/dev/null
	-lsof -ti :8000 | xargs kill -9 2>/dev/null
