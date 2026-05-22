package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MemoryStore handles memory CRUD operations
type MemoryStore struct {
	store *Store
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(store *Store) *MemoryStore {
	return &MemoryStore{store: store}
}

// Create creates a new memory
func (s *MemoryStore) Create(ctx context.Context, memory *Memory) error {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}
	if memory.Visibility == "" {
		memory.Visibility = VisibilityShared
	}
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}
	memory.UpdatedAt = memory.CreatedAt

	_, err := s.store.DB.Exec(ctx, `
		INSERT INTO memories (id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, memory.ID, memory.Type, memory.Content, memory.AgentID, memory.SessionID,
		memory.Project, memory.Visibility, memory.Importance, memory.Confidence,
		memory.Metadata, memory.ParentID, memory.CreatedAt, memory.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}

	return nil
}

// GetByID retrieves a memory by ID
func (s *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (*Memory, error) {
	memory := &Memory{}
	err := s.store.DB.QueryRow(ctx, `
		SELECT id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, parent_id, is_archived, archived_at, created_at, updated_at
		FROM memories WHERE id = $1
	`, id).Scan(
		&memory.ID, &memory.Type, &memory.Content, &memory.AgentID, &memory.SessionID,
		&memory.Project, &memory.Visibility, &memory.Importance, &memory.Confidence,
		&memory.Metadata, &memory.ParentID, &memory.IsArchived, &memory.ArchivedAt,
		&memory.CreatedAt, &memory.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get memory: %w", err)
	}
	return memory, nil
}

// Search searches memories by query text
func (s *MemoryStore) Search(ctx context.Context, agentID uuid.UUID, query string, memoryType MemoryType, limit int) ([]Memory, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Build query dynamically
	var args []interface{}
	argIdx := 1

	sql := `
		SELECT id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, created_at, updated_at
		FROM memories
		WHERE is_archived = false
		AND (
			visibility != 'private'
			OR agent_id = $1
		)
	`
	args = append(args, agentID)
	argIdx++

	if memoryType != "" {
		sql += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, memoryType)
		argIdx++
	}

	if query != "" {
		sql += fmt.Sprintf(` AND (
			to_tsvector('english', content) @@ plainto_tsquery('english', $%d)
			OR content ILIKE '%%' || $%d || '%%'
		)`, argIdx, argIdx)
		args = append(args, query)
		argIdx++
	}

	sql += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.store.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.Type, &m.Content, &m.AgentID, &m.SessionID,
			&m.Project, &m.Visibility, &m.Importance, &m.Confidence,
			&m.Metadata, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// ListByAgent lists memories for a specific agent
func (s *MemoryStore) ListByAgent(ctx context.Context, agentID uuid.UUID, memoryType MemoryType, limit int) ([]Memory, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var sql string
	var args []interface{}

	if memoryType != "" {
		sql = `
			SELECT id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, created_at, updated_at
			FROM memories
			WHERE agent_id = $1 AND type = $2 AND is_archived = false
			ORDER BY created_at DESC LIMIT $3
		`
		args = []interface{}{agentID, memoryType, limit}
	} else {
		sql = `
			SELECT id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, created_at, updated_at
			FROM memories
			WHERE agent_id = $1 AND is_archived = false
			ORDER BY created_at DESC LIMIT $2
		`
		args = []interface{}{agentID, limit}
	}

	rows, err := s.store.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.Type, &m.Content, &m.AgentID, &m.SessionID,
			&m.Project, &m.Visibility, &m.Importance, &m.Confidence,
			&m.Metadata, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// ListBySession lists memories for a session
func (s *MemoryStore) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]Memory, error) {
	rows, err := s.store.DB.Query(ctx, `
		SELECT id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, created_at, updated_at
		FROM memories
		WHERE session_id = $1 AND is_archived = false
		ORDER BY created_at DESC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list session memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.Type, &m.Content, &m.AgentID, &m.SessionID,
			&m.Project, &m.Visibility, &m.Importance, &m.Confidence,
			&m.Metadata, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// Update updates a memory
func (s *MemoryStore) Update(ctx context.Context, memory *Memory) error {
	memory.UpdatedAt = time.Now()

	_, err := s.store.DB.Exec(ctx, `
		UPDATE memories
		SET content = $1, project = $2, visibility = $3, importance = $4, confidence = $5, metadata = $6, updated_at = $7
		WHERE id = $8 AND agent_id = $9
	`, memory.Content, memory.Project, memory.Visibility, memory.Importance,
		memory.Confidence, memory.Metadata, memory.UpdatedAt, memory.ID, memory.AgentID)

	if err != nil {
		return fmt.Errorf("update memory: %w", err)
	}

	return nil
}

// Archive marks a memory as archived
func (s *MemoryStore) Archive(ctx context.Context, id uuid.UUID, agentID uuid.UUID) error {
	_, err := s.store.DB.Exec(ctx, `
		UPDATE memories
		SET is_archived = true, archived_at = $1
		WHERE id = $2 AND agent_id = $3
	`, time.Now(), id, agentID)
	if err != nil {
		return fmt.Errorf("archive memory: %w", err)
	}
	return nil
}

// Delete permanently deletes a memory
func (s *MemoryStore) Delete(ctx context.Context, id uuid.UUID, agentID uuid.UUID) error {
	_, err := s.store.DB.Exec(ctx, `
		DELETE FROM memories WHERE id = $1 AND agent_id = $2
	`, id, agentID)
	if err != nil {
		return fmt.Errorf("delete memory: %w", err)
	}
	return nil
}

// CountByAgent counts memories for an agent
func (s *MemoryStore) CountByAgent(ctx context.Context, agentID uuid.UUID) (int, error) {
	var count int
	err := s.store.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM memories WHERE agent_id = $1 AND is_archived = false
	`, agentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count memories: %w", err)
	}
	return count, nil
}

// SearchRecent finds recent memories across all agents (for shared context)
func (s *MemoryStore) SearchRecent(ctx context.Context, query string, project string, limit int) ([]Memory, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var sql strings.Builder
	var args []interface{}
	argIdx := 1

	sql.WriteString(`
		SELECT id, type, content, agent_id, session_id, project, visibility, importance, confidence, metadata, created_at, updated_at
		FROM memories
		WHERE is_archived = false
		AND visibility IN ('shared', 'broadcast')
	`)

	if project != "" {
		sql.WriteString(fmt.Sprintf(" AND project = $%d", argIdx))
		args = append(args, project)
		argIdx++
	}

	if query != "" {
		sql.WriteString(fmt.Sprintf(` AND (
			to_tsvector('english', content) @@ plainto_tsquery('english', $%d)
			OR content ILIKE '%%' || $%d || '%%'
		)`, argIdx, argIdx))
		args = append(args, query)
		argIdx++
	}

	sql.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx))
	args = append(args, limit)

	rows, err := s.store.DB.Query(ctx, sql.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("search recent memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.Type, &m.Content, &m.AgentID, &m.SessionID,
			&m.Project, &m.Visibility, &m.Importance, &m.Confidence,
			&m.Metadata, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}
