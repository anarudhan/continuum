package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Server implements the Model Context Protocol over stdio
type Server struct {
	reader *bufio.Reader
	writer *bufio.Writer
	handlers map[string]Handler
}

// Handler handles MCP requests
type Handler func(params json.RawMessage) (interface{}, error)

// NewServer creates a new MCP server
func NewServer() *Server {
	return &Server{
		reader:   bufio.NewReader(os.Stdin),
		writer:   bufio.NewWriter(os.Stdout),
		handlers: make(map[string]Handler),
	}
}

// RegisterHandler registers a handler for a method
func (s *Server) RegisterHandler(method string, handler Handler) {
	s.handlers[method] = handler
}

// Request represents an MCP request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
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

		handler, ok := s.handlers[req.Method]
		if !ok {
			s.sendError(req.ID, -32601, "Method not found")
			continue
		}

		result, err := handler(req.Params)
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
