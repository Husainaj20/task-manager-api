package store

import (
	"context"

	"github.com/husainaj20/task-manager-api/internal/models"
)

// Store defines the operations used by the API/service layers.
type Store interface {
	CreateOrGetByKey(ctx context.Context, key string, t *models.Task) (*models.Task, bool, error)
	Get(ctx context.Context, id string) (*models.Task, error)
	UpdateStatus(ctx context.Context, id string, status string, result map[string]any) error
}
