package api

import (
	"net/http"
	"time"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/husainaj20/task-manager-api/internal/models"
	"github.com/husainaj20/task-manager-api/internal/service"
	"github.com/husainaj20/task-manager-api/internal/store"
)

type Handler struct {
	store *store.MemoryStore
	q     *service.Queue
}

func New(s *store.MemoryStore, q *service.Queue) *Handler {
	return &Handler{store: s, q: q}
}

func (h *Handler) Router() http.Handler {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/readiness", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ready": true}) })

	r.POST("/tasks", h.createTask)
	r.GET("/tasks/:id", h.getTask)

	return r
}

type createTaskReq struct {
	Type    string         `json:"type" binding:"required"`
	Payload map[string]any `json:"payload"`
}

func (h *Handler) createTask(c *gin.Context) {
	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idemKey := c.GetHeader("Idempotency-Key")
	t := &models.Task{
		Type:   req.Type,
		Payload: req.Payload,
		Status: "queued",
	}
	ctx := context.Background()
	task, existed, err := h.store.CreateOrGetByKey(ctx, idemKey, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !existed {
		h.q.Enqueue(&service.TaskWork{
			ID: task.ID,
			Result: map[string]any{"echo": req.Payload, "processedAt": time.Now().UTC()},
		})
	}
	c.JSON(http.StatusAccepted, task)
}

func (h *Handler) getTask(c *gin.Context) {
	id := c.Param("id")
	t, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}
