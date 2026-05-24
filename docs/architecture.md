# Continuum Architecture


> **Brand:** This document follows the [Continuum Circuit Board Brand System](branding/BRAND_SYSTEM.md).
> Colors: Deep Navy `#0A1628`, Electric Cyan `#00D4FF`, Soft Lavender `#A78BFA`.


## Overview

Continuum is a self-hosted cross-agent memory mesh. It provides structured memory persistence, real-time synchronization, and cost tracking across any AI agent that can speak HTTP or MCP.

## Design Principles

1. **Agent-agnostic** — Any agent that can make HTTP requests or use MCP can connect
2. **Self-hosted first** — No cloud dependency, no vendor lock-in
3. **Structured memory** — Not just text blobs; typed memory with relationships
4. **Real-time sync** — WebSocket for live updates across agents
5. **Cost-aware** — Track token burn per agent, per task, per session

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Agent Layer                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │  Hermes  │  │  Claude  │  │   Codex  │  │ OpenClaw │  │   Custom     │  │
│  │  Agent   │  │  Code    │  │   CLI    │  │   CLI    │  │   Agent      │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────┬───────┘  │
│       │             │             │             │               │          │
│       └─────────────┴──────┬──────┴─────────────┘               │          │
│                            │                                     │          │
│                     ┌──────▼──────┐                    ┌────────▼────────┐  │
│                     │  MCP Server │                    │  HTTP Client    │  │
│                     │  (stdio)    │                    │  (any language) │  │
│                     └──────┬──────┘                    └────────┬────────┘  │
│                            │                                    │           │
└────────────────────────────┼────────────────────────────────────┼───────────┘
                             │                                    │
                             └──────────────┬─────────────────────┘
                                            │
┌───────────────────────────────────────────▼─────────────────────────────────┐
│                         Continuum API Server (Go)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐ │
│  │   REST API   │  │  WebSocket   │  │    MCP       │  │   GraphQL        │ │
│  │   /api/v1/*  │  │   /ws        │  │   Server     │  │   /graphql       │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘ │
│         │                 │                  │                   │          │
│         └─────────────────┴────────┬─────────┴───────────────────┘          │
│                                    │                                        │
│                         ┌──────────▼──────────┐                            │
│                         │   Core Services     │                            │
│                         ├─────────────────────┤                            │
│                         │ • Memory Manager    │                            │
│                         │ • Search Engine     │                            │
│                         │ • Sync Coordinator  │                            │
│                         │ • Cost Tracker      │                            │
│                         │ • Graph Builder     │                            │
│                         └──────────┬──────────┘                            │
│                                    │                                        │
│         ┌──────────────────────────┼──────────────────────────┐            │
│         ▼                          ▼                          ▼            │
│  ┌──────────────┐        ┌──────────────┐        ┌──────────────┐         │
│  │   Episodic   │        │   Semantic   │        │  Procedural  │         │
│  │   Memory     │        │   Memory     │        │   Memory     │         │
│  │   Service    │        │   Service    │        │   Service    │         │
│  └──────┬───────┘        └──────┬───────┘        └──────┬───────┘         │
│         │                       │                       │                  │
└─────────┼───────────────────────┼───────────────────────┼──────────────────┘
          │                       │                       │
          └───────────────────────┼───────────────────────┘
                                  │
┌─────────────────────────────────▼───────────────────────────────────────────┐
│                              Storage Layer                                   │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────┐  │
│  │   PostgreSQL     │  │      Redis       │  │   Vector Store           │  │
│  │   (primary DB)   │  │   (cache/pubsub) │  │   (pgvector extension)   │  │
│  │                  │  │                  │  │                          │  │
│  │  • sessions      │  │  • hot memory    │  │  • semantic search       │  │
│  │  • memories      │  │  • pub/sub       │  │  • similarity queries    │  │
│  │  • entities      │  │  • rate limits   │  │  • embeddings            │  │
│  │  • relationships │  │  • sessions      │  │                          │  │
│  │  • cost_logs     │  │                  │  │                          │  │
│  └──────────────────┘  └──────────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Memory Write

```
Agent → API → Validation → Memory Type Router → Storage → Indexing → Broadcast
```

1. Agent sends memory via HTTP POST or MCP tool call
2. API validates input (schema, size limits, auth)
3. Router determines memory type (episodic/semantic/procedural)
4. Stored in PostgreSQL with transaction
5. Embeddings generated and stored in pgvector
6. Graph relationships updated
7. WebSocket broadcast to connected agents

### Memory Read

```
Agent → API → Cache Check → Search/Retrieve → Response
```

1. Agent requests memory via HTTP GET or MCP tool call
2. Check Redis cache first
3. If miss, query PostgreSQL + pgvector
4. Format response with context window budget
5. Return to agent

### Cross-Agent Sync

```
Agent A writes → Continuum → WebSocket broadcast → Agent B receives
```

1. Agent A writes memory
2. Continuum persists and indexes
3. WebSocket event emitted to all connected agents
4. Agent B receives event and can choose to incorporate

## Memory Types

### Episodic Memory
- **What:** Session transcripts, decisions, mistakes
- **Structure:** `{ session_id, timestamp, role, content, metadata }`
- **Use case:** "What did we decide about auth last Tuesday?"
- **Retention:** Auto-summarize after 7 days, archive after 30 days

### Semantic Memory
- **What:** Facts, entities, relationships
- **Structure:** `{ entity_type, name, attributes, relationships[], confidence }`
- **Use case:** "Who is Alex on this project? What's the tech stack?"
- **Retention:** Persistent, manual deletion only

### Procedural Memory
- **What:** Skills, workflows, patterns
- **Structure:** `{ name, steps, context, success_rate, last_used }`
- **Use case:** "How do we deploy to production?"
- **Retention:** Persistent, update on success/failure

## Technology Choices

| Component | Choice | Rationale |
|-----------|--------|-----------|
| API Server | Go 1.24+ | Fast, single binary, excellent concurrency |
| Database | PostgreSQL 16+ | Reliable, pgvector for embeddings |
| Cache | Redis 7+ | Pub/sub for real-time sync |
| Frontend | React + TypeScript + Vite | Modern, fast dev, type safety |
| Styling | Tailwind CSS | Utility-first, consistent |
| Charts | D3.js / Recharts | Knowledge graph visualization |
| Protocols | REST + WebSocket + MCP | Maximum compatibility |

## Scalability

- Horizontal scaling via read replicas for PostgreSQL
- Redis Cluster for cache/pubsub
- Stateless API servers behind load balancer
- Embedding generation can be offloaded to background workers

## Security

- API key authentication per agent
- Input validation at all boundaries
- SQL injection prevention via parameterized queries
- XSS prevention in frontend
- Rate limiting per agent
- Audit logs for all mutations
