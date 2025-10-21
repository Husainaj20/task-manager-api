package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/husainaj20/task-manager-api/internal/service"
	"github.com/husainaj20/task-manager-api/internal/store"
)

func TestHealthz(t *testing.T) {
	mem := store.NewMemoryStore()
	q := service.NewQueue(1)
	defer q.Stop()
	h := New(mem, q)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateTask_Success(t *testing.T) {
	mem := store.NewMemoryStore()
	q := service.NewQueue(1)
	defer q.Stop()
	h := New(mem, q)

	payload := map[string]interface{}{
		"type":    "echo",
		"payload": map[string]interface{}{"msg": "hello"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-123")
	
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	// Verify response contains task data
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	
	if result["id"] == "" {
		t.Error("expected task ID in response")
	}
	if result["status"] != "queued" {
		t.Errorf("expected status 'queued', got %v", result["status"])
	}
	if result["type"] != "echo" {
		t.Errorf("expected type 'echo', got %v", result["type"])
	}
}

func TestCreateTask_ValidationError(t *testing.T) {
	mem := store.NewMemoryStore()
	q := service.NewQueue(1)
	defer q.Stop()
	h := New(mem, q)

	// Missing required 'type' field
	payload := map[string]interface{}{
		"payload": map[string]interface{}{"msg": "hello"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	// Check error message
	if !strings.Contains(rec.Body.String(), "error") {
		t.Error("expected error message in response")
	}
}

func TestCreateTask_Idempotency(t *testing.T) {
	mem := store.NewMemoryStore()
	q := service.NewQueue(1)
	defer q.Stop()
	h := New(mem, q)

	payload := map[string]interface{}{
		"type":    "echo",
		"payload": map[string]interface{}{"msg": "hello"},
	}
	body, _ := json.Marshal(payload)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "duplicate-test")
	
	rec1 := httptest.NewRecorder()
	h.Router().ServeHTTP(rec1, req1)

	// Second request with same idempotency key
	body2, _ := json.Marshal(payload) // Fresh body
	req2 := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "duplicate-test")
	
	rec2 := httptest.NewRecorder()
	h.Router().ServeHTTP(rec2, req2)

	// Both should return 202
	if rec1.Code != http.StatusAccepted || rec2.Code != http.StatusAccepted {
		t.Fatalf("expected both requests to return 202, got %d and %d", rec1.Code, rec2.Code)
	}

	// Should return the same task
	var task1, task2 map[string]interface{}
	json.Unmarshal(rec1.Body.Bytes(), &task1)
	json.Unmarshal(rec2.Body.Bytes(), &task2)

	if task1["id"] != task2["id"] {
		t.Errorf("expected same task ID, got %v and %v", task1["id"], task2["id"])
	}
}

func TestGetTask_Success(t *testing.T) {
	mem := store.NewMemoryStore()
	q := service.NewQueue(1)
	defer q.Stop()
	h := New(mem, q)

	// First create a task
	payload := map[string]interface{}{
		"type":    "echo",
		"payload": map[string]interface{}{"msg": "hello"},
	}
	body, _ := json.Marshal(payload)

	createReq := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Idempotency-Key", "get-test")
	
	createRec := httptest.NewRecorder()
	h.Router().ServeHTTP(createRec, createReq)

	var createdTask map[string]interface{}
	json.Unmarshal(createRec.Body.Bytes(), &createdTask)
	taskID := createdTask["id"].(string)

	// Now get the task
	getReq := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	getRec := httptest.NewRecorder()
	h.Router().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}

	var retrievedTask map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &retrievedTask)

	if retrievedTask["id"] != taskID {
		t.Errorf("expected task ID %s, got %v", taskID, retrievedTask["id"])
	}
}

func TestGetTask_NotFound(t *testing.T) {
	mem := store.NewMemoryStore()
	q := service.NewQueue(1)
	defer q.Stop()
	h := New(mem, q)

	req := httptest.NewRequest(http.MethodGet, "/tasks/nonexistent-id", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "not found") {
		t.Error("expected 'not found' message in response")
	}
}
