# Continuum Development Guidelines

Behavioral guidelines for building Continuum, derived from Andrej Karpathy's observations on LLM coding pitfalls, adapted for systems programming and distributed systems.

**Tradeoff:** These guidelines bias toward correctness and simplicity over speed. For trivial tasks (typos, comments), use judgment.

---

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

**Distributed systems specific:**
- Always consider failure modes: network partitions, crashes, race conditions.
- State consistency guarantees explicitly (eventual, strong, causal).
- If a design adds coordination, justify why it's necessary.

---

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

**The test:** Would a senior engineer say this is overcomplicated? If yes, simplify.

**Go specific:**
- Prefer explicit over clever. Go is verbose by design — embrace it.
- Avoid generics unless they genuinely reduce duplication.
- Use structs with methods, not deep inheritance hierarchies.
- One package, one responsibility.

---

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it — don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

**The test:** Every changed line should trace directly to the user's request.

---

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

**Strong success criteria let you loop independently.** Weak criteria ("make it work") require constant clarification.

---

## 5. Data Integrity Is Non-Negotiable

**Memory is the product. Corruption is catastrophic.**

- Every write path must be tested with failure injection.
- Use transactions for multi-table operations.
- Validate all inputs at the API boundary.
- Never trust agent-provided data without sanitization.
- Maintain audit logs for all mutations.

---

## 6. Observability By Design

**If you can't observe it, you can't debug it.**

- Every operation gets a trace ID.
- Log at appropriate levels: ERROR for failures, WARN for anomalies, INFO for state changes, DEBUG for internals.
- Metrics for everything: latency, throughput, error rates, memory pressure.
- Health endpoints must be meaningful, not just "200 OK".

---

## 7. API Contract Discipline

**The API is a promise. Breaking it breaks agents.**

- Version all API endpoints from day one (`/api/v1/`).
- Never remove fields, only deprecate with sunset headers.
- Document every endpoint with request/response examples.
- Return consistent error shapes: `{ "error": "code", "message": "human", "details": {} }`.
- WebSocket events must have a schema and version.
