package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ezerops/continuum/internal/models"
)

// AgentHandler handles agent-related requests
type AgentHandler struct {
	agentStore *models.AgentStore
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(agentStore *models.AgentStore) *AgentHandler {
	return &AgentHandler{agentStore: agentStore}
}

// CreateAgentRequest represents an agent creation request
type CreateAgentRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateAgentResponse includes the API key (shown only once)
type CreateAgentResponse struct {
	Agent  *models.Agent `json:"agent"`
	APIKey string        `json:"api_key"`
}

// Create creates a new agent
func (h *AgentHandler) Create(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	agent, apiKey, err := h.agentStore.Create(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateAgentResponse{
		Agent:  agent,
		APIKey: apiKey,
	})
}

// List lists all agents
func (h *AgentHandler) List(c *gin.Context) {
	agents, err := h.agentStore.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}
