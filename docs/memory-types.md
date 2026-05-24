# 🧠 Memory Types


> **Brand:** This document follows the [Continuum Circuit Board Brand System](branding/BRAND_SYSTEM.md).
> Colors: Deep Navy `#0A1628`, Electric Cyan `#00D4FF`, Soft Lavender `#A78BFA`.


When to write to which store.

---

## The Rule of Thumb

| Question | Memory Type |
|----------|-------------|
| "What happened?" | **Episodic** |
| "What is true?" | **Semantic** |
| "How do I do this?" | **Procedural** |

---

## 📼 Episodic

**Stores:** Time-anchored events, session transcripts, tool calls, user corrections.

**Best for:** Resuming where you left off, audit trails, debugging.

**Write when:**
- A session starts or ends
- A tool is called with specific arguments
- The user corrects or overrides something
- An error occurs with context

**Example:**
```json
{
  "content": "User asked to refactor auth middleware. Claude Code applied the change. Tests passed.",
  "memory_type": "episodic",
  "tags": ["session", "refactor", "auth"],
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Query pattern:**
```bash
curl "http://localhost:8080/api/v1/memories/search?q=auth+refactor+last+week"
```

---

## 🧩 Semantic

**Stores:** Facts, entities, preferences, relationships, vector embeddings.

**Best for:** Answering project questions, building a knowledge base.

**Write when:**
- You learn a user preference
- You discover a system constraint
- You identify an entity (person, service, repo)
- A configuration value is confirmed

**Example:**
```json
{
  "content": "Production database runs PostgreSQL 16 on node-2. Owner: Platform team. Connection limit: 200.",
  "memory_type": "semantic",
  "tags": ["database", "production", "infrastructure"]
}
```

**Query pattern:**
```bash
curl "http://localhost:8080/api/v1/memories/search?q=database+connection+limit"
```

---

## 🛠️ Procedural

**Stores:** Workflows, skill recipes, command sequences, verified runbooks.

**Best for:** Repeatable tasks, onboarding new agents, standardising processes.

**Write when:**
- A multi-step task succeeds and should be reusable
- You discover an efficient command sequence
- A deployment process is validated
- A debugging playbook is proven

**Example:**
```json
{
  "content": "Deploy to staging: 1) git tag vX.Y.Z-staging 2) docker build -t app:staging . 3) kubectl rollout restart deployment/staging 4) wait for pods 5) run smoke tests",
  "memory_type": "procedural",
  "tags": ["deploy", "staging", "workflow"]
}
```

**Query pattern:**
```bash
curl "http://localhost:8080/api/v1/memories/search?q=how+to+deploy+staging"
```

---

## Visibility

Control who sees what.

| Level | Access |
|-------|--------|
| `shared` | All agents (default) |
| `private` | Only the writing agent |
| `team` | Agents with matching scope |

```json
{
  "content": "Internal API key for staging",
  "memory_type": "semantic",
  "visibility": "private",
  "tags": ["secret", "staging"]
}
```

---

## Tagging Strategy

Good tags make search work. Bad tags make noise.

| Good | Bad |
|------|-----|
| `["auth", "middleware", "refactor"]` | `["stuff", "thing", "update"]` |
| `["deploy", "production", "rollback"]` | `["done", "fix", "ok"]` |
| `["user-preference", "ui", "dark-mode"]` | `["misc", "general"]` |

Use 2–4 tags per memory. Be specific. Use kebab-case.
