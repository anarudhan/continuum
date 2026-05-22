package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ezerops/continuum/internal/core/memory"
)

// RegisterMemoryTools registers memory-related MCP tools
func RegisterMemoryTools(server *Server, memoryService *memory.Service) {
	server.RegisterHandler("continuum/memory_write", func(params json.RawMessage) (interface{}, error) {
		var req memory.WriteMemoryRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		// Agent ID would come from context/auth in real implementation
		agentID := uuid.Nil // TODO: Get from auth context

		mem, err := memoryService.Write(nil, agentID, req)
		if err != nil {
			return nil, err
		}

		return mem, nil
	})

	server.RegisterHandler("continuum/memory_search", func(params json.RawMessage) (interface{}, error) {
		var req struct {
			Query string `json:"query"`
			Type  string `json:"type,omitempty"`
			Limit int    `json:"limit,omitempty"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		agentID := uuid.Nil // TODO: Get from auth context
		memories, err := memoryService.Search(nil, agentID, req.Query, "", req.Limit)
		if err != nil {
			return nil, err
		}

		return memories, nil
	})

	server.RegisterHandler("continuum/session_start", func(params json.RawMessage) (interface{}, error) {
		var req struct {
			Project string `json:"project,omitempty"`
			Task    string `json:"task,omitempty"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		// TODO: Implement session start
		return map[string]string{"status": "started"}, nil
	})

	server.RegisterHandler("continuum/session_end", func(params json.RawMessage) (interface{}, error) {
		var req struct {
			SessionID string `json:"session_id"`
			Summary   string `json:"summary,omitempty"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}

		// TODO: Implement session end
		return map[string]string{"status": "ended"}, nil
	})
}
