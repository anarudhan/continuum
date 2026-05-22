package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ezerops/continuum/internal/models"
)

// MemoryHandler handles memory-related requests
type MemoryHandler struct {
	memoryStore *models.MemoryStore
}

// NewMemoryHandler creates a new memory handler
func NewMemoryHandler(memoryStore *models.MemoryStore) *MemoryHandler {
	return &MemoryHandler{memoryStore: memoryStore}
}

// CreateMemoryRequest represents a memory creation request
type CreateMemoryRequest struct {
	Type       string                 `json:"type" binding:"required,oneof=episodic semantic procedural"`
	Content    string                 `json:"content" binding:"required"`
	Project    string                 `json:"project,omitempty"`
	Visibility string                 `json:"visibility,omitempty" binding:"omitempty,oneof=private shared broadcast"`
	SessionID  *uuid.UUID             `json:"session_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Create creates a new memory
func (h *MemoryHandler) Create(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	// Validate and sanitize content
	sanitized, err := sanitizeMemoryContent(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_content", "message": err.Error()})
		return
	}

	// Set default visibility
	visibility := models.VisibilityShared
	if req.Visibility != "" {
		visibility = models.Visibility(req.Visibility)
	}

	memory := &models.Memory{
		Type:       models.MemoryType(req.Type),
		Content:    sanitized,
		AgentID:    agentID,
		SessionID:  req.SessionID,
		Project:    req.Project,
		Visibility: visibility,
		Metadata:   req.Metadata,
	}

	if err := h.memoryStore.Create(c.Request.Context(), memory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, memory)
}

// Get retrieves a memory by ID
func (h *MemoryHandler) Get(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid memory ID"})
		return
	}

	memory, err := h.memoryStore.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch_failed", "message": err.Error()})
		return
	}

	if memory == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Memory not found"})
		return
	}

	// Check visibility
	if memory.Visibility == models.VisibilityPrivate && memory.AgentID != agentID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Memory not found"})
		return
	}

	c.JSON(http.StatusOK, memory)
}

// Search searches memories
func (h *MemoryHandler) Search(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	query := c.Query("q")
	memoryType := c.Query("type")
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var mType models.MemoryType
	if memoryType != "" {
		mType = models.MemoryType(memoryType)
	}

	memories, err := h.memoryStore.Search(c.Request.Context(), agentID, query, mType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"count":    len(memories),
	})
}

// List lists memories for the current agent
func (h *MemoryHandler) List(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	memoryType := c.Query("type")
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var mType models.MemoryType
	if memoryType != "" {
		mType = models.MemoryType(memoryType)
	}

	memories, err := h.memoryStore.ListByAgent(c.Request.Context(), agentID, mType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"count":    len(memories),
	})
}

// Delete deletes a memory
func (h *MemoryHandler) Delete(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid memory ID"})
		return
	}

	if err := h.memoryStore.Delete(c.Request.Context(), id, agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory deleted"})
}

// sanitizeMemoryContent validates and sanitizes memory content
func sanitizeMemoryContent(content string) (string, error) {
	const maxSize = 10 * 1024 * 1024 // 10MB

	if utf8.RuneCountInString(content) > maxSize {
		return "", fmt.Errorf("content exceeds 10MB")
	}

	// Remove null bytes (PostgreSQL requirement)
	content = strings.ReplaceAll(content, "\x00", "")

	// Trim whitespace
	content = strings.TrimSpace(content)

	if content == "" {
		return "", fmt.Errorf("content is empty")
	}

	return content, nil
}
