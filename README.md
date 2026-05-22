# Continuum

> **The persistent brain for your agent swarm.**

Cross-agent memory mesh with structured memory types, real-time sync, cost guardrails, and a beautiful knowledge graph. Self-hosted. Open source. Agent-agnostic.

```bash
docker compose up -d
# Every agent on your network now shares a brain
```

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://golang.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7+-3178C6.svg)](https://typescriptlang.org)

## What Problem Does This Solve?

You run multiple AI agents: Hermes for frontend, Claude Code for features, Codex for PRs, OpenClaw for backend. Each session starts from zero. You re-explain your project. You re-teach preferences. You burn tokens and patience.

**Continuum fixes session amnesia across your entire agent swarm.**

## Quick Start

```bash
# Clone and start
git clone https://github.com/ezerops/continuum.git
cd continuum
docker compose up -d

# Dashboard at http://localhost:8080
# API at http://localhost:8080/api/v1
# WebSocket at ws://localhost:8080/ws
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Your Agents                             │
│  Hermes ◄────► Claude Code ◄────► Codex ◄────► OpenClaw    │
│     │              │                │            │           │
│     └──────────────┴────────────────┴────────────┘           │
│                        │                                     │
│           ┌────────────▼─────────────┐                      │
│           │   Continuum API (Go)     │                      │
│           │  REST + WebSocket + MCP  │                      │
│           └────────────┬─────────────┘                      │
│                        │                                     │
│     ┌──────────────────┼──────────────────┐                │
│     ▼                  ▼                  ▼                │
│ ┌─────────┐      ┌──────────┐      ┌──────────┐           │
│ │Episodic │      │ Semantic │      │Procedural│           │
│ │ Memory  │      │ Memory   │      │ Memory   │           │
│ │(session │      │(entities,│      │(skills,  │           │
│ │ history)│      │ facts)   │      │workflows)│           │
│ └────┬────┘      └────┬─────┘      └────┬─────┘           │
│      │                │                 │                  │
│      └────────────────┼─────────────────┘                  │
│                       ▼                                     │
│           ┌─────────────────────┐                          │
│           │  PostgreSQL +       │                          │
│           │  pgvector + Redis   │                          │
│           └─────────────────────┘                          │
└─────────────────────────────────────────────────────────────┘
```

## Documentation

- [Architecture](docs/architecture.md)
- [API Reference](docs/api.md)
- [Agent Integration](docs/agent-integration.md)
- [Memory Types](docs/memory-types.md)
- [Self-Hosting](docs/self-hosting.md)

## License

MIT © Ezerops
