package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ezerops/continuum/internal/models"
)

const (
	APIKeyHeader = "X-API-Key"
	APIKeyPrefix = "ctm_"
)

// AuthMiddleware handles API key authentication
type AuthMiddleware struct {
	agentStore *models.AgentStore
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(agentStore *models.AgentStore) *AuthMiddleware {
	return &AuthMiddleware{agentStore: agentStore}
}

// RequireAPIKey ensures the request has a valid API key
func (a *AuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := extractAPIKey(c)
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "missing_api_key",
				Message: "Set CONTINUUM_API_KEY environment variable",
			})
			return
		}

		if !strings.HasPrefix(key, APIKeyPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "invalid_api_key_format",
				Message: "API key must start with 'ctm_'",
			})
			return
		}

		hash := hashAPIKey(key)
		agent, err := a.agentStore.GetByAPIKeyHash(c.Request.Context(), hash)
		if err != nil || agent == nil || !agent.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "invalid_api_key",
				Message: "Check your CONTINUUM_API_KEY",
			})
			return
		}

		// Update last seen (fire and forget)
		go a.agentStore.UpdateLastSeen(c.Request.Context(), agent.ID)

		// Set agent context for downstream handlers
		c.Set("agent_id", agent.ID)
		c.Set("agent_name", agent.Name)
		c.Set("agent_trust_level", agent.TrustLevel)
		c.Set("agent_scopes", agent.Scopes)

		c.Next()
	}
}

// extractAPIKey extracts the API key from the request
func extractAPIKey(c *gin.Context) string {
	// Check header
	key := c.GetHeader(APIKeyHeader)
	if key != "" {
		return key
	}

	// Check Authorization header
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check query param (for WebSocket)
	return c.Query("api_key")
}

// hashAPIKey hashes an API key for lookup
func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
