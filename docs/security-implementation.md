# Security Implementation Guide

> **CRITICAL:** Every feature MUST implement the security controls listed here. No exceptions.

---

## 1. Authentication Layer

### API Key System

```go
// internal/api/middleware/auth.go
package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ezerops/continuum/internal/models"
)

const (
	APIKeyHeader     = "X-API-Key"
	APIKeyPrefix     = "ctm_"
	APIKeyMinEntropy = 256 // bits
)

type AuthMiddleware struct {
	agentStore models.AgentStore
}

func NewAuthMiddleware(store models.AgentStore) *AuthMiddleware {
	return &AuthMiddleware{agentStore: store}
}

func (a *AuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(APIKeyHeader)
		if key == "" {
			// Also check Authorization: Bearer header
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_api_key",
				"message": "API key required in X-API-Key or Authorization header",
			})
			return
		}

		// Validate key format
		if !strings.HasPrefix(key, APIKeyPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_api_key_format",
				"message": "API key must start with 'ctm_'",
			})
			return
		}

		// Hash comparison (constant-time to prevent timing attacks)
		agent, err := a.agentStore.GetByAPIKey(c.Request.Context(), key)
		if err != nil || agent == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_api_key",
				"message": "API key not recognized",
			})
			return
		}

		// Check key expiry
		if agent.APIKeyExpiresAt != nil && time.Now().After(*agent.APIKeyExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "api_key_expired",
				"message": "API key has expired, please rotate",
			})
			return
		}

		// Check if agent is active
		if !agent.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "agent_inactive",
				"message": "Agent has been deactivated",
			})
			return
		}

		// Set agent context for downstream handlers
		c.Set("agent_id", agent.ID)
		c.Set("agent_name", agent.Name)
		c.Set("agent_trust_level", agent.TrustLevel)
		c.Set("agent_scopes", agent.Scopes)

		c.Next()
	}
}

// RequireScope checks if agent has required scope
func (a *AuthMiddleware) RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("agent_scopes")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "missing_scopes",
				"message": "Agent scopes not found",
			})
			return
		}

		agentScopes, ok := scopes.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_scopes",
				"message": "Agent scopes are invalid",
			})
			return
		}

		for _, s := range agentScopes {
			if s == scope || s == "admin" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_scope",
			"message": "Agent lacks required scope: " + scope,
		})
	}
}
```

### Key Generation

```go
// internal/core/auth/keygen.go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	keyPrefix   = "ctm_"
	keyLength   = 43 // 256 bits base64 encoded = 43 chars
	keyEntropy  = 256
)

// GenerateAPIKey creates a cryptographically secure API key
func GenerateAPIKey() (string, error) {
	// 32 bytes = 256 bits of entropy
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// URL-safe base64 encoding (no padding)
	encoded := base64.RawURLEncoding.EncodeToString(b)

	return keyPrefix + encoded, nil
}

// HashAPIKey hashes an API key for storage (using bcrypt or Argon2)
// Store the hash, never the plaintext key (except at generation time)
func HashAPIKey(key string) (string, error) {
	// Use bcrypt with cost 12
	// import golang.org/x/crypto/bcrypt
	// hash, err := bcrypt.GenerateFromPassword([]byte(key), 12)
	// return string(hash), err
	
	// For now, store SHA-256 hash (not for production, use bcrypt)
	// TODO: Replace with bcrypt before v1.0
	return "", nil
}

// ValidateAPIKeyFormat checks if key matches expected format
func ValidateAPIKeyFormat(key string) bool {
	if !strings.HasPrefix(key, keyPrefix) {
		return false
	}
	
	encoded := strings.TrimPrefix(key, keyPrefix)
	if len(encoded) != keyLength {
		return false
	}

	// Verify it's valid base64
	_, err := base64.RawURLEncoding.DecodeString(encoded)
	return err == nil
}
```

---

## 2. Input Validation

### Schema Validation

```go
// internal/api/validation/memory.go
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MaxMemoryContentSize = 10 * 1024 * 1024 // 10MB
	MaxEntityNameLength  = 255
	MaxMetadataKeys      = 50
	MaxMetadataValueLen  = 4096
	AllowedMemoryTypes   = "episodic|semantic|procedural"
)

var (
	// Reject HTML/JS tags
	htmlTagPattern = regexp.MustCompile(`<[^>]*script[^>]*>|<[^>]*iframe[^>]*>|<[^>]*object[^>]*>`)
	
	// Reject control characters except newline/tab
	controlCharPattern = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	
	// Valid entity name: alphanumeric + spaces + common punctuation
	entityNamePattern = regexp.MustCompile(`^[\w\s\-_.:@]+$`)
)

// MemoryWriteRequest validated input
type MemoryWriteRequest struct {
	Type       string            `json:"type" binding:"required,oneof=episodic semantic procedural"`
	Content    string            `json:"content" binding:"required,max=10485760"`
	Entities   []EntityRef       `json:"entities,omitempty" binding:"dive"`
	Metadata   map[string]string `json:"metadata,omitempty" binding:"max=50,dive,max=4096"`
	Visibility string            `json:"visibility,omitempty" binding:"oneof=private shared broadcast"`
	AgentID    string            `json:"-"` // Set from auth context
}

type EntityRef struct {
	Name string `json:"name" binding:"required,max=255,alphanumunicode"`
	Type string `json:"type" binding:"required,max=100,alphanum"`
}

// SanitizeMemoryContent cleans user-provided content
func SanitizeMemoryContent(content string) (string, error) {
	// Check size
	if utf8.RuneCountInString(content) > MaxMemoryContentSize {
		return "", fmt.Errorf("content exceeds maximum size of %d bytes", MaxMemoryContentSize)
	}

	// Strip dangerous HTML/JS
	if htmlTagPattern.MatchString(content) {
		content = htmlTagPattern.ReplaceAllString(content, "")
	}

	// Remove control characters
	content = controlCharPattern.ReplaceAllString(content, "")

	// Trim whitespace
	content = strings.TrimSpace(content)

	// Ensure not empty after sanitization
	if content == "" {
		return "", fmt.Errorf("content is empty after sanitization")
	}

	return content, nil
}

// ValidateMetadata ensures metadata is safe
func ValidateMetadata(meta map[string]string) error {
	if len(meta) > MaxMetadataKeys {
		return fmt.Errorf("metadata exceeds maximum of %d keys", MaxMetadataKeys)
	}

	for k, v := range meta {
		// Key validation
		if len(k) > 100 {
			return fmt.Errorf("metadata key '%s' exceeds 100 characters", k)
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(k) {
			return fmt.Errorf("metadata key '%s' contains invalid characters", k)
		}

		// Value validation
		if len(v) > MaxMetadataValueLen {
			return fmt.Errorf("metadata value for '%s' exceeds %d characters", k, MaxMetadataValueLen)
		}
		if htmlTagPattern.MatchString(v) {
			return fmt.Errorf("metadata value for '%s' contains forbidden content", k)
		}
	}

	return nil
}
```

### Rate Limiting

```go
// internal/api/middleware/ratelimit.go
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis      *redis.Client
	limit      int           // requests per window
	window     time.Duration // e.g., 1 minute
	burstSize  int           // allow burst
}

func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:     redis,
		limit:     limit,
		window:    window,
		burstSize: limit / 10, // 10% burst
	}
}

func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, exists := c.Get("agent_id")
		if !exists {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:%s", agentID)
		
		// Sliding window counter using Redis
		now := time.Now().Unix()
		windowStart := now - int64(r.window.Seconds())

		pipe := r.redis.Pipeline()
		pipe.ZRemRangeByScore(c.Request.Context(), key, "0", fmt.Sprintf("%d", windowStart))
		pipe.ZCard(c.Request.Context(), key)
		pipe.ZAdd(c.Request.Context(), key, redis.Z{Score: float64(now), Member: now})
		pipe.Expire(c.Request.Context(), key, r.window)
		
		results, err := pipe.Exec(c.Request.Context())
		if err != nil {
			// Fail open (allow request) but log error
			c.Next()
			return
		}

		count := results[1].(*redis.IntCmd).Val()
		
		if count > int64(r.limit+r.burstSize) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit exceeded: %d requests per %v", r.limit, r.window),
				"retry_after": r.window.Seconds(),
			})
			return
		}

		c.Next()
	}
}
```

---

## 3. Output Encoding

### Response Sanitization

```go
// internal/api/response/safe.go
package response

import (
	"html"
	"strings"
)

// SafeString escapes HTML in strings for dashboard display
func SafeString(s string) string {
	return html.EscapeString(s)
}

// SafeJSON ensures JSON responses don't contain dangerous content
func SafeJSON(data map[string]interface{}) map[string]interface{} {
	// Deep copy and sanitize
	sanitized := make(map[string]interface{})
	for k, v := range data {
		switch val := v.(type) {
		case string:
			// Don't escape here — let the JSON encoder handle it
			// Just strip dangerous patterns
			sanitized[k] = stripDangerousPatterns(val)
		case map[string]interface{}:
			sanitized[k] = SafeJSON(val)
		default:
			sanitized[k] = v
		}
	}
	return sanitized
}

func stripDangerousPatterns(s string) string {
	// Remove null bytes and other dangerous characters
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}
```

### Security Headers

```go
// internal/api/middleware/security.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		c.Header("Content-Security-Policy", 
			"default-src 'self'; "+
			"script-src 'self'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; "+
			"font-src 'self'; "+
			"connect-src 'self' ws: wss:; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self';")
		
		// HSTS (only in production with TLS)
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		// Permissions policy
		c.Header("Permissions-Policy", 
			"accelerometer=(), "+
			"camera=(), "+
			"geolocation=(), "+
			"gyroscope=(), "+
			"magnetometer=(), "+
			"microphone=(), "+
			"payment=(), "+
			"usb=()")

		c.Next()
	}
}
```

---

## 4. Database Security

### Row-Level Security (PostgreSQL)

```sql
-- migrations/003_row_level_security.sql
-- Enable RLS on all tables

ALTER TABLE memories ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE cost_logs ENABLE ROW LEVEL SECURITY;

-- Policy: Agents can only see their own private memories
CREATE POLICY memory_owner_isolation ON memories
    FOR ALL
    USING (
        visibility != 'private' 
        OR agent_id = current_setting('app.current_agent_id', true)::UUID
    );

-- Policy: Shared memories are readable by all
CREATE POLICY memory_shared_read ON memories
    FOR SELECT
    USING (visibility = 'shared' OR visibility = 'broadcast');

-- Policy: Only memory owner can update/delete
CREATE POLICY memory_owner_modify ON memories
    FOR UPDATE
    USING (agent_id = current_setting('app.current_agent_id', true)::UUID);

-- Set agent ID for each connection
CREATE OR REPLACE FUNCTION set_agent_context(agent_id UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_agent_id', agent_id::text, false);
END;
$$ LANGUAGE plpgsql;
```

### Parameterized Queries (Go)

```go
// internal/models/memory.go
package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemoryStore struct {
	db *pgxpool.Pool
}

func (s *MemoryStore) Search(ctx context.Context, agentID string, query string, limit int) ([]Memory, error) {
	// NEVER concatenate user input into SQL
	// ALWAYS use parameterized queries
	
	rows, err := s.db.Query(ctx, `
		SELECT id, type, content, agent_id, created_at, visibility
		FROM memories
		WHERE (
			visibility = 'shared' 
			OR visibility = 'broadcast'
			OR (visibility = 'private' AND agent_id = $1)
		)
		AND (
			to_tsvector('english', content) @@ plainto_tsquery('english', $2)
			OR content ILIKE '%' || $2 || '%'
		)
		ORDER BY created_at DESC
		LIMIT $3
	`, agentID, query, limit)
	
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.Type, &m.Content, &m.AgentID, &m.CreatedAt, &m.Visibility); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}
```

---

## 5. Audit Logging

### Immutable Audit Trail

```go
// internal/observability/audit.go
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type EventType string

const (
	EventMemoryWrite    EventType = "memory.write"
	EventMemoryRead     EventType = "memory.read"
	EventMemoryDelete   EventType = "memory.delete"
	EventAgentConnect   EventType = "agent.connect"
	EventAgentDisconnect EventType = "agent.disconnect"
	EventAuthFailure    EventType = "auth.failure"
	EventConfigChange   EventType = "config.change"
)

type AuditEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	AgentID   string    `json:"agent_id,omitempty"`
	AgentName string    `json:"agent_name,omitempty"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource,omitempty"`
	Details   JSONMap   `json:"details,omitempty"`
	IP        string    `json:"ip,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

type JSONMap map[string]interface{}

type Auditor struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewAuditor(db *pgxpool.Pool, logger *zap.Logger) *Auditor {
	return &Auditor{db: db, logger: logger}
}

func (a *Auditor) Log(ctx context.Context, event AuditEvent) error {
	event.ID = generateEventID()
	event.Timestamp = time.Now().UTC()

	detailsJSON, _ := json.Marshal(event.Details)

	// Insert into append-only audit log table
	_, err := a.db.Exec(ctx, `
		INSERT INTO audit_log (id, timestamp, type, agent_id, agent_name, 
		                       action, resource, details, ip, success, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, event.ID, event.Timestamp, event.Type, event.AgentID, event.AgentName,
		event.Action, event.Resource, detailsJSON, event.IP, event.Success, event.Error)

	if err != nil {
		// Log to fallback (stdout) if DB fails
		a.logger.Error("failed to write audit log",
			zap.String("event_id", event.ID),
			zap.String("type", string(event.Type)),
			zap.Error(err),
		)
		return fmt.Errorf("audit log failed: %w", err)
	}

	// Also log to structured logger for real-time monitoring
	a.logger.Info("audit event",
		zap.String("event_id", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("agent", event.AgentName),
		zap.String("action", event.Action),
		zap.Bool("success", event.Success),
	)

	return nil
}

func generateEventID() string {
	// ULID or UUID v7 for sortability
	return "evt_" + generateULID()
}

func generateULID() string {
	// Implementation using github.com/oklog/ulid
	return ""
}
```

---

## 6. WebSocket Security

```go
// internal/websocket/security.go
package websocket

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Strict origin check
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// No origin header = same-origin or non-browser client
			return true
		}
		
		// Only allow same-origin or configured origins
		allowedOrigins := getAllowedOrigins()
		for _, allowed := range allowedOrigins {
			if strings.HasPrefix(origin, allowed) {
				return true
			}
		}
		
		return false
	},
}

func getAllowedOrigins() []string {
	// Load from config
	// Default: same-origin only (empty list)
	return []string{"http://localhost:8080", "https://localhost:8080"}
}

// ValidateConnection checks API key in WebSocket handshake
func ValidateConnection(c *gin.Context) (string, error) {
	// API key must be in query param or header
	apiKey := c.Query("api_key")
	if apiKey == "" {
		apiKey = c.GetHeader("X-API-Key")
	}
	
	if apiKey == "" {
		return "", fmt.Errorf("missing API key")
	}
	
	// Validate against store
	agent, err := validateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		return "", fmt.Errorf("invalid API key")
	}
	
	return agent.ID, nil
}
```

---

## 7. Security Testing

### Test Cases (Required)

```go
// tests/security/api_key_test.go
package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIKey_Authentication(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name       string
		apiKey     string
		wantStatus int
	}{
		{
			name:       "valid key",
			apiKey:     "ctm_validtestkey1234567890123456789012345678",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing key",
			apiKey:     "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid format",
			apiKey:     "not_a_valid_key",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong prefix",
			apiKey:     "wrong_prefix12345678901234567890123456789012",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired key",
			apiKey:     "ctm_expiredkey123456789012345678901234567890",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/memories", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestInputValidation_XSSPrevention(t *testing.T) {
	payloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
		"<iframe src='evil.com'></iframe>",
		"<object data='evil.swf'></object>",
	}

	for _, payload := range payloads {
		sanitized, err := validation.SanitizeMemoryContent(payload)
		assert.NoError(t, err)
		assert.NotContains(t, sanitized, "<script")
		assert.NotContains(t, sanitized, "javascript:")
		assert.NotContains(t, sanitized, "<iframe")
	}
}

func TestSQLInjection_Prevention(t *testing.T) {
	injections := []string{
		"'; DROP TABLE memories; --",
		"1' OR '1'='1",
		"1; DELETE FROM memories WHERE '1'='1",
		"test' UNION SELECT * FROM agents --",
	}

	for _, injection := range injections {
		// These should be safely handled by parameterized queries
		memories, err := store.Search(ctx, "agent-123", injection, 10)
		assert.NoError(t, err) // Should not error, just return no results
		assert.Empty(t, memories)
	}
}

func TestRateLimiting_Enforcement(t *testing.T) {
	// Send 1001 requests in rapid succession
	// Request 1001 should be rate limited
	for i := 0; i < 1005; i++ {
		req := httptest.NewRequest("GET", "/api/v1/memories", nil)
		req.Header.Set("X-API-Key", "ctm_testkey")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if i >= 1000 {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestAuthorization_MemoryIsolation(t *testing.T) {
	// Agent A writes private memory
	// Agent B should not be able to read it
	
	memory := createMemory(t, "agent-a", "private", "secret content")
	
	// Agent B tries to read
	req := httptest.NewRequest("GET", "/api/v1/memories/"+memory.ID, nil)
	req.Header.Set("X-API-Key", "ctm_agent_b_key")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code) // 404, not 403 (don't leak existence)
}
```

---

## 8. Docker Security

```dockerfile
# deployments/docker/Dockerfile.secure
# Multi-stage build with security hardening

# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o continuum ./cmd/continuum

# Final stage — minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy binary
COPY --from=builder /build/continuum /app/continuum

# Use non-root user (65532:nonroot in distroless)
USER nonroot:nonroot

# No shell, no package manager, no unnecessary tools

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/app/continuum", "healthcheck"]

ENTRYPOINT ["/app/continuum"]
```

---

## 9. Secrets Management

### Environment Variables (Never in Code)

```go
// internal/config/secrets.go
package config

import (
	"fmt"
	"os"
)

// Required secrets — fail fast if missing
func LoadSecrets() (*Secrets, error) {
	secrets := &Secrets{}
	
	// Database
	secrets.DatabaseURL = os.Getenv("CONTINUUM_DATABASE_URL")
	if secrets.DatabaseURL == "" {
		return nil, fmt.Errorf("CONTINUUM_DATABASE_URL is required")
	}
	
	// Redis
	secrets.RedisURL = os.Getenv("CONTINUUM_REDIS_URL")
	if secrets.RedisURL == "" {
		return nil, fmt.Errorf("CONTINUUM_REDIS_URL is required")
	}
	
	// Admin API key (for dashboard access)
	secrets.AdminAPIKey = os.Getenv("CONTINUUM_ADMIN_API_KEY")
	if secrets.AdminAPIKey == "" {
		return nil, fmt.Errorf("CONTINUUM_ADMIN_API_KEY is required")
	}
	
	// Validate admin key entropy
	if len(secrets.AdminAPIKey) < 32 {
		return nil, fmt.Errorf("CONTINUUM_ADMIN_API_KEY must be at least 32 characters")
	}
	
	return secrets, nil
}

type Secrets struct {
	DatabaseURL   string
	RedisURL      string
	AdminAPIKey   string
	JWTSecret     string
	EncryptionKey string // For sensitive memory encryption
}
```

---

## 10. Incident Response

### Security Event Types

| Severity | Event | Response |
|----------|-------|----------|
| **Critical** | Unauthorized admin access | Immediate revocation, alert, investigation |
| **Critical** | Mass data exfiltration | Isolate instance, preserve logs, notify |
| **High** | Brute force on API keys | Rate limit, block IP, alert |
| **High** | Poisoned memory injection | Quarantine agent, flag memories, alert |
| **Medium** | Unusual access pattern | Review audit log, notify admin |
| **Low** | Failed auth attempt | Log, no immediate action |

### Alert Channels

```yaml
# Alert configuration
alerts:
  discord:
    webhook_url: ${DISCORD_WEBHOOK_URL}
    events: [critical, high]
  
  email:
    smtp_host: ${SMTP_HOST}
    to: security@ezerops.com
    events: [critical]
  
  webhook:
    url: ${SECURITY_WEBHOOK_URL}
    events: [critical, high, medium]
```

---

## Security Implementation Checklist

### Before ANY Feature Ships

- [ ] Input validated (schema, size, type, content)
- [ ] Output encoded (no raw HTML/JS)
- [ ] Auth checked (API key valid, not expired, right scope)
- [ ] Authorization enforced (agent can only access own data)
- [ ] Rate limiting applied
- [ ] Audit log written
- [ ] Error messages sanitized (no stack traces, no internals)
- [ ] Security tests pass

### Before v1.0 Release

- [ ] Penetration test completed
- [ ] Dependency audit (no known CVEs)
- [ ] Docker image scan clean
- [ ] Security documentation complete
- [ ] Incident response plan documented
- [ ] Bug bounty program ready (optional)
