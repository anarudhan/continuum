---
name: continuum-memory
description: Cross-agent memory mesh integration for Continuum. Use when an agent needs to persist context, share memory with other agents, or retrieve historical decisions and facts.
license: MIT
---

# Continuum Memory Skill

Integration guide for connecting any AI agent to the Continuum memory mesh.

## When to Use

- Starting a new session and want to load previous context
- Making a decision that other agents should know about
- Learning a pattern or workflow that should be reusable
- Ending a session and want to summarize for future agents
- Searching for previous decisions, facts, or procedures

## Core Concepts

### Memory Types

| Type | Stores | Example |
|------|--------|---------|
| **episodic** | Session transcripts, decisions | "Decided OAuth2 + PKCE for auth" |
| **semantic** | Facts, entities, relationships | "Project: Anarudhan, Stack: Go + React" |
| **procedural** | Skills, workflows, patterns | "Deploy: 1) test 2) build 3) push" |

### Key Operations

```
continuum.session_start   → Begin a tracked session
continuum.memory_write    → Store a memory
continuum.memory_read     → Retrieve memories
continuum.memory_search   → Search across all memories
continuum.session_end     → End session with summary
```

## Usage Examples

### Start a Session

```
Before starting work, initialize a session:
→ continuum.session_start(
    agent="hermes",
    project="anarudhan",
    task="implement-auth-flow"
  )
```

### Write Semantic Memory

```
When you learn something permanent about the project:
→ continuum.memory_write(
    type="semantic",
    content="Anarudhan uses OAuth2 + PKCE for authentication. Go backend, React frontend.",
    entities=[
      {name: "Anarudhan", type: "project"},
      {name: "OAuth2", type: "technology"},
      {name: "PKCE", type: "technology"}
    ]
  )
```

### Write Episodic Memory

```
When you make a decision during a session:
→ continuum.memory_write(
    type="episodic",
    content="Decided to use httpOnly cookies instead of localStorage for JWT storage",
    context="PR review feedback from Codex agent"
  )
```

### Write Procedural Memory

```
When you establish a repeatable workflow:
→ continuum.memory_write(
    type="procedural",
    name="anarudhan-deploy",
    steps=[
      "Run npm run build (zero errors)",
      "Run npm audit (zero vulnerabilities)",
      "Run security scan",
      "Git commit to feature branch",
      "Open PR for Kimi audit"
    ],
    context="Frontend deployment workflow for Anarudhan"
  )
```

### Search Memory

```
Before implementing, check what we already know:
→ continuum.memory_search(
    query="authentication decisions",
    types=["episodic", "semantic"],
    limit=5
  )
```

### End Session

```
When finishing work, summarize for future agents:
→ continuum.session_end(
    summary="Implemented OAuth2 flow with PKCE. httpOnly cookies for JWT. Tests passing.",
    key_decisions=[
      "OAuth2 + PKCE over simple JWT",
      "httpOnly cookies over localStorage"
    ]
  )
```

## Best Practices

1. **Write early, write often** — Don't wait until session end
2. **Be specific** — "Use OAuth2" is better than "Fixed auth"
3. **Tag entities** — Helps build the knowledge graph
4. **Summarize on end** — Future agents read summaries first
5. **Search before deciding** — Avoid re-making decisions

## Integration Methods

### MCP (Recommended)

Install the Continuum MCP server in your agent:
```bash
# Claude Code
/plugin marketplace add anarudhan/continuum
/plugin install continuum@continuum-memory

# Or manual: add to your agent's MCP config
{
  "mcpServers": {
    "continuum": {
      "command": "continuum-mcp",
      "args": ["--api-key", "$CONTINUUM_API_KEY"]
    }
  }
}
```

### REST API

```bash
export CONTINUUM_API_KEY="your-key"
export CONTINUUM_URL="http://localhost:8080"

# Write memory
curl -X POST $CONTINUUM_URL/api/v1/memories \
  -H "Authorization: Bearer $CONTINUUM_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "semantic",
    "content": "Project uses Go 1.24 with Gin framework",
    "metadata": {"agent": "claude", "project": "anarudhan"}
  }'

# Search
curl "$CONTINUUM_URL/api/v1/memories/search?q=go+framework"
```

### WebSocket (Real-Time)

```javascript
const ws = new WebSocket(
  'ws://localhost:8080/ws?agent=hermes&api_key=xxx'
);

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('New memory from another agent:', update);
};
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Cannot connect to Continuum" | Check `docker compose ps`. Ensure continuum container is healthy. |
| "Authentication failed" | Verify `CONTINUUM_API_KEY` matches the one in docker-compose.yml. |
| "Memory not found" | Try broader search terms. Check memory type filter. |
| "Too many memories returned" | Use `limit` parameter. Add more specific `entities` filters. |
