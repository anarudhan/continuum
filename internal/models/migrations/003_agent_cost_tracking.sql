-- Migration: Add cost tracking fields to agents table
-- This allows agents to track cumulative token usage and cost

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'agents' AND column_name = 'tokens_used'
    ) THEN
        ALTER TABLE agents ADD COLUMN tokens_used INTEGER DEFAULT 0;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'agents' AND column_name = 'cost_usd'
    ) THEN
        ALTER TABLE agents ADD COLUMN cost_usd DECIMAL(10, 6) DEFAULT 0;
    END IF;
END $$;
