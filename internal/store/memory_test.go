package store

import (
    "context"
    "sync"
    "testing"

    "github.com/husainaj20/task-manager-api/internal/models"
)

func TestMemoryStore_CreateOrGetByKey_Concurrent(t *testing.T) {
    ms := NewMemoryStore()
    key := "concurrent-key"
    var wg sync.WaitGroup
    ids := make([]string, 10)
    wg.Add(10)
    for i := 0; i < 10; i++ {
        idx := i
        go func() {
            defer wg.Done()
            task := &models.Task{Type: "echo", Payload: map[string]any{"i": idx}}
            got, _, _ := ms.CreateOrGetByKey(context.Background(), key, task)
            ids[idx] = got.ID
        }()
    }
    wg.Wait()
    // All ids should be the same non-empty string
    if ids[0] == "" {
        t.Fatalf("expected non-empty id")
    }
    for i := 1; i < len(ids); i++ {
        if ids[i] != ids[0] {
            t.Fatalf("expected same id for all concurrent creates, got %s and %s", ids[0], ids[i])
        }
    }
}

func TestMemoryStore_UpdateStatus(t *testing.T) {
    ms := NewMemoryStore()
    task := &models.Task{Type: "echo", Payload: map[string]any{"x": 1}}
    created, _, _ := ms.CreateOrGetByKey(context.Background(), "k1", task)
    err := ms.UpdateStatus(context.Background(), created.ID, "done", map[string]any{"echo": map[string]any{"x": 1}})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    fetched, err := ms.Get(context.Background(), created.ID)
    if err != nil {
        t.Fatalf("get failed: %v", err)
    }
    if fetched.Status != "done" {
        t.Fatalf("expected status done, got %s", fetched.Status)
    }
    if fetched.Result == nil {
        t.Fatalf("expected result to be set")
    }
}
