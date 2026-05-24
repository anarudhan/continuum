# 🤝 Agent Integration


> **Brand:** This document follows the [Continuum Circuit Board Brand System](branding/BRAND_SYSTEM.md).
> Colors: Deep Navy `#0A1628`, Electric Cyan `#00D4FF`, Soft Lavender `#A78BFA`.


Connect any agent to Continuum in under 5 minutes.

---

## Quick Start

### 1. Get an API key

```bash
curl -X POST http://localhost:8080/admin/agents \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent"}'
```

Save the returned `api_key`. You will never see it again.

### 2. Write a memory

```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "X-API-Key: $YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "User prefers tabs over spaces",
    "memory_type": "semantic",
    "tags": ["preference", "coding-style"]
  }'
```

### 3. Search memories

```bash
curl "http://localhost:8080/api/v1/memories/search?q=coding+style&limit=5" \
  -H "X-API-Key: $YOUR_KEY"
```

---

## Integration Patterns

### Pattern A: REST polling (simplest)

Read and write memories via HTTP on every turn.

```python
import requests

API_KEY = "ctm_..."
BASE = "http://localhost:8080/api/v1"

def remember(content, memory_type="semantic", tags=None):
    requests.post(f"{BASE}/memories", headers={
        "X-API-Key": API_KEY
    }, json={
        "content": content,
        "memory_type": memory_type,
        "tags": tags or []
    })

def recall(query, limit=5):
    r = requests.get(f"{BASE}/memories/search", headers={
        "X-API-Key": API_KEY
    }, params={"q": query, "limit": limit})
    return r.json()["results"]
```

### Pattern B: WebSocket real-time (recommended)

Connect once, get push updates for every memory event.

```javascript
import WebSocket from 'ws';

const ws = new WebSocket('ws://localhost:8080/ws', {
  headers: { 'X-API-Key': 'ctm_...' }
});

ws.on('message', (data) => {
  const event = JSON.parse(data);
  if (event.type === 'memory_created') {
    console.log('New memory:', event.payload.content);
  }
});
```

### Pattern C: MCP native (Claude Code, Codex)

Use the built-in MCP server for zero-code integration.

```json
{
  "mcpServers": {
    "continuum": {
      "command": "continuum",
      "args": ["mcp"]
    }
  }
}
```

Tools exposed:
- `memory_write` — persist a memory
- `memory_search` — semantic recall
- `session_start` / `session_end` — lifecycle tracking

---

## Agent-Specific Guides

### Claude Code

Add to `.claude-plugin/plugin.json`:

```json
{
  "id": "continuum",
  "name": "Continuum Memory",
  "tools": ["memory_write", "memory_search"],
  "endpoint": "http://localhost:8080"
}
```

### Hermes

Use the bundled skill:

```yaml
skills:
  - continuum-memory
```

Environment:
```bash
export CONTINUUM_URL=http://localhost:8080
export CONTINUUM_API_KEY=ctm_...
```

### Codex

Enable via CLI flag:

```bash
codex --mcp continuum
```

Or set in `~/.codex/config.json`:

```json
{
  "mcpServers": ["continuum"]
}
```

---

## Best Practices

| Do | Don't |
|----|-------|
| Tag every memory with context | Dump raw transcripts without structure |
| Start a session per task | Reuse one session for everything |
| Search before writing to avoid duplicates | Blindly write without checking |
| Use `semantic` for facts, `episodic` for events, `procedural` for skills | Put everything in one bucket |
| Set `visibility: private` for sensitive data | Leak secrets into shared memory |

---

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `CONTINUUM_URL` | `http://localhost:8080` | API base URL |
| `CONTINUUM_API_KEY` | — | Your agent's API key |
| `CONTINUUM_WS_URL` | `ws://localhost:8080/ws` | WebSocket endpoint |
