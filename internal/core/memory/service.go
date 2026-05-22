package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ezerops/continuum/internal/models"
)

// Service handles memory operations with business logic
type Service struct {
	memoryStore  *models.MemoryStore
	sessionStore *models.SessionStore
	costStore    *models.CostStore
}

// NewService creates a new memory service
func NewService(memoryStore *models.MemoryStore, sessionStore *models.SessionStore, costStore *models.CostStore) *Service {
	return &Service{
		memoryStore:  memoryStore,
		sessionStore: sessionStore,
		costStore:    costStore,
	}
}

// WriteMemoryRequest represents a request to write memory
type WriteMemoryRequest struct {
	Type       models.MemoryType      `json:"type"`
	Content    string                 `json:"content"`
	Project    string                 `json:"project,omitempty"`
	Visibility models.Visibility      `json:"visibility,omitempty"`
	SessionID  *uuid.UUID             `json:"session_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Write creates a new memory with validation and processing
func (s *Service) Write(ctx context.Context, agentID uuid.UUID, req WriteMemoryRequest) (*models.Memory, error) {
	// Validate memory type
	if req.Type == "" {
		req.Type = models.MemoryTypeEpisodic
	}

	// Default visibility
	if req.Visibility == "" {
		req.Visibility = models.VisibilityShared
	}

	// If no session provided, try to get active session
	if req.SessionID == nil {
		activeSession, err := s.sessionStore.GetActiveByAgent(ctx, agentID)
		if err == nil && activeSession != nil {
			req.SessionID = &activeSession.ID
		}
	}

	memory := &models.Memory{
		Type:       req.Type,
		Content:    req.Content,
		AgentID:    agentID,
		SessionID:  req.SessionID,
		Project:    req.Project,
		Visibility: req.Visibility,
		Importance: 0.5, // Default importance
		Confidence: 1.0,
		Metadata:   req.Metadata,
	}

	if err := s.memoryStore.Create(ctx, memory); err != nil {
		return nil, fmt.Errorf("create memory: %w", err)
	}

	return memory, nil
}

// ReadMemory retrieves a memory by ID with access control
func (s *Service) Read(ctx context.Context, agentID uuid.UUID, memoryID uuid.UUID) (*models.Memory, error) {
	memory, err := s.memoryStore.GetByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}

	if memory == nil {
		return nil, fmt.Errorf("memory not found")
	}

	// Check visibility
	if memory.Visibility == models.VisibilityPrivate && memory.AgentID != agentID {
		return nil, fmt.Errorf("memory not found")
	}

	return memory, nil
}

// Search searches memories with access control
func (s *Service) Search(ctx context.Context, agentID uuid.UUID, query string, memoryType models.MemoryType, limit int) ([]models.Memory, error) {
	return s.memoryStore.Search(ctx, agentID, query, memoryType, limit)
}

// GetRecentShared gets recent shared memories across all agents
func (s *Service) GetRecentShared(ctx context.Context, agentID uuid.UUID, project string, limit int) ([]models.Memory, error) {
	return s.memoryStore.SearchRecent(ctx, "", project, limit)
}

// GetSessionContext gets all memories for a session
func (s *Service) GetSessionContext(ctx context.Context, sessionID uuid.UUID) ([]models.Memory, error) {
	return s.memoryStore.ListBySession(ctx, sessionID)
}

// ArchiveMemory archives a memory
func (s *Service) Archive(ctx context.Context, agentID uuid.UUID, memoryID uuid.UUID) error {
	return s.memoryStore.Archive(ctx, memoryID, agentID)
}
