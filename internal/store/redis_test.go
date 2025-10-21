package store

import (
    "context"
    "testing"

    miniredis "github.com/alicebob/miniredis/v2"
    "github.com/husainaj20/task-manager-api/internal/models"
)

func TestRedisStore_CreateGetUpdate_Idempotent(t *testing.T) {
    mr, err := miniredis.Run()
    if err != nil {
        t.Fatalf("failed to start miniredis: %v", err)
    }
    defer mr.Close()

    ctx := context.Background()
    rs := NewRedisStore(mr.Addr(), "test")

    task := &models.Task{Type: "echo", Payload: map[string]any{"msg": "hello"}, Status: "queued"}
    created, existed, err := rs.CreateOrGetByKey(ctx, "k1", task)
    if err != nil { t.Fatalf("create error: %v", err) }
    if existed { t.Fatalf("expected not existed on first create") }
    if created.ID == "" { t.Fatalf("expected id set") }

    // Second create with same idempotency key should return same ID
    created2, existed2, err := rs.CreateOrGetByKey(ctx, "k1", task)
    if err != nil { t.Fatalf("second create error: %v", err) }
    if !existed2 { t.Fatalf("expected existed on second create") }
    if created.ID != created2.ID { t.Fatalf("expected same id, got %s vs %s", created.ID, created2.ID) }

    // Get
    got, err := rs.Get(ctx, created.ID)
    if err != nil { t.Fatalf("get error: %v", err) }
    if got.Type != task.Type { t.Fatalf("expected type %s, got %s", task.Type, got.Type) }

    // Update status
    if err := rs.UpdateStatus(ctx, created.ID, "done", map[string]any{"ok": true}); err != nil {
        t.Fatalf("update status error: %v", err)
    }
    got2, err := rs.Get(ctx, created.ID)
    if err != nil { t.Fatalf("get after update error: %v", err) }
    if got2.Status != "done" { t.Fatalf("expected status done, got %s", got2.Status) }
}

func TestRedisStore_Get_NotFound(t *testing.T) {
    mr, err := miniredis.Run()
    if err != nil {
        t.Fatalf("failed to start miniredis: %v", err)
    }
    defer mr.Close()

    ctx := context.Background()
    rs := NewRedisStore(mr.Addr(), "test")

    _, err = rs.Get(ctx, "missing-id")
    if err == nil {
        t.Fatalf("expected error for missing id")
    }
}
