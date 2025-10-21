package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/husainaj20/task-manager-api/internal/models"
	"github.com/redis/go-redis/v9"
)

var errRedisNotFound = errors.New("task not found")

type RedisStore struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisStore(addr string, prefix string) *RedisStore {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisStore{rdb: rdb, prefix: prefix}
}

func (r *RedisStore) key(id string) string { return r.prefix + ":task:" + id }

func (r *RedisStore) CreateOrGetByKey(ctx context.Context, key string, t *models.Task) (*models.Task, bool, error) {
	// Simple idempotency via separate key -> id mapping
	if key != "" {
		if id, err := r.rdb.Get(ctx, r.prefix+":idem:"+key).Result(); err == nil {
			data, err := r.rdb.Get(ctx, r.key(id)).Result()
			if err != nil {
				return nil, false, err
			}
			var existing models.Task
			if err := json.Unmarshal([]byte(data), &existing); err != nil {
				return nil, false, err
			}
			return &existing, true, nil
		}
	}

	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	t.CreatedAt, t.UpdatedAt = now, now
	b, err := json.Marshal(t)
	if err != nil {
		return nil, false, err
	}

	if err := r.rdb.Set(ctx, r.key(t.ID), b, 0).Err(); err != nil {
		return nil, false, err
	}
	if key != "" {
		if err := r.rdb.Set(ctx, r.prefix+":idem:"+key, t.ID, 0).Err(); err != nil {
			return nil, false, err
		}
	}
	return t, false, nil
}

func (r *RedisStore) Get(ctx context.Context, id string) (*models.Task, error) {
	s, err := r.rdb.Get(ctx, r.key(id)).Result()
	if err == redis.Nil {
		return nil, errRedisNotFound
	}
	if err != nil {
		return nil, err
	}
	var t models.Task
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *RedisStore) UpdateStatus(ctx context.Context, id string, status string, result map[string]any) error {
	t, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	t.Status = status
	if result != nil {
		t.Result = result
	}
	t.UpdatedAt = time.Now().UTC()
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, r.key(id), b, 0).Err()
}
