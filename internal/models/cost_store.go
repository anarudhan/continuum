package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CostStore handles cost tracking operations
type CostStore struct {
	store *Store
}

// NewCostStore creates a new cost store
func NewCostStore(store *Store) *CostStore {
	return &CostStore{store: store}
}

// LogCost records a cost event
func (s *CostStore) LogCost(ctx context.Context, agentID uuid.UUID, sessionID *uuid.UUID, provider, model string, tokensInput, tokensOutput int, costUSD float64, requestType string) error {
	_, err := s.store.DB.Exec(ctx, `
		INSERT INTO cost_logs (agent_id, session_id, provider, model, tokens_input, tokens_output, tokens_total, cost_usd, request_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, agentID, sessionID, provider, model, tokensInput, tokensOutput, tokensInput+tokensOutput, costUSD, requestType, time.Now())
	if err != nil {
		return fmt.Errorf("log cost: %w", err)
	}
	return nil
}

// GetDailyCost gets total cost for an agent today
func (s *CostStore) GetDailyCost(ctx context.Context, agentID uuid.UUID) (float64, error) {
	var cost float64
	err := s.store.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost_usd), 0)
		FROM cost_logs
		WHERE agent_id = $1 AND created_at >= CURRENT_DATE
	`, agentID).Scan(&cost)
	if err != nil {
		return 0, fmt.Errorf("get daily cost: %w", err)
	}
	return cost, nil
}

// GetWeeklyCost gets total cost for an agent this week
func (s *CostStore) GetWeeklyCost(ctx context.Context, agentID uuid.UUID) (float64, error) {
	var cost float64
	err := s.store.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost_usd), 0)
		FROM cost_logs
		WHERE agent_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '7 days'
	`, agentID).Scan(&cost)
	if err != nil {
		return 0, fmt.Errorf("get weekly cost: %w", err)
	}
	return cost, nil
}

// GetMonthlyCost gets total cost for an agent this month
func (s *CostStore) GetMonthlyCost(ctx context.Context, agentID uuid.UUID) (float64, error) {
	var cost float64
	err := s.store.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost_usd), 0)
		FROM cost_logs
		WHERE agent_id = $1 AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`, agentID).Scan(&cost)
	if err != nil {
		return 0, fmt.Errorf("get monthly cost: %w", err)
	}
	return cost, nil
}

// GetCostBreakdown gets cost breakdown by provider/model
func (s *CostStore) GetCostBreakdown(ctx context.Context, agentID uuid.UUID, days int) ([]CostBreakdown, error) {
	if days <= 0 {
		days = 30
	}

	rows, err := s.store.DB.Query(ctx, `
		SELECT provider, model, SUM(tokens_input) as tokens_in, SUM(tokens_output) as tokens_out, SUM(cost_usd) as cost
		FROM cost_logs
		WHERE agent_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '$2 days'
		GROUP BY provider, model
		ORDER BY cost DESC
	`, agentID, days)
	if err != nil {
		return nil, fmt.Errorf("get cost breakdown: %w", err)
	}
	defer rows.Close()

	var breakdowns []CostBreakdown
	for rows.Next() {
		var b CostBreakdown
		err := rows.Scan(&b.Provider, &b.Model, &b.TokensInput, &b.TokensOutput, &b.CostUSD)
		if err != nil {
			return nil, fmt.Errorf("scan cost breakdown: %w", err)
		}
		breakdowns = append(breakdowns, b)
	}

	return breakdowns, nil
}

// GetProviderPricing gets pricing for a provider/model
func (s *CostStore) GetProviderPricing(ctx context.Context, provider, model string) (*ProviderPricing, error) {
	pricing := &ProviderPricing{}
	err := s.store.DB.QueryRow(ctx, `
		SELECT id, provider, model, price_per_1k_input, price_per_1k_output, currency, effective_from
		FROM provider_pricing
		WHERE provider = $1 AND model = $2
		ORDER BY effective_from DESC
		LIMIT 1
	`, provider, model).Scan(
		&pricing.ID, &pricing.Provider, &pricing.Model, &pricing.PricePer1KInput,
		&pricing.PricePer1KOutput, &pricing.Currency, &pricing.EffectiveFrom,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get provider pricing: %w", err)
	}
	return pricing, nil
}

// CostBreakdown represents cost aggregation
type CostBreakdown struct {
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	TokensInput  int     `json:"tokens_input"`
	TokensOutput int     `json:"tokens_output"`
	CostUSD      float64 `json:"cost_usd"`
}
