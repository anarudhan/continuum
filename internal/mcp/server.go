package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/anarudhan/continuum/internal/models"
)

// Authenticator verifies API keys and returns agent context
type Authenticator interface {
	Authenticate(ctx context.Context, apiKey string) (*models.Agent, error)
}

// Server implements the Model Context Protocol over stdio
type Server struct {
	reader      *bufio.Reader
	writer      *bufio.Writer
	handlers    map[string]ToolHandler
	authenticator Authenticator
}

// ToolHandler handles MCP requests with authenticated context
type ToolHandler func(ctx context.Context, toolCtx *ToolContext, params json.RawMessage) (interface{}, error)

// NewServer creates a new MCP server
func NewServer(authenticator Authenticator) *Server {
	return &Server{
		reader:        bufio.NewReader(os.Stdin),
		writer:        bufio.NewWriter(os.Stdout),
		handlers:      make(map[string]ToolHandler),
		authenticator: authenticator,
	}
}

// RegisterHandler registers a handler for a method
func (s *Server) RegisterHandler(method string, handler ToolHandler) {
	s.handlers[method] = handler
}

// Request represents an MCP request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Meta    *RequestMeta    `json:"_meta,omitempty"`
}

// RequestMeta contains MCP request metadata including auth
type RequestMeta struct {
	APIKey string `json:"api_key,omitempty"`
}

// Response represents an MCP response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an MCP error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Run starts the MCP server loop
func (s *Server) Run() error {
	ctx := context.Background()

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		if req.JSONRPC != "2.0" {
			s.sendError(req.ID, -32600, "Invalid Request")
			continue
		}

		// Handle initialize method (MCP protocol handshake)
		if req.Method == "initialize" {
			s.sendResponse(req.ID, map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{},
				"serverInfo": map[string]string{
					"name":    "continuum",
					"version": "0.1.0",
				},
			})
			continue
		}

		// Handle notifications/initialized
		if req.Method == "notifications/initialized" {
			continue // No response needed for notifications
		}

		handler, ok := s.handlers[req.Method]
		if !ok {
			s.sendError(req.ID, -32601, "Method not found")
			continue
		}

		// Authenticate if API key provided in meta
		var toolCtx *ToolContext
		if req.Meta != nil && req.Meta.APIKey != "" {
			agent, err := s.authenticator.Authenticate(ctx, req.Meta.APIKey)
			if err == nil && agent != nil && agent.IsActive {
				toolCtx = &ToolContext{
					AgentID:    agent.ID,
					AgentName:  agent.Name,
					TrustLevel: agent.TrustLevel,
					Scopes:     agent.Scopes,
				}
			}
		}

		result, err := handler(ctx, toolCtx, req.Params)
		if err != nil {
			s.sendError(req.ID, -32603, err.Error())
			continue
		}

		s.sendResponse(req.ID, result)
	}
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintln(s.writer, string(data))
	s.writer.Flush()
}

func (s *Server) sendError(id interface{}, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintln(s.writer, string(data))
	s.writer.Flush()
}
