.PHONY: build run test fmt clean

build:
	go build -o bin/api ./cmd/main.go

run:
	go run ./cmd/main.go

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	go clean ./...
