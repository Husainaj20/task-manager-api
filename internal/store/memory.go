package store

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/husainaj20/task-manager-api/internal/models"
)

var (
	errNotFound = errors.New("task not found")
)

type MemoryStore struct {
	mu        sync.RWMutex
	tasks     map[string]*models.Task
	idemIndex map[string]string // idempotency key -> taskID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks:     make(map[string]*models.Task),
		idemIndex: make(map[string]string),
	}
}

func (m *MemoryStore) CreateOrGetByKey(ctx context.Context, key string, t *models.Task) (*models.Task, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key != "" {
		if id, ok := m.idemIndex[key]; ok {
			if existing, ok := m.tasks[id]; ok {
				return clone(existing), true, nil
			}
		}
	}

	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	t.CreatedAt, t.UpdatedAt = now, now
	m.tasks[t.ID] = clone(t)
	if key != "" {
		m.idemIndex[key] = t.ID
	}
	return clone(t), false, nil
}

func (m *MemoryStore) Get(ctx context.Context, id string) (*models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.tasks[id]; ok {
		return clone(t), nil
	}
	return nil, errNotFound
}

func (m *MemoryStore) UpdateStatus(ctx context.Context, id string, status string, result map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return errNotFound
	}
	t.Status = status
	if result != nil {
		t.Result = result
	}
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func clone(t *models.Task) *models.Task {
	if t == nil { return nil }
	c := *t
	return &c
}
