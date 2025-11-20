-- Function to automatically cleanup expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM tokens WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled job to run cleanup (requires pg_cron extension)
-- If pg_cron is not available, run this manually or via external cron:
-- SELECT cleanup_expired_tokens();

COMMENT ON FUNCTION cleanup_expired_tokens IS 'Removes all expired tokens from the database';
