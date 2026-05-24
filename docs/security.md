# Continuum Security Model


> **Brand:** This document follows the [Continuum Circuit Board Brand System](branding/BRAND_SYSTEM.md).
> Colors: Deep Navy `#0A1628`, Electric Cyan `#00D4FF`, Soft Lavender `#A78BFA`.


> **Design Principle:** Security must be invisible to users. Zero friction. Zero interruptions. Zero "security theater" that gets in the way of agents doing their job.

## Philosophy

Continuum's security model follows three rules:

1. **Secure by default** — Everything is locked down out of the box
2. **Zero friction** — Security never blocks legitimate agent workflows
3. **Transparent** — Users can see what's protected and how, but don't need to manage it

---

## Threat Model

### Assets
1. **Memory Data** — User's project knowledge, decisions, code context
2. **API Keys** — Agent authentication tokens
3. **Cost Data** — Spending information, budget settings

### Threat Actors
| Actor | Risk | Our Approach |
|-------|------|-------------|
| External attacker | Steal memory data | Network isolation + API keys |
| Compromised agent | Use stolen key | Rate limits + audit logs |
| Network attacker | MITM | TLS 1.3 everywhere |

---

## Security Architecture: Invisible Defense

```
┌─────────────────────────────────────────────────────────────┐
│                      Agent Layer                             │
│  Hermes ──► Claude Code ──► Codex ──► OpenClaw             │
│     │           │            │           │                  │
│     └───────────┴────────────┴───────────┘                  │
│                      │                                       │
│         ┌────────────▼────────────┐                         │
│         │   Continuum API (Go)    │  ← Security is HERE    │
│         │  REST + WebSocket + MCP │    (invisible to user)  │
│         └────────────┬────────────┘                         │
│                      │                                       │
│  ┌───────────────────┼───────────────────┐                 │
│  ▼                   ▼                   ▼                 │
│ ┌────────┐     ┌──────────┐     ┌──────────┐              │
│ │Episodic│     │ Semantic │     │Procedural│              │
│ │Memory  │     │ Memory   │     │ Memory   │              │
│ └────────┘     └──────────┘     └──────────┘              │
└─────────────────────────────────────────────────────────────┘
```

**Key insight:** Agents connect the same way they always have. Security happens *inside* Continuum, not at the agent boundary.

---

## What Users DON'T Have to Do

| Security Task | Traditional Approach | Continuum Approach |
|--------------|---------------------|-------------------|
| Manage certificates | Manual TLS setup | Auto-generated, auto-renewed |
| Rotate API keys | Calendar reminders | Automatic, seamless |
| Configure firewalls | iptables rules | Docker network isolation |
| Audit access | Manual log review | Automated, alerts only on anomalies |
| Backup encryption | GPG keys management | Encrypted by default, no user action |
| Rate limiting | nginx config | Built-in, auto-tuned |

---

## Security Controls (Zero-Config)

### 1. Authentication: One API Key, Zero Hassle

```bash
# User does this ONCE during setup
docker compose up -d
# API key is auto-generated and printed to logs
# Agents use it via env var — same as any other service
```

```bash
# In agent config (same pattern as OpenAI, Anthropic, etc.)
export CONTINUUM_API_KEY="ctm_auto_generated_secure_key"
```

**What we handle:**
- Key generation with 256-bit entropy
- Secure storage (hashed, never plaintext in DB)
- Automatic rotation (old key works during grace period)
- Revocation (instant, no restart needed)

**What users DON'T do:**
- Generate keys manually
- Store keys in password managers
- Remember rotation schedules
- Distribute keys to agents

### 2. Network: TLS Without the Pain

```yaml
# docker-compose.yml — TLS is automatic
services:
  continuum:
    environment:
      # Auto-generates self-signed cert on first run
      # Or provide your own:
      - CONTINUUM_TLS_CERT=/certs/cert.pem
      - CONTINUUM_TLS_KEY=/certs/key.pem
```

**Default behavior:**
- Localhost: HTTP (no TLS needed for local dev)
- Network exposure: Auto-TLS via Let's Encrypt
- Custom certs: Drop files in `/certs`, done

### 3. Memory Isolation: Automatic

```go
// This happens automatically — agents don't change behavior
// Agent A writes:
continuum.memory_write(
    type="semantic",
    content="OAuth2 + PKCE decision",
    visibility="shared"  // ← optional, defaults to shared
)

// Agent B reads:
continuum.memory_search(query="OAuth2")
// ← Only sees shared memories, private ones invisible
```

**Rules (enforced automatically):**
- Private memories: Only visible to writing agent
- Shared memories: Visible to all agents (default)
- No configuration needed — agents use same API

### 4. Rate Limiting: Invisible Protection

```
Agent: "I want to write 10,000 memories"
Continuum: "Sure, but I'll smooth that out so the DB doesn't choke"

Agent: "I want to search every millisecond"
Continuum: "I'll cache that for you, no need to hammer the DB"
```

**Defaults:**
- 1,000 requests/minute per agent (generous for normal use)
- Burst allowance for batch operations
- WebSocket: 10 messages/second (plenty for real-time sync)

**What users see:** Nothing. Rate limits are high enough that legitimate workflows never hit them.

### 5. Content Safety: Sanitize Without Stripping

```python
# Agent writes this (legitimate code snippet):
memory = """
<script src="/auth.js"></script>
function login() { ... }
"""

# Continuum stores it safely:
# - HTML tags preserved in raw storage
# - Dashboard displays as code block (not rendered)
# - Search indexes the text content
```

**What we DON'T do:**
- ❌ Reject memories with `<script>` tags (legitimate code has these)
- ❌ Strip HTML from content (breaks code snippets)
- ❌ Block "dangerous" words (false positives)

**What we DO:**
- ✅ Dashboard renders all content as plain text / code blocks
- ✅ CSP headers prevent execution even if something slips through
- ✅ Output encoding ensures browser treats it as text

---

## Attack Scenarios: How We Handle Them

### Scenario 1: Stolen API Key

**What happens:**
1. Attacker uses key from compromised agent
2. Rate limiting kicks in (unusual pattern detected)
3. Audit log records everything
4. Admin gets alert (Discord/webhook)
5. Admin revokes key via dashboard (one click)
6. New key auto-generated, agents reconnect seamlessly

**User impact:** Zero. Legitimate agents keep working.

### Scenario 2: Malicious Agent Writes Nonsense

**What happens:**
1. Agent writes false memory: "We decided to use HTTP not HTTPS"
2. Memory stored with agent attribution (always visible)
3. Other agents see: "[Claude] suggested HTTP not HTTPS" 
4. Human reviews, flags as incorrect
5. Reputation score adjusts, future memories from this agent flagged for review

**User impact:** Memories are labeled by source. Users decide what to trust.

### Scenario 3: Network Attacker (MITM)

**What happens:**
1. Attacker intercepts local network traffic
2. TLS 1.3 prevents decryption
3. Certificate pinning (optional) detects tampering

**User impact:** Zero. Encryption is automatic.

### Scenario 4: DoS via Memory Bomb

**What happens:**
1. Agent tries to write 10,000 × 10MB memories
2. Per-agent quota prevents storage exhaustion
3. Compression reduces bloat automatically
4. Admin gets alert

**User impact:** Zero for legitimate users. Attacking agent is throttled, others unaffected.

---

## Security Configuration (Optional Overrides)

```yaml
# continuum.yaml — Everything below is OPTIONAL
# Defaults are secure, only override if you need to

security:
  # Only change if you have specific compliance needs
  api_keys:
    rotation_days: 90        # default: 90
    max_per_agent: 3         # default: 3
  
  # Only change if you have unusual workloads
  rate_limiting:
    requests_per_minute: 1000  # default: 1000
    burst_size: 100            # default: 100
  
  # Only change if you need stricter isolation
  memory:
    default_visibility: shared   # default: shared
    max_content_size: 10MB       # default: 10MB
  
  # Only change if you have custom TLS certs
  network:
    tls_auto: true           # default: auto (Let's Encrypt)
    # tls_cert: /path/to/cert.pem
    # tls_key: /path/to/key.pem
  
  # Only change if you need compliance logging
  audit:
    enabled: true            # default: true
    retention_days: 365      # default: 365
```

**Default rule:** If you don't specify it, it's secure.

---

## What Makes Continuum Different

| Aspect | Traditional Security | Continuum Security |
|--------|---------------------|-------------------|
| Setup time | Hours of configuration | Zero — works out of the box |
| Maintenance | Regular rotation, updates | Automatic, seamless |
| User friction | MFA prompts, CAPTCHAs | Invisible to agents |
| Flexibility | Rigid policies | Adaptive, learns normal patterns |
| Visibility | Security dashboards everywhere | Only alerts when needed |

---

## Security Checklist (For Us, Not Users)

Users don't see this. This is our internal quality bar:

- [ ] Agent can connect with zero config beyond API key
- [ ] Agent never sees a security error during normal use
- [ ] Agent never needs to retry due to rate limits (under normal load)
- [ ] Memory write succeeds on first attempt (no validation rejections for legitimate content)
- [ ] Dashboard shows security status but doesn't require action
- [ ] Alerts only fire for actual anomalies, not normal patterns
- [ ] Key rotation happens without agent restart
- [ ] TLS works without user providing certificates

---

## Incident Response (Automated)

| Severity | Trigger | Response | User Notification |
|----------|---------|----------|-----------------|
| **Critical** | Mass data exfiltration | Auto-block IP, revoke keys | Immediate Discord alert |
| **High** | Brute force detected | Rate limit, flag agent | Dashboard notification |
| **Medium** | Unusual access pattern | Log for review | Weekly digest only |
| **Low** | Single failed auth | Log silently | None |

---

## Reporting Security Issues

Security issues should be reported to pinches-tapioca0v@icloud.com.

Do NOT open public issues for security vulnerabilities.
