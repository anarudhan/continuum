# 🔌 API Reference

Base URL: `http://localhost:8080/api/v1`

All endpoints require an API key in the `X-API-Key` header except `/health`, `/ready`, and `/live`.

---

## Authentication

| Header | Value |
|--------|-------|
| `X-API-Key` | `ctm_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` |

Generate a key via `POST /admin/agents` (admin only).

---

## Memories

### `POST /memories`

Create a new memory.

**Request:**
```json
{
  "content": "User prefers dark mode",
  "memory_type": "semantic",
  "tags": ["preference", "ui"],
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "User prefers dark mode",
  "memory_type": "semantic",
  "tags": ["preference", "ui"],
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_id": "550e8400-e29b-41d4-a716-446655440002",
  "visibility": "shared",
  "created_at": "2026-05-23T10:00:00Z"
}
```

### `GET /memories`

List memories with optional filters.

**Query params:**
- `type` — `episodic`, `semantic`, or `procedural`
- `agent_id` — filter by agent
- `session_id` — filter by session
- `limit` — default 50
- `offset` — default 0

**Response:**
```json
{
  "memories": [...],
  "count": 42,
  "limit": 50,
  "offset": 0
}
```

### `GET /memories/search`

Semantic search via pgvector.

**Query params:**
- `q` — search query (required)
- `type` — optional memory type filter
- `limit` — default 10

**Response:**
```json
{
  "results": [
    {
      "memory": {...},
      "score": 0.92
    }
  ],
  "query": "dark mode preference",
  "count": 3
}
```

### `GET /memories/:id`

Get a single memory by ID.

### `DELETE /memories/:id`

Archive a memory (soft delete).

---

## Sessions

### `POST /sessions`

Start a new session.

**Request:**
```json
{
  "agent_id": "550e8400-e29b-41d4-a716-446655440002",
  "metadata": {
    "project": "continuum",
    "branch": "main"
  }
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "agent_id": "550e8400-e29b-41d4-a716-446655440002",
  "status": "active",
  "started_at": "2026-05-23T10:00:00Z",
  "metadata": {
    "project": "continuum",
    "branch": "main"
  }
}
```

### `GET /sessions`

List sessions.

### `GET /sessions/:id`

Get session details. Agents can only read their own sessions.

### `POST /sessions/:id/end`

End a session. Agents can only end their own sessions.

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "status": "ended",
  "ended_at": "2026-05-23T12:00:00Z"
}
```

---

## Agents

### `GET /agents` (Admin only)

List all agents in the system.

**Response:**
```json
{
  "agents": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "hermes-frontend",
      "trust_level": "trusted",
      "scopes": ["memory:read", "memory:write"],
      "is_active": true,
      "last_seen_at": "2026-05-23T10:00:00Z"
    }
  ],
  "count": 1
}
```

### `POST /admin/agents` (Admin only)

Create a new agent. Returns the API key once.

**Request:**
```json
{
  "name": "codex-pr"
}
```

**Response:**
```json
{
  "agent": {
    "id": "550e8400-e29b-41d4-a716-446655440004",
    "name": "codex-pr",
    "trust_level": "trusted"
  },
  "api_key": "ctm_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

> ⚠️ The API key is shown **once**. Save it immediately.

---

## Health

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `GET /health` | None | Liveness check |
| `GET /ready` | None | Readiness (DB + Redis connected) |
| `GET /live` | None | Alias for `/health` |

**Response:**
```json
{
  "status": "healthy",
  "version": "0.1.0"
}
```

---

## WebSocket

Connect to `ws://localhost:8080/ws` with the `X-API-Key` header for real-time memory sync.

**Events:**
- `memory_created` — a new memory was written
- `memory_updated` — a memory was modified
- `session_started` — a new session began
- `session_ended` — a session concluded

---

## Error Format

All errors follow this shape:

```json
{
  "error": "error_code",
  "message": "Human readable description"
}
```

Common codes:

| Code | HTTP | Meaning |
|------|------|---------|
| `unauthorized` | 401 | Invalid or missing API key |
| `forbidden` | 403 | Insufficient trust level |
| `not_found` | 404 | Resource does not exist |
| `rate_limited` | 429 | Too many requests |
| `internal_error` | 500 | Something went wrong |
