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

smoke:
	@echo "Running local smoke tests against http://localhost:8080"
	@echo "GET /healthz"
	@curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:8080/healthz || true

	@echo "POST /tasks (create)"
	@curl -s -X POST http://localhost:8080/tasks \
		-H 'Content-Type: application/json' \
		-H 'Idempotency-Key: local-smoke-1' \
		-d '{"type":"echo","payload":{"msg":"local-smoke"}}' | jq || true

	@echo "POST /tasks (idempotency repeat)"
	@curl -s -X POST http://localhost:8080/tasks \
		-H 'Content-Type: application/json' \
		-H 'Idempotency-Key: local-smoke-1' \
		-d '{"type":"echo","payload":{"msg":"local-smoke"}}' | jq || true
