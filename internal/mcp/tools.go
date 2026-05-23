package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/anarudhan/continuum/internal/core/memory"
	"github.com/anarudhan/continuum/internal/models"
)

// ToolContext holds the authenticated agent context for MCP tool calls
type ToolContext struct {
	AgentID    uuid.UUID
	AgentName  string
	TrustLevel string
	Scopes     []string
}

// RegisterMemoryTools registers memory-related MCP tools with full auth context
func RegisterMemoryTools(server *Server, memoryService *memory.Service, sessionStore *models.SessionStore) {
	server.RegisterHandler("continuum/memory_write", func(ctx context.Context, toolCtx *ToolContext, params json.RawMessage) (interface{}, error) {
		if toolCtx == nil {
			return nil, fmt.Errorf("unauthenticated: MCP tools require authenticated agent context")
		}

		var req memory.WriteMemoryRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		mem, err := memoryService.Write(ctx, toolCtx.AgentID, req)
		if err != nil {
			return nil, fmt.Errorf("write memory: %w", err)
		}

		return map[string]interface{}{
			"id":      mem.ID,
			"type":    mem.Type,
			"content": mem.Content,
			"status":  "created",
		}, nil
	})

	server.RegisterHandler("continuum/memory_search", func(ctx context.Context, toolCtx *ToolContext, params json.RawMessage) (interface{}, error) {
		if toolCtx == nil {
			return nil, fmt.Errorf("unauthenticated: MCP tools require authenticated agent context")
		}

		var req struct {
			Query string `json:"query"`
			Type  string `json:"type,omitempty"`
			Limit int    `json:"limit,omitempty"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		memType := models.MemoryType(req.Type)
		if req.Limit <= 0 || req.Limit > 100 {
			req.Limit = 20
		}

		memories, err := memoryService.Search(ctx, toolCtx.AgentID, req.Query, memType, req.Limit)
		if err != nil {
			return nil, fmt.Errorf("search memory: %w", err)
		}

		return map[string]interface{}{
			"memories": memories,
			"count":    len(memories),
		}, nil
	})

	server.RegisterHandler("continuum/session_start", func(ctx context.Context, toolCtx *ToolContext, params json.RawMessage) (interface{}, error) {
		if toolCtx == nil {
			return nil, fmt.Errorf("unauthenticated: MCP tools require authenticated agent context")
		}

		var req struct {
			Project string `json:"project,omitempty"`
			Task    string `json:"task,omitempty"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		session, err := sessionStore.Create(ctx, toolCtx.AgentID, req.Project, req.Task)
		if err != nil {
			return nil, fmt.Errorf("start session: %w", err)
		}

		return map[string]interface{}{
			"session_id": session.ID,
			"status":     "active",
			"started_at": session.StartedAt,
		}, nil
	})

	server.RegisterHandler("continuum/session_end", func(ctx context.Context, toolCtx *ToolContext, params json.RawMessage) (interface{}, error) {
		if toolCtx == nil {
			return nil, fmt.Errorf("unauthenticated: MCP tools require authenticated agent context")
		}

		var req struct {
			SessionID string `json:"session_id"`
			Summary   string `json:"summary,omitempty"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		sessionID, err := uuid.Parse(req.SessionID)
		if err != nil {
			return nil, fmt.Errorf("invalid session_id: %w", err)
		}

		// Verify ownership
		session, err := sessionStore.GetByID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("fetch session: %w", err)
		}
		if session == nil || session.AgentID != toolCtx.AgentID {
			return nil, fmt.Errorf("session not found")
		}

		if err := sessionStore.EndSession(ctx, sessionID, req.Summary); err != nil {
			return nil, fmt.Errorf("end session: %w", err)
		}

		return map[string]string{"status": "ended"}, nil
	})
}
