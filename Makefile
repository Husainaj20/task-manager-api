SHELL := /bin/bash

.PHONY: run test bench tidy
.PHONY: run test bench tidy build docker-build compose-up compose-down


run:
	go run ./cmd/server

test:
	go test ./... -race

bench:
	go test -bench . -benchmem ./...

tidy:

build:
	go build ./cmd/server

docker-build:
	docker build -t task-manager-api:local .

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down -v
	go mod tidy
