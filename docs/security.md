# Continuum Security Model

## Threat Model

### Assets
1. **Memory Data** — User's project knowledge, decisions, credentials mentioned in context
2. **API Keys** — Agent authentication tokens
3. **Cost Data** — Spending information, budget settings
4. **System State** — Running configuration, agent identities

### Threat Actors
1. **External attacker** — Unauthorized access to memory data
2. **Malicious agent** — Agent that writes poisoned/false memories
3. **Compromised agent** — Legitimate agent whose API key is stolen
4. **Insider** — User with dashboard access who shouldn't see all data
5. **Network attacker** — Man-in-the-middle on WebSocket/HTTP

---

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Perimeter                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   WAF/Ratel │  │   mTLS      │  │   API Key + JWT         │  │
│  │   Limiting  │  │   (opt)     │  │   Authentication        │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
│         └─────────────────┴─────────────────────┘                │
│                              │                                   │
├──────────────────────────────┼───────────────────────────────────┤
│                         API Layer                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Input     │  │   Output    │  │   CORS + CSRF           │  │
│  │   Validation│  │   Encoding  │  │   Protection            │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│                              │                                   │
├──────────────────────────────┼───────────────────────────────────┤
│                      Data Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   SQL       │  │   XSS       │  │   Audit Logging         │  │
│  │   Injection │  │   Sanitiz.  │  │   (immutable)           │  │
│  │   Prevention│  │             │  │                         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│                              │                                   │
├──────────────────────────────┼───────────────────────────────────┤
│                    Storage Layer                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Encrypted │  │   Network   │  │   Backup                │  │
│  │   Volumes   │  │   Isolation │  │   Encryption            │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Security Controls (Implemented)

### Authentication & Authorization

| Control | Implementation | Status |
|---------|---------------|--------|
| API Key per agent | UUID v4, 256-bit entropy | P0 |
| Key rotation | 90-day TTL with grace period | P1 |
| Scope restriction | Read-only vs read-write keys | P2 |
| Admin dashboard auth | Separate JWT with 2FA option | P2 |
| mTLS for MCP | Optional client certificates | P2 |

### Input Validation

| Control | Implementation | Status |
|---------|---------------|--------|
| Schema validation | JSON Schema for all endpoints | P0 |
| Size limits | 10MB max memory content | P0 |
| Rate limiting | Per-agent, per-IP, global | P0 |
| Content sanitization | Strip HTML/JS from memory content | P0 |
| SQL injection prevention | Parameterized queries (pgx) | P0 |
| NoSQL injection prevention | Redis command validation | P0 |

### Output Protection

| Control | Implementation | Status |
|---------|---------------|--------|
| XSS prevention | Content-Security-Policy headers | P0 |
| Output encoding | HTML escape in dashboard | P0 |
| JSON serialization | Strict struct marshaling | P0 |
| Error sanitization | No stack traces to clients | P0 |

### Network Security

| Control | Implementation | Status |
|---------|---------------|--------|
| TLS 1.3 | All external communication | P0 |
| HSTS headers | Enforce HTTPS | P0 |
| CORS policy | Strict origin whitelist | P0 |
| WebSocket origin validation | Reject cross-origin WS | P0 |
| Request timeout | 30s max for all endpoints | P0 |

### Data Protection

| Control | Implementation | Status |
|---------|---------------|--------|
| Encryption at rest | PostgreSQL TDE + volume encryption | P1 |
| Encryption in transit | TLS 1.3 for all connections | P0 |
| Memory content encryption | AES-256-GCM for sensitive memories | P2 |
| Backup encryption | GPG-encrypted backups | P2 |
| Secure deletion | Overwrite before volume release | P2 |

### Audit & Monitoring

| Control | Implementation | Status |
|---------|---------------|--------|
| Immutable audit log | Append-only, tamper-evident | P0 |
| Failed auth logging | Track brute force attempts | P0 |
| Anomaly detection | Unusual memory access patterns | P2 |
| Admin alerts | Security events to Discord/webhook | P1 |

---

## Agent Isolation Model

### Memory Visibility

```
┌─────────────────────────────────────────┐
│           Memory Access Matrix           │
├─────────────┬──────────┬────────────────┤
│   Memory    │  Owner   │  Other Agents  │
├─────────────┼──────────┼────────────────┤
│ Private     │   Full   │   None         │
│ Shared      │   Full   │   Read         │
│ Broadcast   │   Full   │   Read+Write   │
└─────────────┴──────────┴────────────────┘
```

### Trust Levels

| Level | Description | Capabilities |
|-------|-------------|--------------|
| **Untrusted** | New/unverified agent | Read shared, write private only |
| **Trusted** | Verified agent (manual approval) | Read shared, write shared |
| **Admin** | Dashboard user / owner | Full access, config changes |

---

## Attack Scenarios & Mitigations

### Scenario 1: Stolen API Key

**Attack:** Attacker obtains agent's API key and reads all memories.

**Mitigations:**
- Keys are scoped (read-only vs read-write)
- Rate limiting detects abnormal usage
- Audit log tracks all access
- Key rotation forces re-authentication
- IP allowlisting (optional)

### Scenario 2: Poisoned Memory Injection

**Attack:** Malicious agent writes false memories to mislead other agents.

**Mitigations:**
- Memory attribution (always tagged with agent ID)
- Confidence scoring (low confidence for new/untrusted agents)
- Conflict detection (flag contradictions)
- Manual review queue for high-impact memories
- Agent reputation system (track accuracy over time)

### Scenario 3: Memory Content XSS

**Attack:** Agent writes memory containing `<script>alert('xss')</script>`.

**Mitigations:**
- Content sanitization on write (strip HTML/JS)
- CSP headers prevent inline script execution
- Output encoding in dashboard (HTML escape)
- Memory content treated as plain text, never rendered as HTML

### Scenario 4: SQL Injection via Search

**Attack:** Agent sends `q='; DROP TABLE memories; --` in search.

**Mitigations:**
- All queries use pgx parameterized statements
- Full-text search uses PostgreSQL `to_tsquery` with escaping
- No raw SQL concatenation anywhere
- Read-only search endpoint (separate from write)

### Scenario 5: DoS via Memory Bomb

**Attack:** Agent writes 10,000 x 10MB memories to exhaust storage.

**Mitigations:**
- Per-agent memory quota (configurable)
- Rate limiting on writes
- Size limits per memory (10MB max)
- Total storage alerts
- Automatic compression reduces bloat

### Scenario 6: WebSocket Hijacking

**Attack:** Attacker opens WebSocket with stolen API key to eavesdrop.

**Mitigations:**
- WebSocket origin validation (same-origin policy)
- API key required in connection handshake
- Per-connection rate limiting
- Heartbeat/ping to detect stale connections
- Connection limit per agent

### Scenario 7: Privilege Escalation

**Attack:** Agent exploits bug to access other agents' private memories.

**Mitigations:**
- Row-level security in PostgreSQL
- Every query includes `agent_id = ?` filter
- Service layer enforces access control, not just API layer
- Unit tests for authorization boundaries

### Scenario 8: Supply Chain Attack

**Attack:** Compromised dependency introduces backdoor.

**Mitigations:**
- Go modules with checksum verification
- Minimal dependencies (only essential packages)
- Docker image scanning (Trivy/Snyk)
- No `latest` tags in Dockerfiles
- SBOM generation for compliance

---

## Security Checklist (Pre-Release)

### Code Level
- [ ] No hardcoded secrets (use env vars)
- [ ] All inputs validated (schema + size + type)
- [ ] All outputs encoded (no raw HTML/JS)
- [ ] SQL injection tests pass
- [ ] XSS tests pass
- [ ] CSRF protection enabled
- [ ] Rate limiting configured
- [ ] Error messages don't leak internals

### Infrastructure
- [ ] TLS 1.3 enforced
- [ ] HSTS headers set
- [ ] Security headers (CSP, X-Frame-Options, etc.)
- [ ] Docker non-root user
- [ ] Read-only filesystem where possible
- [ ] No unnecessary ports exposed
- [ ] Secrets in Docker secrets/env, not images

### Data
- [ ] Encryption at rest enabled
- [ ] Backup encryption enabled
- [ ] Audit log immutable
- [ ] Data retention policy documented
- [ ] Secure deletion implemented

### Operations
- [ ] Security event alerting
- [ ] Incident response plan
- [ ] Dependency vulnerability scanning
- [ ] Penetration testing (before v1.0)
- [ ] Security documentation published

---

## Secure Configuration Defaults

```yaml
# continuum.yaml — Secure defaults
security:
  api_keys:
    min_entropy: 256
    rotation_days: 90
    max_per_agent: 3
  
  rate_limiting:
    requests_per_minute: 1000
    burst_size: 100
    websocket_messages_per_second: 10
  
  memory:
    max_content_size: 10MB
    max_per_agent: 10000
    sanitize_content: true
    encrypt_sensitive: true
  
  network:
    tls_min_version: "1.3"
    hsts_max_age: 31536000
    cors_origins: []  # Empty = same-origin only
    websocket_origins: []
  
  audit:
    enabled: true
    retention_days: 365
    immutable: true
  
  dashboard:
    auth_required: true
    session_timeout: 3600
    mfa_optional: true
```

---

## Reporting Security Issues

Security issues should be reported to security@ezerops.com.

Do NOT open public issues for security vulnerabilities.
