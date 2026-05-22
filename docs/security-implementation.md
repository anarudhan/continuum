# Security Implementation Guide

> **Design Principle:** Security must be invisible to users. Implement it behind the scenes. Never make agents jump through hoops.

---

## 1. Authentication: Zero-Config API Keys

### Key Generation (Auto, On First Start)

```go
// internal/core/auth/keygen.go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const keyPrefix = "ctm_"

// GenerateAPIKey creates a cryptographically secure API key
// Called automatically on first docker compose up
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32) // 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return keyPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}
```

### Key Storage (Hashed, Never Plaintext)

```go
// internal/models/agent.go
package models

import (
	"crypto/sha256"
	"encoding/hex"
)

// Store only the hash. The plaintext key is shown ONCE during generation.
type Agent struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
	APIKeyHash    string    `db:"api_key_hash"`    // SHA-256 hash
	APIKeyHint    string    `db:"api_key_hint"`    // Last 4 chars for identification
	IsActive      bool      `db:"is_active"`
	CreatedAt     time.Time `db:"created_at"`
}

func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
```

### Authentication Middleware (Invisible)

```go
// internal/api/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireAPIKey(agentStore models.AgentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Accept from header or query param (for WebSocket)
		key := c.GetHeader("X-API-Key")
		if key == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if key == "" {
			key = c.Query("api_key")
		}

		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "missing_api_key",
				Message: "Set CONTINUUM_API_KEY environment variable",
			})
			return
		}

		agent, err := agentStore.GetByAPIKeyHash(c, HashAPIKey(key))
		if err != nil || agent == nil || !agent.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "invalid_api_key",
				Message: "Check your CONTINUUM_API_KEY",
			})
			return
		}

		// Set agent context — downstream handlers use this, not the key
		c.Set("agent_id", agent.ID)
		c.Set("agent_name", agent.Name)
		c.Next()
	}
}
```

**User experience:**
```bash
# One-time setup
docker compose up -d
# Logs show: "Generated API key: ctm_xxxxxxxx..."
# Copy to agent env, done. Never think about it again.
```

---

## 2. Rate Limiting: Invisible Protection

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

// RateLimiter — high limits that normal agents never hit
type RateLimiter struct {
	redis *redis.Client
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{redis: redis}
}

func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("agent_id")
		if agentID == nil {
			c.Next()
			return
		}

		key := fmt.Sprintf("rl:%s:%d", agentID, time.Now().Minute())
		
		// Increment counter
		count, err := r.redis.Incr(c, key).Result()
		if err != nil {
			c.Next() // Fail open — don't block on Redis error
			return
		}
		
		if count == 1 {
			r.redis.Expire(c, key, time.Minute)
		}

		// 1000 req/min — generous for any legitimate agent workflow
		if count > 1000 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorResponse{
				Error:   "rate_limit",
				Message: "Slow down — you're making too many requests",
			})
			return
		}

		c.Next()
	}
}
```

**User experience:** Normal agents never see rate limit errors. Only abusive patterns trigger this.

---

## 3. Content Safety: Preserve Everything, Render Safely

```go
// internal/api/validation/memory.go
package validation

import (
	"strings"
	"unicode/utf8"
)

const MaxMemorySize = 10 * 1024 * 1024 // 10MB

// ValidateMemoryInput — minimal checks, maximum preservation
func ValidateMemoryInput(content string) error {
	// Size check only
	if utf8.RuneCountInString(content) > MaxMemorySize {
		return fmt.Errorf("memory too large (max 10MB)")
	}
	
	// Null bytes break PostgreSQL text fields
	if strings.Contains(content, "\x00") {
		return fmt.Errorf("null bytes not allowed")
	}
	
	return nil
}
```

**Dashboard rendering (React):**
```tsx
// web/src/components/MemoryCard.tsx
function MemoryCard({ memory }: { memory: Memory }) {
  return (
    <div className="memory-card">
      {/* Render as PRE/CODE — never as HTML */}
      <pre className="whitespace-pre-wrap font-mono text-sm">
        {memory.content}
      </pre>
      
      {/* Metadata shown as plain text */}
      <div className="text-xs text-gray-500">
        From: {memory.agent_name} • {formatDate(memory.created_at)}
      </div>
    </div>
  );
}
```

**What we DON'T do:**
- ❌ Strip `<script>` tags (breaks code snippets)
- ❌ Block "dangerous" words (false positives)
- ❌ Reject HTML content (legitimate docs have HTML)

**What we DO:**
- ✅ Render everything as plain text / code blocks
- ✅ CSP headers prevent execution
- ✅ Size limits prevent storage abuse

---

## 4. Database Security: Parameterized Queries

```go
// internal/models/memory.go
package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MemoryStore struct {
	db *pgxpool.Pool
}

func (s *MemoryStore) Search(ctx context.Context, agentID, query string, limit int) ([]Memory, error) {
	// ALWAYS use $N placeholders — NEVER string concatenation
	rows, err := s.db.Query(ctx, `
		SELECT id, type, content, agent_id, created_at, visibility
		FROM memories
		WHERE (
			visibility != 'private' 
			OR agent_id = $1
		)
		AND to_tsvector('english', content) @@ plainto_tsquery('english', $2)
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

## 5. Memory Isolation: Automatic, Zero Config

```go
// internal/core/memory/service.go
package memory

import "context"

type MemoryService struct {
	store MemoryStore
}

func (s *MemoryService) Read(ctx context.Context, agentID string, memoryID string) (*Memory, error) {
	memory, err := s.store.GetByID(ctx, memoryID)
	if err != nil {
		return nil, err
	}
	
	// Automatic isolation — no user configuration needed
	if memory.Visibility == "private" && memory.AgentID != agentID {
		return nil, ErrNotFound // 404, not 403 — don't leak existence
	}
	
	return memory, nil
}

func (s *MemoryService) Write(ctx context.Context, agentID string, req WriteRequest) (*Memory, error) {
	// Default visibility if not specified
	visibility := req.Visibility
	if visibility == "" {
		visibility = "shared" // Default: share with all agents
	}
	
	memory := &Memory{
		Type:       req.Type,
		Content:    req.Content,
		AgentID:    agentID,
		Visibility: visibility,
		CreatedAt:  time.Now(),
	}
	
	return s.store.Create(ctx, memory)
}
```

**Agent API (unchanged from agent's perspective):**
```python
# Agent writes — same API regardless of visibility
continuum.memory_write(
    type="semantic",
    content="OAuth2 decision",
    # visibility is optional, defaults to shared
)
```

---

## 6. Audit Logging: Silent, Immutable

```go
// internal/observability/audit.go
package audit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Auditor struct {
	db *pgxpool.Pool
}

func (a *Auditor) Log(ctx context.Context, eventType string, agentID string, details map[string]interface{}) {
	// Fire and forget — never block agent operations
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		a.db.Exec(ctx, `
			INSERT INTO audit_log (timestamp, type, agent_id, details)
			VALUES ($1, $2, $3, $4)
		`, time.Now().UTC(), eventType, agentID, details)
	}()
}
```

**User experience:** Agents never wait for audit logging. It happens in the background.

---

## 7. WebSocket Security: Transparent

```go
// internal/websocket/server.go
package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow same-origin and localhost (development)
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Non-browser client
		}
		// In production, check against configured origins
		return isAllowedOrigin(origin)
	},
}

func isAllowedOrigin(origin string) bool {
	allowed := []string{
		"http://localhost",
		"https://localhost",
		// Add user's domains here via config
	}
	for _, a := range allowed {
		if strings.HasPrefix(origin, a) {
			return true
		}
	}
	return false
}
```

---

## 8. Docker Security: Minimal Attack Surface

```dockerfile
# deployments/docker/Dockerfile
# Multi-stage build with distroless final image

FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o continuum ./cmd/continuum

# Distroless: no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /build/continuum /app/continuum
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/app/continuum", "healthcheck"]
ENTRYPOINT ["/app/continuum"]
```

---

## 9. Security Testing: Automated, Not Manual

```go
// tests/security/auth_test.go
package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth_ValidKey(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/v1/memories", nil)
	req.Header.Set("X-API-Key", "ctm_validtestkey")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_MissingKey(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/v1/memories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidKey(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/v1/memories", nil)
	req.Header.Set("X-API-Key", "wrong_key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSQLInjection_Prevention(t *testing.T) {
	injections := []string{
		"'; DROP TABLE memories; --",
		"1' OR '1'='1",
		"test' UNION SELECT * FROM agents --",
	}

	for _, injection := range injections {
		memories, err := store.Search(ctx, "agent-123", injection, 10)
		assert.NoError(t, err) // Should not error
		assert.Empty(t, memories) // Should return no results
	}
}

func TestMemoryIsolation_Private(t *testing.T) {
	// Agent A creates private memory
	memory := createMemory(t, "agent-a", "private", "secret")
	
	// Agent B tries to read — should get 404 (not 403)
	req := httptest.NewRequest("GET", "/api/v1/memories/"+memory.ID, nil)
	req.Header.Set("X-API-Key", "ctm_agent_b")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}
```

---

## 10. Incident Response: Automated

```go
// internal/observability/alerts.go
package alerts

import (
	"context"
	"fmt"
	"time"
)

type AlertManager struct {
	discordWebhook string
}

func (a *AlertManager) Send(ctx context.Context, severity string, message string) {
	// Only alert on actual anomalies
	if severity == "critical" || severity == "high" {
		go a.sendDiscord(ctx, fmt.Sprintf("[%s] %s", severity, message))
	}
	
	// Medium/low go to audit log only
	audit.Log(ctx, "alert", "system", map[string]interface{}{
		"severity": severity,
		"message":  message,
	})
}
```

---

## Security Implementation Checklist (Internal)

Before ANY feature ships:

- [ ] Agent can use feature with zero config beyond API key
- [ ] Agent never sees a security error during normal use
- [ ] Agent never needs to retry due to rate limits (normal load)
- [ ] Memory write succeeds on first attempt (legitimate content)
- [ ] Dashboard shows security status but requires no action
- [ ] Alerts only fire for actual anomalies
- [ ] Key rotation happens without agent restart
- [ ] TLS works without user providing certificates
- [ ] All security tests pass
- [ ] No hardcoded secrets

---

## Key Principles (Reiterated)

1. **Secure by default** — Works safely out of the box
2. **Zero friction** — Security never blocks legitimate workflows
3. **Transparent** — Users can see protection, don't need to manage it
4. **Fail open** — If security system fails, agents keep working (degraded security, not downtime)
5. **Minimal config** — One env var (API key), everything else automatic
