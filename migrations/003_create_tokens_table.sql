-- Create tokens table for JWT storage
CREATE TABLE IF NOT EXISTS tokens (
    token TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    type TEXT NOT NULL DEFAULT 'access',
    CONSTRAINT fk_tokens_user FOREIGN KEY (subject) REFERENCES users(email) ON DELETE CASCADE
);
