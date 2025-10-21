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
