-- Database initialization script for Marketplace Application
-- This script sets up the database with necessary extensions and initial configuration

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pg_stat_statements for query performance monitoring
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create custom types
DO $$ BEGIN
    CREATE TYPE token_status AS ENUM ('active', 'expired', 'refresh_needed', 'invalid');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE integration_type AS ENUM ('quickbooks', 'xero', 'sage', 'netsuite', 'other');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create indexes for better performance (these will be created by GORM as well, but ensuring they exist)
-- Note: GORM will handle table creation, this is just for additional setup

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create a function to generate connection IDs
CREATE OR REPLACE FUNCTION generate_connection_id(company_name TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN LOWER(REGEXP_REPLACE(company_name, '[^a-zA-Z0-9]', '_', 'g')) || '_' || EXTRACT(EPOCH FROM NOW())::TEXT;
END;
$$ LANGUAGE plpgsql;

-- Create a function to check token expiry
CREATE OR REPLACE FUNCTION is_token_expired(expires_at TIMESTAMP WITH TIME ZONE)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN expires_at IS NOT NULL AND expires_at <= NOW();
END;
$$ LANGUAGE plpgsql;

-- Create a function to get tokens expiring soon (within 1 hour)
CREATE OR REPLACE FUNCTION get_tokens_expiring_soon()
RETURNS TABLE(
    company_id UUID,
    business_name TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    minutes_until_expiry INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.id,
        c.business_name,
        c.token_expires_at,
        EXTRACT(EPOCH FROM (c.token_expires_at - NOW()))/60 AS minutes_until_expiry
    FROM companies c
    WHERE c.token_expires_at IS NOT NULL
    AND c.token_expires_at > NOW()
    AND c.token_expires_at <= NOW() + INTERVAL '1 hour'
    ORDER BY c.token_expires_at ASC;
END;
$$ LANGUAGE plpgsql;

-- Create a function to clean up old token refresh records
CREATE OR REPLACE FUNCTION cleanup_old_token_refreshes(days_old INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM token_refreshes 
    WHERE created_at < NOW() - (days_old || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a view for company statistics
CREATE OR REPLACE VIEW company_stats AS
SELECT 
    c.id,
    c.business_name,
    c.integration_id,
    c.created_at,
    COUNT(DISTINCT l.id) as location_count,
    COUNT(DISTINCT co.id) as contact_count,
    COUNT(DISTINCT p.id) as product_count,
    CASE 
        WHEN c.token_expires_at IS NULL THEN 'no_expiry'
        WHEN c.token_expires_at <= NOW() THEN 'expired'
        WHEN c.token_expires_at <= NOW() + INTERVAL '1 day' THEN 'expiring_soon'
        ELSE 'active'
    END as token_status,
    c.token_expires_at,
    c.last_sync_at
FROM companies c
LEFT JOIN locations l ON c.id = l.company_id AND l.deleted_at IS NULL
LEFT JOIN contacts co ON c.id = co.company_id AND co.deleted_at IS NULL
LEFT JOIN products p ON c.id = p.company_id AND p.deleted_at IS NULL
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.business_name, c.integration_id, c.created_at, c.token_expires_at, c.last_sync_at;

-- Create a view for token refresh statistics
CREATE OR REPLACE VIEW token_refresh_stats AS
SELECT 
    DATE_TRUNC('day', created_at) as refresh_date,
    COUNT(*) as total_refreshes,
    COUNT(CASE WHEN success = true THEN 1 END) as successful_refreshes,
    COUNT(CASE WHEN success = false THEN 1 END) as failed_refreshes,
    ROUND(COUNT(CASE WHEN success = true THEN 1 END) * 100.0 / COUNT(*), 2) as success_rate
FROM token_refreshes
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY refresh_date DESC;

-- Insert some initial configuration data if needed
-- This could include default settings, admin users, etc.

-- Create a settings table for application configuration
CREATE TABLE IF NOT EXISTS app_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default settings
INSERT INTO app_settings (key, value, description) VALUES
('token_refresh_interval', '3600', 'Token refresh interval in seconds')
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_settings (key, value, description) VALUES
('token_expiry_warning_hours', '24', 'Hours before expiry to send warnings')
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_settings (key, value, description) VALUES
('max_retry_attempts', '3', 'Maximum retry attempts for failed operations')
ON CONFLICT (key) DO NOTHING;

INSERT INTO app_settings (key, value, description) VALUES
('cache_ttl_seconds', '300', 'Default cache TTL in seconds')
ON CONFLICT (key) DO NOTHING;

-- Create trigger for updating updated_at on app_settings
DROP TRIGGER IF EXISTS update_app_settings_updated_at ON app_settings;
CREATE TRIGGER update_app_settings_updated_at
    BEFORE UPDATE ON app_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Grant necessary permissions
-- Note: In production, you should create specific users with limited permissions
GRANT USAGE ON SCHEMA public TO PUBLIC;
GRANT CREATE ON SCHEMA public TO PUBLIC;

-- Performance optimization settings
-- These are suggestions and should be tuned based on your specific workload
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET max_connections = 100;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- Log configuration for monitoring
ALTER SYSTEM SET log_statement = 'mod';
ALTER SYSTEM SET log_duration = on;
ALTER SYSTEM SET log_min_duration_statement = 1000; -- Log queries taking more than 1 second

-- Note: After running this script, you may need to restart PostgreSQL for some settings to take effect
-- SELECT pg_reload_conf(); -- This reloads configuration without restart for most settings

SELECT 'Database initialization completed successfully' as status;