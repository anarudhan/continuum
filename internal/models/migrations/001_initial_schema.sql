-- Continuum Initial Schema
-- PostgreSQL 16+ with pgvector extension

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "vector";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";      -- trigram for fuzzy search
CREATE EXTENSION IF NOT EXISTS "pgcrypto";     -- hashing

-- Agents table: registered AI agents
CREATE TABLE IF NOT EXISTS agents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    api_key_hash    VARCHAR(64) NOT NULL,       -- SHA-256 hash
    api_key_hint    VARCHAR(8),                  -- last 4 chars for identification
    trust_level     VARCHAR(20) DEFAULT 'trusted' CHECK (trust_level IN ('untrusted', 'trusted', 'admin')),
    scopes          TEXT[] DEFAULT ARRAY['read', 'write'],
    is_active       BOOLEAN DEFAULT true,
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agents_api_key_hash ON agents(api_key_hash);
CREATE INDEX IF NOT EXISTS idx_agents_active ON agents(is_active) WHERE is_active = true;

-- Sessions table: agent work sessions
CREATE TABLE IF NOT EXISTS sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    project         VARCHAR(255),
    task            VARCHAR(500),
    status          VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned')),
    summary         TEXT,
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    token_count     INTEGER DEFAULT 0,
    cost_usd        DECIMAL(10, 6) DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_agent ON sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project);

-- Memories table: all memory types (episodic, semantic, procedural)
CREATE TABLE IF NOT EXISTS memories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type            VARCHAR(20) NOT NULL CHECK (type IN ('episodic', 'semantic', 'procedural')),
    content         TEXT NOT NULL,
    content_vector  VECTOR(384),                 -- for semantic search (384-dim, lightweight)
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    session_id      UUID REFERENCES sessions(id) ON DELETE SET NULL,
    project         VARCHAR(255),
    visibility      VARCHAR(20) DEFAULT 'shared' CHECK (visibility IN ('private', 'shared', 'broadcast')),
    importance      DECIMAL(3, 2) DEFAULT 0.5,   -- 0.0 to 1.0
    confidence      DECIMAL(3, 2) DEFAULT 1.0,   -- for semantic memories
    metadata        JSONB DEFAULT '{}',
    parent_id       UUID REFERENCES memories(id) ON DELETE SET NULL,  -- for versioning
    is_archived     BOOLEAN DEFAULT false,
    archived_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_memories_agent ON memories(agent_id);
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
CREATE INDEX IF NOT EXISTS idx_memories_visibility ON memories(visibility);
CREATE INDEX IF NOT EXISTS idx_memories_project ON memories(project);
CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memories_session ON memories(session_id);
CREATE INDEX IF NOT EXISTS idx_memories_archived ON memories(is_archived) WHERE is_archived = false;

-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_memories_fts ON memories USING GIN (to_tsvector('english', content));

-- Semantic search index (ivfflat for approximate nearest neighbor)
CREATE INDEX IF NOT EXISTS idx_memories_vector ON memories USING ivfflat (content_vector vector_cosine_ops)
    WITH (lists = 100);

-- Trigram index for fuzzy text search
CREATE INDEX IF NOT EXISTS idx_memories_trgm ON memories USING GIN (content gin_trgm_ops);

-- Entities table: extracted named entities from semantic memories
CREATE TABLE IF NOT EXISTS entities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    entity_type     VARCHAR(100) NOT NULL,       -- 'project', 'person', 'technology', etc.
    attributes      JSONB DEFAULT '{}',
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    confidence      DECIMAL(3, 2) DEFAULT 1.0,
    first_seen_at   TIMESTAMPTZ DEFAULT NOW(),
    last_seen_at    TIMESTAMPTZ DEFAULT NOW(),
    mention_count   INTEGER DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(name, entity_type, agent_id)
);

CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name);
CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_entities_agent ON entities(agent_id);

-- Entity-Memory junction table
CREATE TABLE IF NOT EXISTS memory_entities (
    memory_id       UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
    entity_id       UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    confidence      DECIMAL(3, 2) DEFAULT 1.0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (memory_id, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_mem_entities_memory ON memory_entities(memory_id);
CREATE INDEX IF NOT EXISTS idx_mem_entities_entity ON memory_entities(entity_id);

-- Relationships table: graph edges between entities
CREATE TABLE IF NOT EXISTS relationships (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id       UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_id       UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    relation_type   VARCHAR(100) NOT NULL,       -- 'uses', 'depends_on', 'created_by', etc.
    attributes      JSONB DEFAULT '{}',
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    confidence      DECIMAL(3, 2) DEFAULT 1.0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(source_id, target_id, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_relations_source ON relationships(source_id);
CREATE INDEX IF NOT EXISTS idx_relations_target ON relationships(target_id);
CREATE INDEX IF NOT EXISTS idx_relations_type ON relationships(relation_type);

-- Cost logs table: token usage tracking
CREATE TABLE IF NOT EXISTS cost_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    session_id      UUID REFERENCES sessions(id) ON DELETE SET NULL,
    provider        VARCHAR(50) NOT NULL,        -- 'openai', 'anthropic', etc.
    model           VARCHAR(100) NOT NULL,       -- 'gpt-4o', 'claude-sonnet-4'
    tokens_input    INTEGER DEFAULT 0,
    tokens_output   INTEGER DEFAULT 0,
    tokens_total    INTEGER DEFAULT 0,
    cost_usd        DECIMAL(10, 6) DEFAULT 0,
    request_type    VARCHAR(50),                 -- 'memory_write', 'memory_read', etc.
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cost_logs_agent ON cost_logs(agent_id);
CREATE INDEX IF NOT EXISTS idx_cost_logs_session ON cost_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_cost_logs_created ON cost_logs(created_at DESC);

-- Audit log table: immutable security events
CREATE TABLE IF NOT EXISTS audit_log (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp       TIMESTAMPTZ DEFAULT NOW(),
    event_type      VARCHAR(50) NOT NULL,
    agent_id        UUID REFERENCES agents(id) ON DELETE SET NULL,
    agent_name      VARCHAR(255),
    action          VARCHAR(100) NOT NULL,
    resource        VARCHAR(255),
    details         JSONB DEFAULT '{}',
    ip_address      INET,
    success         BOOLEAN DEFAULT true,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_agent ON audit_log(agent_id);
CREATE INDEX IF NOT EXISTS idx_audit_type ON audit_log(event_type);

-- Budgets table: per-agent spending limits
CREATE TABLE IF NOT EXISTS budgets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    daily_limit     DECIMAL(10, 2) DEFAULT 10.00,
    weekly_limit    DECIMAL(10, 2) DEFAULT 50.00,
    monthly_limit   DECIMAL(10, 2) DEFAULT 100.00,
    alert_threshold DECIMAL(3, 2) DEFAULT 0.80,  -- alert at 80%
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_id)
);

-- Provider pricing table
CREATE TABLE IF NOT EXISTS provider_pricing (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider        VARCHAR(50) NOT NULL,
    model           VARCHAR(100) NOT NULL,
    price_per_1k_input  DECIMAL(10, 6) NOT NULL,
    price_per_1k_output DECIMAL(10, 6) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    effective_from  TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider, model, effective_from)
);

-- Insert default pricing
INSERT INTO provider_pricing (provider, model, price_per_1k_input, price_per_1k_output) VALUES
    ('openai', 'gpt-4o', 0.005000, 0.015000),
    ('openai', 'gpt-4o-mini', 0.000150, 0.000600),
    ('anthropic', 'claude-sonnet-4', 0.003000, 0.015000),
    ('anthropic', 'claude-opus-4', 0.015000, 0.075000),
    ('google', 'gemini-1.5-pro', 0.003500, 0.010500),
    ('xai', 'grok-2', 0.002000, 0.010000)
ON CONFLICT DO NOTHING;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE OR REPLACE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_memories_updated_at BEFORE UPDATE ON memories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_entities_updated_at BEFORE UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_relationships_updated_at BEFORE UPDATE ON relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_budgets_updated_at BEFORE UPDATE ON budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
