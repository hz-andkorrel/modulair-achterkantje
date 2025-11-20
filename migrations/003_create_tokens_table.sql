-- Create tokens table for JWT storage
CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    subject TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_tokens_user FOREIGN KEY (subject) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tokens_subject ON tokens(subject);
CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_tokens_revoked ON tokens(revoked);

-- Add comments
COMMENT ON TABLE tokens IS 'Stores all issued JWT tokens for tracking and revocation';
COMMENT ON COLUMN tokens.token IS 'The full JWT token string';
COMMENT ON COLUMN tokens.subject IS 'The subject (user identifier) the token was issued for';
COMMENT ON COLUMN tokens.issued_at IS 'When the token was issued';
COMMENT ON COLUMN tokens.expires_at IS 'When the token expires';
COMMENT ON COLUMN tokens.revoked IS 'Whether the token has been manually revoked';
COMMENT ON COLUMN tokens.created_at IS 'Database record creation timestamp';
