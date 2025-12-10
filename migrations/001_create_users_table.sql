-- Create users table for user management and authentication
CREATE TABLE IF NOT EXISTS users (
    email TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    password_hash TEXT,
    role TEXT DEFAULT 'user',
    enabled BOOLEAN DEFAULT TRUE
);

-- Create default admin user (password: 'admin123' - CHANGE THIS IN PRODUCTION!)
-- Password hash generated with: bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
INSERT INTO users (email, name, password_hash, role, enabled)
VALUES (
    'admin@example.com',
    'System Administrator',
    '$2a$10$RXUz3Sq2kiZ2NKZciHVnFuqFyPHgf1G6OIfiMJIeAJETmIi8yeBsu',
    'admin',
    true
)
ON CONFLICT (email) DO NOTHING;
