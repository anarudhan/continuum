package models

import (
	"time"

	"github.com/google/uuid"
)

// Agent represents a registered AI agent
type Agent struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Name         string     `db:"name" json:"name"`
	APIKeyHash   string     `db:"api_key_hash" json:"-"`
	APIKeyHint   string     `db:"api_key_hint" json:"api_key_hint,omitempty"`
	TrustLevel   string     `db:"trust_level" json:"trust_level"`
	Scopes       []string   `db:"scopes" json:"scopes"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	LastSeenAt   *time.Time `db:"last_seen_at" json:"last_seen_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// Session represents an agent work session
type Session struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	AgentID    uuid.UUID  `db:"agent_id" json:"agent_id"`
	Project    string     `db:"project" json:"project,omitempty"`
	Task       string     `db:"task" json:"task,omitempty"`
	Status     string     `db:"status" json:"status"`
	Summary    string     `db:"summary" json:"summary,omitempty"`
	StartedAt  time.Time  `db:"started_at" json:"started_at"`
	EndedAt    *time.Time `db:"ended_at" json:"ended_at,omitempty"`
	TokenCount int        `db:"token_count" json:"token_count"`
	CostUSD    float64    `db:"cost_usd" json:"cost_usd"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

// MemoryType represents the type of memory
type MemoryType string

const (
	MemoryTypeEpisodic   MemoryType = "episodic"
	MemoryTypeSemantic   MemoryType = "semantic"
	MemoryTypeProcedural MemoryType = "procedural"
)

// Visibility represents memory visibility
type Visibility string

const (
	VisibilityPrivate   Visibility = "private"
	VisibilityShared    Visibility = "shared"
	VisibilityBroadcast Visibility = "broadcast"
)

// Memory represents a stored memory
type Memory struct {
	ID            uuid.UUID              `db:"id" json:"id"`
	Type          MemoryType             `db:"type" json:"type"`
	Content       string                 `db:"content" json:"content"`
	ContentVector []float32              `db:"content_vector" json:"-"`
	AgentID       uuid.UUID              `db:"agent_id" json:"agent_id"`
	SessionID     *uuid.UUID             `db:"session_id" json:"session_id,omitempty"`
	Project       string                 `db:"project" json:"project,omitempty"`
	Visibility    Visibility             `db:"visibility" json:"visibility"`
	Importance    float64                `db:"importance" json:"importance"`
	Confidence    float64                `db:"confidence" json:"confidence"`
	Metadata      map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	ParentID      *uuid.UUID             `db:"parent_id" json:"parent_id,omitempty"`
	IsArchived    bool                   `db:"is_archived" json:"is_archived"`
	ArchivedAt    *time.Time             `db:"archived_at" json:"archived_at,omitempty"`
	CreatedAt     time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `db:"updated_at" json:"updated_at"`
}

// Entity represents a named entity extracted from memories
type Entity struct {
	ID           uuid.UUID              `db:"id" json:"id"`
	Name         string                 `db:"name" json:"name"`
	EntityType   string                 `db:"entity_type" json:"entity_type"`
	Attributes   map[string]interface{} `db:"attributes" json:"attributes,omitempty"`
	AgentID      uuid.UUID              `db:"agent_id" json:"agent_id"`
	Confidence   float64                `db:"confidence" json:"confidence"`
	FirstSeenAt  time.Time              `db:"first_seen_at" json:"first_seen_at"`
	LastSeenAt   time.Time              `db:"last_seen_at" json:"last_seen_at"`
	MentionCount int                    `db:"mention_count" json:"mention_count"`
	CreatedAt    time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at" json:"updated_at"`
}

// Relationship represents a graph edge between entities
type Relationship struct {
	ID           uuid.UUID              `db:"id" json:"id"`
	SourceID     uuid.UUID              `db:"source_id" json:"source_id"`
	TargetID     uuid.UUID              `db:"target_id" json:"target_id"`
	RelationType string                 `db:"relation_type" json:"relation_type"`
	Attributes   map[string]interface{} `db:"attributes" json:"attributes,omitempty"`
	AgentID      uuid.UUID              `db:"agent_id" json:"agent_id"`
	Confidence   float64                `db:"confidence" json:"confidence"`
	CreatedAt    time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at" json:"updated_at"`
}

// CostLog represents a token usage record
type CostLog struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	AgentID      uuid.UUID  `db:"agent_id" json:"agent_id"`
	SessionID    *uuid.UUID `db:"session_id" json:"session_id,omitempty"`
	Provider     string     `db:"provider" json:"provider"`
	Model        string     `db:"model" json:"model"`
	TokensInput  int        `db:"tokens_input" json:"tokens_input"`
	TokensOutput int        `db:"tokens_output" json:"tokens_output"`
	TokensTotal  int        `db:"tokens_total" json:"tokens_total"`
	CostUSD      float64    `db:"cost_usd" json:"cost_usd"`
	RequestType  string     `db:"request_type" json:"request_type,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

// Budget represents spending limits for an agent
type Budget struct {
	ID             uuid.UUID `db:"id" json:"id"`
	AgentID        uuid.UUID `db:"agent_id" json:"agent_id"`
	DailyLimit     float64   `db:"daily_limit" json:"daily_limit"`
	WeeklyLimit    float64   `db:"weekly_limit" json:"weekly_limit"`
	MonthlyLimit   float64   `db:"monthly_limit" json:"monthly_limit"`
	AlertThreshold float64   `db:"alert_threshold" json:"alert_threshold"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	ID           uuid.UUID              `db:"id" json:"id"`
	Timestamp    time.Time              `db:"timestamp" json:"timestamp"`
	EventType    string                 `db:"event_type" json:"event_type"`
	AgentID      *uuid.UUID             `db:"agent_id" json:"agent_id,omitempty"`
	AgentName    string                 `db:"agent_name" json:"agent_name,omitempty"`
	Action       string                 `db:"action" json:"action"`
	Resource     string                 `db:"resource" json:"resource,omitempty"`
	Details      map[string]interface{} `db:"details" json:"details,omitempty"`
	IPAddress    string                 `db:"ip_address" json:"ip_address,omitempty"`
	Success      bool                   `db:"success" json:"success"`
	ErrorMessage string                 `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time              `db:"created_at" json:"created_at"`
}

// ProviderPricing represents LLM provider pricing
type ProviderPricing struct {
	ID               uuid.UUID `db:"id" json:"id"`
	Provider         string    `db:"provider" json:"provider"`
	Model            string    `db:"model" json:"model"`
	PricePer1KInput  float64   `db:"price_per_1k_input" json:"price_per_1k_input"`
	PricePer1KOutput float64   `db:"price_per_1k_output" json:"price_per_1k_output"`
	Currency         string    `db:"currency" json:"currency"`
	EffectiveFrom    time.Time `db:"effective_from" json:"effective_from"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}
