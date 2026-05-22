package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Timestamp string            `json:"timestamp"`
}

// Health performs a basic health check
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Version:   "0.1.0",
		Services:  map[string]string{"api": "up"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready checks if the service is ready to accept traffic
func (h *HealthHandler) Ready(c *gin.Context) {
	services := make(map[string]string)

	// Check database
	if err := h.db.Ping(c.Request.Context()); err != nil {
		services["database"] = "down"
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    "not_ready",
			Version:   "0.1.0",
			Services:  services,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}
	services["database"] = "up"

	// Check Redis
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		services["redis"] = "down"
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    "not_ready",
			Version:   "0.1.0",
			Services:  services,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}
	services["redis"] = "up"

	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ready",
		Version:   "0.1.0",
		Services:  services,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Live is a lightweight liveness probe
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
