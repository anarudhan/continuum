package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow same-origin and localhost
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		return isAllowedOrigin(origin)
	},
}

func isAllowedOrigin(origin string) bool {
	allowed := []string{"http://localhost", "https://localhost"}
	for _, a := range allowed {
		if len(origin) >= len(a) && origin[:len(a)] == a {
			return true
		}
	}
	return false
}

// Message represents a WebSocket message
type Message struct {
	Action    string          `json:"action"`
	Data      json.RawMessage `json:"data"`
	AgentID   string          `json:"agent_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Server manages WebSocket connections
type Server struct {
	clients    map[string]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a WebSocket client
type Client struct {
	server *Server
	conn   *websocket.Conn
	send   chan Message
	agentID uuid.UUID
}

// NewServer creates a new WebSocket server
func NewServer() *Server {
	return &Server{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the WebSocket server
func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.agentID.String()] = client
			s.mu.Unlock()

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.agentID.String()]; ok {
				delete(s.clients, client.agentID.String())
				close(client.send)
			}
			s.mu.Unlock()

		case message := <-s.broadcast:
			s.mu.RLock()
			for _, client := range s.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full, skip
				}
			}
			s.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(msg Message) {
	select {
	case s.broadcast <- msg:
	default:
		// Broadcast channel full, drop message
	}
}

// HandleConnection handles WebSocket upgrade and client management
func (s *Server) HandleConnection(c *gin.Context) {
	agentID, exists := c.Get("agent_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		server:  s,
		conn:    conn,
		send:    make(chan Message, 256),
		agentID: agentID.(uuid.UUID),
	}

	s.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}

		msg.AgentID = c.agentID.String()
		msg.Timestamp = time.Now()

		// Process message based on action
		c.handleMessage(msg)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages
func (c *Client) handleMessage(msg Message) {
	switch msg.Action {
	case "memory.write":
		// Broadcast to all clients (including sender for confirmation)
		c.server.Broadcast(msg)
	case "memory.search":
		// Handle search request
		// This would typically query the database and return results
	default:
		// Unknown action, ignore
	}
}

// GetClientCount returns the number of connected clients
func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}
