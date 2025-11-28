-- Create tokens table for JWT storage
CREATE TABLE IF NOT EXISTS tokens (
    token TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_tokens_user FOREIGN KEY (subject) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tokens_subject ON tokens(subject);
CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_tokens_revoked ON tokens(revoked);
