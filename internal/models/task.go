package models

import "time"

type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]any         `json:"payload,omitempty"`
	Status    string                 `json:"status"`
	Result    map[string]any         `json:"result,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}
