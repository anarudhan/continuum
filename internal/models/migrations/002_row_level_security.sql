-- Row-Level Security for multi-agent isolation
-- Applied automatically on migration

-- Enable RLS on sensitive tables
ALTER TABLE memories ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE TABLE ROW LEVEL SECURITY;
ALTER TABLE relationships ENABLE ROW LEVEL SECURITY;
ALTER TABLE cost_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;

-- Create policy: agents can only see their own private memories
CREATE POLICY memory_owner_private ON memories
    FOR ALL
    USING (
        visibility != 'private'
        OR agent_id = current_setting('app.current_agent_id', true)::UUID
    );

-- Create policy: shared memories readable by all active agents
CREATE POLICY memory_shared_read ON memories
    FOR SELECT
    USING (
        visibility IN ('shared', 'broadcast')
        AND is_archived = false
    );

-- Sessions: agents can only see their own sessions
CREATE POLICY session_owner ON sessions
    FOR ALL
    USING (agent_id = current_setting('app.current_agent_id', true)::UUID);

-- Cost logs: agents can only see their own costs
CREATE POLICY cost_owner ON cost_logs
    FOR ALL
    USING (agent_id = current_setting('app.current_agent_id', true)::UUID);

-- Budgets: agents can only see their own budgets
CREATE POLICY budget_owner ON budgets
    FOR ALL
    USING (agent_id = current_setting('app.current_agent_id', true)::UUID);

-- Function to set agent context for RLS
CREATE OR REPLACE FUNCTION set_agent_context(agent_id UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_agent_id', agent_id::text, false);
END;
$$ LANGUAGE plpgsql;
