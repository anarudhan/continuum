package mcp

import "github.com/google/uuid"

// ToolContext holds the authenticated agent context for MCP tool calls
type ToolContext struct {
	AgentID    uuid.UUID
	AgentName  string
	TrustLevel string
	Scopes     []string
}
