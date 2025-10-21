# Task Manager API (Go)

A progressive 14-day implementation demonstrating Go backend fundamentals (REST + concurrency).

## Quick start

```bash
make test
make run
# server on :8080
```

## Endpoints

- `GET /healthz` - Health check
- `GET /readiness` - Readiness check
- `POST /tasks` - Create task (Idempotency-Key supported)
- `GET /tasks/:id` - Get task by ID

## Day 2 — Task API Examples

### Create a task (idempotent if repeated with same key):

```bash
curl -s -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: abc123' \
  -d '{"type":"echo","payload":{"msg":"hello"}}'
```

### Fetch task by ID:

```bash
curl -s localhost:8080/tasks/<TASK_ID>
```

### Example workflow:

```bash
# Create a task
RESPONSE=$(curl -s -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: test-001' \
  -d '{"type":"processing","payload":{"data":"example"}}')

# Extract task ID and fetch it
TASK_ID=$(echo $RESPONSE | jq -r '.id')
curl -s localhost:8080/tasks/$TASK_ID | jq
```

## Running with Docker Compose

Start API + Redis:

```bash
docker compose up --build
```

Stop and remove:

```bash
docker compose down
```

The API will be reachable at http://localhost:8080 and will use Redis as the backing store when `STORE=redis` is set by the compose file.

## CI

This repository includes a GitHub Actions workflow at `.github/workflows/ci.yml` which runs on pushes and pull requests to `main` and performs:

- `go vet ./...`
- a `gofmt` formatting check
- `go test ./... -race -v`

Ensure your local changes pass the same checks before pushing.

Additionally the workflow includes a `docker-smoke` job that builds the Docker image using Buildx and runs a short containerized smoke test (health, create/idempotency, fetch). This verifies the built image runs in CI similarly to local Docker.
