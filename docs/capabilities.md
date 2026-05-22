# Continuum Capabilities

## Complete Feature Matrix

### Core Memory System

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 1 | **Episodic Memory** | Store and retrieve session transcripts with timestamps, roles, and metadata | P0 |
| 2 | **Semantic Memory** | Store entities, facts, and attributes with confidence scores | P0 |
| 3 | **Procedural Memory** | Store skills, workflows, and patterns with success tracking | P0 |
| 4 | **Memory Search** | Full-text + vector + hybrid search across all memory types | P0 |
| 5 | **Memory Graph** | Entity-relationship graph with visual exploration | P0 |
| 6 | **Memory Compression** | Auto-summarize old episodic memories to reduce token bloat | P1 |
| 7 | **Memory Importance** | Score and prioritize memories by relevance and recency | P1 |
| 8 | **Memory Conflicts** | Detect and resolve conflicting memories across agents | P2 |
| 9 | **Memory Versioning** | Track changes to semantic memory over time | P2 |
| 10 | **Memory Export/Import** | Backup and restore memory in JSON/Markdown format | P1 |

### Cross-Agent Sync

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 11 | **Real-Time WebSocket** | Live memory sync across connected agents | P0 |
| 12 | **Agent Registration** | Dynamic agent discovery and authentication | P0 |
| 13 | **Agent Presence** | Show which agents are online and active | P1 |
| 14 | **Selective Sync** | Agents subscribe to memory topics of interest | P1 |
| 15 | **Conflict Resolution** | Automatic merge strategies for concurrent writes | P2 |
| 16 | **Offline Queue** | Buffer writes when agents are disconnected | P2 |
| 17 | **Broadcast Control** | Agent can choose to broadcast or keep memory private | P1 |

### Cost Tracking & Guardrails

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 18 | **Token Counter** | Real-time token usage per request/response | P0 |
| 19 | **Cost Calculator** | Convert tokens to USD based on provider pricing | P0 |
| 20 | **Budget Per Agent** | Set daily/weekly/monthly spending caps per agent | P1 |
| 21 | **Budget Per Project** | Set spending caps per project/workspace | P1 |
| 22 | **Budget Alerts** | Webhook/Discord alerts at 50%, 80%, 100% of budget | P1 |
| 23 | **Hard Stops** | Reject requests when budget is exhausted | P2 |
| 24 | **Cost Dashboard** | Visual breakdown of spending by agent, project, time | P0 |
| 25 | **Provider Pricing** | Built-in pricing for OpenAI, Anthropic, Google, etc. | P1 |
| 26 | **Custom Pricing** | Override with custom provider pricing | P2 |

### Agent Integration

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 27 | **MCP Server** | Native Model Context Protocol server (stdio + TCP) | P0 |
| 28 | **REST API** | Full CRUD API with versioning | P0 |
| 29 | **WebSocket API** | Real-time event stream for agents | P0 |
| 30 | **Hermes Plugin** | Native Hermes Agent integration | P1 |
| 31 | **Claude Code Plugin** | Claude Code plugin via marketplace | P1 |
| 32 | **Codex Integration** | Codex CLI wrapper script | P1 |
| 33 | **OpenClaw Plugin** | OpenClaw native plugin | P1 |
| 34 | **Cursor Rules** | Cursor IDE rules file | P2 |
| 35 | **SDK (Go)** | Go client library | P1 |
| 36 | **SDK (Python)** | Python client library | P2 |
| 37 | **SDK (TypeScript)** | TypeScript/JS client library | P2 |

### Dashboard & Visualization

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 38 | **Memory Explorer** | Browse all memories with filters and search | P0 |
| 39 | **Knowledge Graph** | Interactive D3.js graph of entities and relationships | P0 |
| 40 | **Session Timeline** | Visual timeline of episodic memories | P1 |
| 41 | **Agent Activity** | Real-time view of connected agents and their actions | P1 |
| 42 | **Cost Analytics** | Charts for spending trends and projections | P0 |
| 43 | **Memory Health** | Stats on memory size, compression ratio, duplicates | P2 |
| 44 | **Dark Mode** | Native dark theme (default) | P0 |
| 45 | **Mobile Responsive** | Dashboard works on phone/tablet | P2 |

### Observability

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 46 | **Structured Logging** | JSON logs with trace IDs | P0 |
| 47 | **Metrics** | Prometheus-compatible metrics endpoint | P1 |
| 48 | **Health Checks** | /health, /ready, /live endpoints | P0 |
| 49 | **Tracing** | OpenTelemetry-compatible distributed tracing | P2 |
| 50 | **Audit Logs** | Immutable log of all mutations | P1 |

### Operations

| # | Capability | Description | Priority |
|---|-----------|-------------|----------|
| 51 | **Docker Compose** | One-command full stack deployment | P0 |
| 52 | **Single Binary** | API server as single Go binary | P0 |
| 53 | **Configuration** | Environment variables + config file | P0 |
| 54 | **Migrations** | Automatic database schema migrations | P0 |
| 55 | **Backups** | Automated backup to S3/local | P2 |
| 56 | **Multi-tenancy** | Isolate memory per workspace/team | P2 |

## Agent Connection Methods

### Method 1: MCP (Recommended)

```json
// .claude-plugin/plugin.json
{
  "name": "continuum",
  "description": "Cross-agent memory mesh for persistent context",
  "version": "1.0.0",
  "skills": ["./skills/continuum-memory"]
}
```

Agents interact via MCP tools:
- `continuum_memory_write` — Store a memory
- `continuum_memory_read` — Retrieve memories
- `continuum_memory_search` — Search across memories
- `continuum_session_start` — Begin a tracked session
- `continuum_session_end` — End session with summary

### Method 2: REST API

```bash
# Write memory
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Authorization: Bearer $CONTINUUM_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "episodic",
    "content": "Decided to use OAuth2 + PKCE for auth",
    "metadata": { "agent": "hermes", "project": "ezerops" }
  }'

# Search memories
curl "http://localhost:8080/api/v1/memories/search?q=auth+decision"
```

### Method 3: WebSocket (Real-Time)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?agent=hermes&api_key=xxx');

ws.onmessage = (event) => {
  const memory = JSON.parse(event.data);
  // Incorporate into agent context
};

// Broadcast memory to all agents
ws.send(JSON.stringify({
  action: 'memory.write',
  data: { type: 'semantic', content: '...' }
}));
```

### Method 4: Environment Wrapper

```bash
# Wrap any agent CLI to auto-sync
continuum wrap --agent codex -- codex agent:start
# All session output auto-synced to Continuum
```

## Seamless Agent Workflow Example

### Scenario: Multi-Agent Development on Ezerops

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Hermes    │     │ Claude Code │     │    Codex    │
│  (Frontend) │     │  (Feature)  │     │   (PRs)     │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Continuum  │
                    │  Memory Mesh│
                    └──────┬──────┘
                           │
       ┌───────────────────┼───────────────────┐
       ▼                   ▼                   ▼
  ┌─────────┐        ┌─────────┐        ┌─────────┐
  │Episodic │        │Semantic │        │Procedural│
  │Session  │        │Entities │        │Skills   │
  │History  │        │Facts    │        │Workflows│
  └─────────┘        └─────────┘        └─────────┘
```

**Step 1: Hermes starts work**
```
Hermes → continuum.session_start(project="ezerops", task="auth-flow")
Hermes → continuum.memory_write(type="semantic",
  content="Ezerops auth: OAuth2 + PKCE, Go backend, React frontend")
```

**Step 2: Claude Code continues**
```
Claude → continuum.session_start(project="ezerops", task="implement-oauth")
Claude → continuum.memory_read(query="ezerops auth")
← Returns: "OAuth2 + PKCE, Go backend, React frontend"
Claude → continuum.memory_write(type="procedural",
  content="OAuth2 flow: 1) Generate PKCE, 2) Redirect to provider,
           3) Exchange code, 4) Set JWT cookie")
```

**Step 3: Codex reviews PR**
```
Codex → continuum.memory_read(query="ezerops oauth implementation")
← Returns: Full context including decisions and workflow
Codex → continuum.memory_write(type="episodic",
  content="PR review: Suggested using httpOnly cookies instead of localStorage")
```

**Step 4: Hermes learns from review**
```
Hermes (WebSocket) ← Real-time: "PR review suggestion received"
Hermes → continuum.memory_read(query="latest pr review")
← Returns: Review feedback
Hermes → Updates code accordingly
```

**Result:** Each agent starts with full context. No re-explaining. No token waste.

## Memory Lifecycle

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Create    │───→│   Active    │───→│  Compress   │───→│   Archive   │
│  (write)    │    │  (hot read) │    │  (summarize)│    │  (cold)     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
      │                  │                  │                  │
      ▼                  ▼                  ▼                  ▼
   Redis +          Redis cache       pgvector only      S3 / local
   PostgreSQL       (7 days)          (30 days)          (forever)
```

## Cost Tracking Flow

```
Agent Request → Token Counter → Cost Calculator → Budget Check → Execute
                                              ↓
                                         Alert if > threshold
                                              ↓
                                         Reject if > limit
```

## Configuration

```yaml
# continuum.yaml
server:
  port: 8080
  host: 0.0.0.0

database:
  postgresql:
    url: postgres://continuum:secret@db:5432/continuum
  redis:
    url: redis://redis:6379

memory:
  max_size: 10MB          # Max memory content size
  compression:
    enabled: true
    after_days: 7         # Auto-summarize after 7 days
  retention:
    episodic_days: 30     # Keep full episodic for 30 days
    semantic_forever: true # Semantic never deleted

agents:
  authentication: api_key  # or: jwt, mTLS
  rate_limit: 1000         # requests per minute

cost:
  enabled: true
  default_budget:
    daily: 10.00          # USD
    monthly: 100.00
  providers:
    openai:
      gpt-4o: 0.005       # per 1K tokens input
    anthropic:
      claude-sonnet-4: 0.003

observability:
  log_level: info
  metrics: true
  tracing: false
```
