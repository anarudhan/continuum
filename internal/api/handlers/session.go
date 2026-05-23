package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/anarudhan/continuum/internal/api/middleware"
	"github.com/anarudhan/continuum/internal/models"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	sessionStore *models.SessionStore
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionStore *models.SessionStore) *SessionHandler {
	return &SessionHandler{sessionStore: sessionStore}
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	Project string `json:"project,omitempty"`
	Task    string `json:"task,omitempty"`
}

// Create creates a new session
func (h *SessionHandler) Create(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	session, err := h.sessionStore.Create(c.Request.Context(), agentID, req.Project, req.Task)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "create_failed", err)
		return
	}

	c.JSON(http.StatusCreated, session)
}

// Get retrieves a session by ID
func (h *SessionHandler) Get(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid session ID"})
		return
	}

	session, err := h.sessionStore.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "fetch_failed", err)
		return
	}

	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Session not found"})
		return
	}

	// Verify ownership
	if session.AgentID != agentID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// List lists sessions for the current agent
func (h *SessionHandler) List(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	sessions, err := h.sessionStore.ListByAgent(c.Request.Context(), agentID, limit)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "list_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// EndSessionRequest represents a session end request
type EndSessionRequest struct {
	Summary string `json:"summary,omitempty"`
}

// End marks a session as completed
func (h *SessionHandler) End(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid session ID"})
		return
	}

	// Verify ownership before ending
	session, err := h.sessionStore.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "fetch_failed", err)
		return
	}

	if session == nil || session.AgentID != agentID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Session not found"})
		return
	}

	var req EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	if err := h.sessionStore.EndSession(c.Request.Context(), id, req.Summary); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "end_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session ended"})
}
