# Task Manager API (Go)

Day 1 scaffold to demonstrate Go backend fundamentals (REST + concurrency).

## Quick start
```bash
make test
make run
# server on :8080
```

## Endpoints
- `GET /healthz`
- `GET /readiness`
- `POST /tasks` (Idempotency-Key supported)
- `GET /tasks/:id`
