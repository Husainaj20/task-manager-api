SHELL := /bin/bash

.PHONY: run test bench tidy

run:
	go run ./cmd/server

test:
	go test ./... -race

bench:
	go test -bench . -benchmem ./...

tidy:
	go mod tidy
