.PHONY: build run test fmt lint clean

build:
	go build ./...

run:
	go run ./cmd/cogniqd

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	go clean