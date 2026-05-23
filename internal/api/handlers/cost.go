package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/anarudhan/continuum/internal/api/middleware"
	"github.com/anarudhan/continuum/internal/models"
)

// CostHandler handles cost tracking requests
type CostHandler struct {
	costStore *models.CostStore
}

// NewCostHandler creates a new cost handler
func NewCostHandler(costStore *models.CostStore) *CostHandler {
	return &CostHandler{costStore: costStore}
}

// GetCosts returns cost summary for the authenticated agent
func (h *CostHandler) GetCosts(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	daily, err := h.costStore.GetDailyCost(c.Request.Context(), agentID)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "fetch_failed", err)
		return
	}

	weekly, err := h.costStore.GetWeeklyCost(c.Request.Context(), agentID)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "fetch_failed", err)
		return
	}

	monthly, err := h.costStore.GetMonthlyCost(c.Request.Context(), agentID)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "fetch_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"daily":   daily,
		"weekly":  weekly,
		"monthly": monthly,
		"currency": "USD",
	})
}

// GetCostBreakdown returns detailed cost breakdown by provider/model
func (h *CostHandler) GetCostBreakdown(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	breakdown, err := h.costStore.GetCostBreakdown(c.Request.Context(), agentID, days)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "fetch_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"breakdown": breakdown,
		"days":      days,
		"currency":  "USD",
	})
}
