package models

import (
	"context"
	"crypto/sha256"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// AgentStore handles agent CRUD operations
type AgentStore struct {
	store *Store
}

// NewAgentStore creates a new agent store
func NewAgentStore(store *Store) *AgentStore {
	return &AgentStore{store: store}
}

// Create creates a new agent
func (s *AgentStore) Create(ctx context.Context, name string) (*Agent, string, error) {
	// Generate API key
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("generate API key: %w", err)
	}

	apiKeyHash := hashAPIKey(apiKey)
	apiKeyHint := ""
	if len(apiKey) >= 4 {
		apiKeyHint = apiKey[len(apiKey)-4:]
	}

	agent := &Agent{
		ID:         uuid.New(),
		Name:       name,
		APIKeyHash: apiKeyHash,
		APIKeyHint: apiKeyHint,
		TrustLevel: "trusted",
		Scopes:     []string{"read", "write"},
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err = s.store.DB.Exec(ctx, `
		INSERT INTO agents (id, name, api_key_hash, api_key_hint, trust_level, scopes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, agent.ID, agent.Name, agent.APIKeyHash, agent.APIKeyHint,
		agent.TrustLevel, agent.Scopes, agent.IsActive, agent.CreatedAt, agent.UpdatedAt)

	if err != nil {
		return nil, "", fmt.Errorf("insert agent: %w", err)
	}

	// Create default budget
	_, err = s.store.DB.Exec(ctx, `
		INSERT INTO budgets (agent_id, daily_limit, weekly_limit, monthly_limit)
		VALUES ($1, $2, $3, $4)
	`, agent.ID, 10.00, 50.00, 100.00)
	if err != nil {
		// Non-critical, log but don't fail
		fmt.Printf("warning: failed to create budget for agent %s: %v\n", agent.ID, err)
	}

	return agent, apiKey, nil
}

// GetByID retrieves an agent by ID
func (s *AgentStore) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	agent := &Agent{}
	err := s.store.DB.QueryRow(ctx, `
		SELECT id, name, api_key_hint, trust_level, scopes, is_active, last_seen_at, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(
		&agent.ID, &agent.Name, &agent.APIKeyHint, &agent.TrustLevel,
		&agent.Scopes, &agent.IsActive, &agent.LastSeenAt, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get agent: %w", err)
	}
	return agent, nil
}

// GetByAPIKeyHash retrieves an agent by API key hash
func (s *AgentStore) GetByAPIKeyHash(ctx context.Context, hash string) (*Agent, error) {
	agent := &Agent{}
	err := s.store.DB.QueryRow(ctx, `
		SELECT id, name, api_key_hash, api_key_hint, trust_level, scopes, is_active, last_seen_at, created_at, updated_at
		FROM agents WHERE api_key_hash = $1 AND is_active = true
	`, hash).Scan(
		&agent.ID, &agent.Name, &agent.APIKeyHash, &agent.APIKeyHint,
		&agent.Scopes, &agent.IsActive, &agent.LastSeenAt, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get agent by API key: %w", err)
	}
	return agent, nil
}

// UpdateLastSeen updates the agent's last seen timestamp
func (s *AgentStore) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	_, err := s.store.DB.Exec(ctx, `
		UPDATE agents SET last_seen_at = $1 WHERE id = $2
	`, time.Now(), id)
	return err
}

// List lists all agents
func (s *AgentStore) List(ctx context.Context) ([]Agent, error) {
	rows, err := s.store.DB.Query(ctx, `
		SELECT id, name, api_key_hint, trust_level, scopes, is_active, last_seen_at, created_at, updated_at
		FROM agents ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.APIKeyHint, &agent.TrustLevel,
			&agent.Scopes, &agent.IsActive, &agent.LastSeenAt, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// generateAPIKey creates a secure API key
func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "ctm_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// hashAPIKey hashes an API key for storage
func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
