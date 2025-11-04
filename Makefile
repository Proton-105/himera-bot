.PHONY: all build run test lint

all: build

build:
	go build -o bin/himera-bot ./cmd/bot

run:
	go run ./cmd/bot

test:
	go test ./...

lint:
	golangci-lint run ./...
