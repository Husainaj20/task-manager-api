# Task Manager API - Testing Guide

## Day 2 Testing Guide

### Run the entire suite (with race detector and coverage)
```bash
go test ./... -race -cover
```

### Run only the API handler tests
```bash
go test ./internal/api -race -v
```

### Expected output (approx):
```
=== RUN   TestHealthz
--- PASS: TestHealthz (0.00s)
=== RUN   TestCreateTask_Success
--- PASS: TestCreateTask_Success (0.00s)
=== RUN   TestCreateTask_ValidationError
--- PASS: TestCreateTask_ValidationError (0.00s)
=== RUN   TestCreateTask_Idempotency
--- PASS: TestCreateTask_Idempotency (0.00s)
=== RUN   TestGetTask_Success
--- PASS: TestGetTask_Success (0.00s)
=== RUN   TestGetTask_NotFound
--- PASS: TestGetTask_NotFound (0.00s)
PASS
ok  	github.com/husainaj20/task-manager-api/internal/api	0.XXXs
coverage: ~80%
```

## Expected Test Results Summary

### 1. POST /tasks (happy path)
- **Status**: 202 Accepted
- **Body**: JSON with non-empty "id", type="echo", status="queued" or "done"

### 2. POST /tasks (same Idempotency-Key)
- **Status**: 202 Accepted
- **Body**: same "id" as first request

### 3. GET /tasks/:id (existing)
- **Status**: 200 OK
- **Body**: task with correct id, type, payload

### 4. GET /tasks/:id (missing)
- **Status**: 404 Not Found
- **Body**: `{"error":"not found"}`

### 5. All tests pass (exit code 0)
- "ok" printed for internal/api with no FAIL lines

## Manual Validation (Optional)

### Start the server
```bash
make run
```

### Create a task
```bash
curl -s -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: abc123' \
  -d '{"type":"echo","payload":{"msg":"hello"}}' | jq .
```

### Recreate same task (same key)
```bash
curl -s -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: abc123' \
  -d '{"type":"echo","payload":{"msg":"hello"}}' | jq .
```

Both should return identical IDs.

### Fetch by ID
```bash
curl -s localhost:8080/tasks/<TASK_ID> | jq .
```

Should return 200 with matching data.

## Test Coverage Guidelines

- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test API endpoints end-to-end
- **Race Detection**: Always run with `-race` flag to catch concurrency issues
- **Idempotency**: Verify that duplicate requests with same key return same result
- **Error Cases**: Test validation failures and not-found scenarios