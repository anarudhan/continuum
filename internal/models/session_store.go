package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SessionStore handles session CRUD operations
type SessionStore struct {
	store *Store
}

// NewSessionStore creates a new session store
func NewSessionStore(store *Store) *SessionStore {
	return &SessionStore{store: store}
}

// Create creates a new session
func (s *SessionStore) Create(ctx context.Context, agentID uuid.UUID, project, task string) (*Session, error) {
	session := &Session{
		ID:        uuid.New(),
		AgentID:   agentID,
		Project:   project,
		Task:      task,
		Status:    "active",
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	_, err := s.store.DB.Exec(ctx, `
		INSERT INTO sessions (id, agent_id, project, task, status, started_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, session.ID, session.AgentID, session.Project, session.Task,
		session.Status, session.StartedAt, session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	return session, nil
}

// GetByID retrieves a session by ID
func (s *SessionStore) GetByID(ctx context.Context, id uuid.UUID) (*Session, error) {
	session := &Session{}
	err := s.store.DB.QueryRow(ctx, `
		SELECT id, agent_id, project, task, status, summary, started_at, ended_at, token_count, cost_usd, created_at
		FROM sessions WHERE id = $1
	`, id).Scan(
		&session.ID, &session.AgentID, &session.Project, &session.Task,
		&session.Status, &session.Summary, &session.StartedAt, &session.EndedAt,
		&session.TokenCount, &session.CostUSD, &session.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return session, nil
}

// GetActiveByAgent gets the active session for an agent
func (s *SessionStore) GetActiveByAgent(ctx context.Context, agentID uuid.UUID) (*Session, error) {
	session := &Session{}
	err := s.store.DB.QueryRow(ctx, `
		SELECT id, agent_id, project, task, status, summary, started_at, ended_at, token_count, cost_usd, created_at
		FROM sessions
		WHERE agent_id = $1 AND status = 'active'
		ORDER BY started_at DESC
		LIMIT 1
	`, agentID).Scan(
		&session.ID, &session.AgentID, &session.Project, &session.Task,
		&session.Status, &session.Summary, &session.StartedAt, &session.EndedAt,
		&session.TokenCount, &session.CostUSD, &session.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get active session: %w", err)
	}
	return session, nil
}

// ListByAgent lists sessions for an agent
func (s *SessionStore) ListByAgent(ctx context.Context, agentID uuid.UUID, limit int) ([]Session, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.store.DB.Query(ctx, `
		SELECT id, agent_id, project, task, status, summary, started_at, ended_at, token_count, cost_usd, created_at
		FROM sessions
		WHERE agent_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		err := rows.Scan(
			&sess.ID, &sess.AgentID, &sess.Project, &sess.Task,
			&sess.Status, &sess.Summary, &sess.StartedAt, &sess.EndedAt,
			&sess.TokenCount, &sess.CostUSD, &sess.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, sess)
	}

	return sessions, nil
}

// EndSession marks a session as completed
func (s *SessionStore) EndSession(ctx context.Context, id uuid.UUID, summary string) error {
	_, err := s.store.DB.Exec(ctx, `
		UPDATE sessions
		SET status = 'completed', summary = $1, ended_at = $2
		WHERE id = $3
	`, summary, time.Now(), id)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}
	return nil
}

// UpdateCost updates the cost for a session
func (s *SessionStore) UpdateCost(ctx context.Context, id uuid.UUID, tokens int, cost float64) error {
	_, err := s.store.DB.Exec(ctx, `
		UPDATE sessions
		SET token_count = token_count + $1, cost_usd = cost_usd + $2
		WHERE id = $3
	`, tokens, cost, id)
	if err != nil {
		return fmt.Errorf("update session cost: %w", err)
	}
	return nil
}
