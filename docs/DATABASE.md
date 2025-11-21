# Database Integration

## Overview

The broker service supports PostgreSQL for persistent token storage, enabling:

- **Token Revocation**: Invalidate tokens before they expire
- **Audit Trail**: Track all issued tokens and their usage
- **Session Management**: View and manage user sessions
- **Cross-Instance Validation**: Share token state across multiple broker instances
- **Security**: Detect and prevent token reuse after revocation

## Database Schema

### Tokens Table

```sql
CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    subject TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
- `idx_tokens_subject` - Fast lookup by user
- `idx_tokens_expires_at` - Efficient cleanup of expired tokens
- `idx_tokens_revoked` - Quick revocation checks

## Configuration

### Connection String Format

```
postgres://username:password@host:port/database?sslmode=mode
```

**Parameters:**
- `username`: PostgreSQL user
- `password`: User password
- `host`: Database server hostname
- `port`: PostgreSQL port (default: 5432)
- `database`: Database name
- `sslmode`: SSL mode (`disable`, `require`, `verify-ca`, `verify-full`)

**Examples:**

```bash
# Local development (no SSL)
DATABASE_URL=postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable

# Production (with SSL)
DATABASE_URL=postgres://broker:secure_pass@db.example.com:5432/broker_db?sslmode=require

# Using connection pooling
DATABASE_URL=postgres://broker:pass@localhost:5432/broker_db?pool_max_conns=25
```

## Setup Instructions

### 1. Using Docker Compose (Recommended)

The easiest way to get started:

```bash
# Start PostgreSQL and broker
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f broker

# Stop everything
docker-compose down

# Remove data volumes
docker-compose down -v
```

### 2. Manual PostgreSQL Setup

**Install PostgreSQL:**
```bash
# macOS
brew install postgresql@16

# Ubuntu/Debian
sudo apt-get install postgresql-16

# Windows (using Chocolatey)
choco install postgresql16
```

**Create Database:**
```bash
# Connect to PostgreSQL
psql -U postgres

# Create user and database
CREATE USER broker WITH PASSWORD 'broker123';
CREATE DATABASE broker_db OWNER broker;
GRANT ALL PRIVILEGES ON DATABASE broker_db TO broker;
\q
```

**Run Migrations:**
```bash
# Apply schema
psql -U broker -d broker_db -f migrations/001_create_tokens_table.sql

# Add cleanup function
psql -U broker -d broker_db -f migrations/002_cleanup_function.sql
```

### 3. Automatic Schema Initialization

The broker automatically creates tables on startup if `DATABASE_URL` is set:

```bash
export DATABASE_URL="postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable"
cd broker
go run main.go
```

Output:
```
Broker Configuration loaded:
  Database: enabled
Database connected and initialized
```

## API Operations

### Token Lifecycle

**1. Token Generation:**
```go
token, expiresAt, err := jwtManager.GenerateToken("user@example.com")
// Token is automatically saved to database
```

**2. Token Validation:**
```go
claims, err := jwtManager.ParseAndVerify(token)
// Checks database for revocation status
```

**3. Token Revocation:**
```go
err := db.RevokeToken(ctx, token)
// Marks token as revoked in database
```

### Database Methods

**Save Token:**
```go
err := db.SaveToken(ctx, token, subject, issuedAt, expiresAt)
```

**Check Revocation:**
```go
revoked, err := db.IsTokenRevoked(ctx, token)
```

**Get User's Tokens:**
```go
tokens, err := db.GetTokensBySubject(ctx, "user@example.com")
```

**Revoke All User Tokens:**
```go
err := db.RevokeAllTokensForSubject(ctx, "user@example.com")
```

**Cleanup Expired Tokens:**
```go
deletedCount, err := db.CleanupExpiredTokens(ctx)
```

## Maintenance

### Automatic Cleanup

Run periodic cleanup of expired tokens:

```go
// In a background goroutine
ticker := time.NewTicker(1 * time.Hour)
for range ticker.C {
    deleted, err := db.CleanupExpiredTokens(context.Background())
    if err != nil {
        log.Printf("Cleanup failed: %v", err)
    } else {
        log.Printf("Cleaned up %d expired tokens", deleted)
    }
}
```

### Manual Cleanup

Using PostgreSQL function:
```sql
SELECT cleanup_expired_tokens();
```

Using broker API (if implemented):
```bash
curl -X POST http://localhost:8081/api/v1/admin/cleanup-tokens \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Database Monitoring

**Check token count:**
```sql
SELECT COUNT(*) FROM tokens;
```

**View active tokens:**
```sql
SELECT subject, COUNT(*) as token_count
FROM tokens
WHERE expires_at > NOW() AND revoked = FALSE
GROUP BY subject;
```

**Find expired tokens:**
```sql
SELECT COUNT(*) 
FROM tokens 
WHERE expires_at < NOW();
```

**Check revoked tokens:**
```sql
SELECT subject, token, revoked, expires_at
FROM tokens
WHERE revoked = TRUE
ORDER BY created_at DESC
LIMIT 10;
```

## Performance Considerations

### Connection Pooling

The broker uses connection pooling:
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

### Indexes

All critical queries are indexed:
- Subject lookups: O(log n)
- Expiration checks: O(log n)
- Revocation checks: O(1) with token uniqueness

### Query Optimization

**Fast token validation:**
```go
// Single query checks both existence and revocation
revoked, err := db.IsTokenRevoked(ctx, token)
```

**Batch operations:**
```go
// Revoke all tokens for a user in one query
err := db.RevokeAllTokensForSubject(ctx, subject)
```

## Security Best Practices

### 1. Secure Connection Strings

Never commit credentials:
```bash
# Use environment variables
export DATABASE_URL="postgres://..."

# Or use secret management
export DATABASE_URL=$(aws secretsmanager get-secret-value --secret-id broker-db-url)
```

### 2. SSL/TLS in Production

Always use SSL in production:
```bash
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
```

### 3. Least Privilege

Grant only necessary permissions:
```sql
-- Create read-only user for monitoring
CREATE USER broker_readonly WITH PASSWORD 'readonly_pass';
GRANT CONNECT ON DATABASE broker_db TO broker_readonly;
GRANT SELECT ON tokens TO broker_readonly;
```

### 4. Regular Cleanup

Schedule automatic cleanup:
```bash
# Using cron
0 3 * * * psql $DATABASE_URL -c "SELECT cleanup_expired_tokens();"
```

## Troubleshooting

### Connection Issues

**Error: `connection refused`**
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection string
echo $DATABASE_URL
```

**Error: `authentication failed`**
```bash
# Verify credentials
psql $DATABASE_URL -c "SELECT version();"
```

### Schema Issues

**Error: `relation "tokens" does not exist`**
```bash
# Run migrations manually
psql $DATABASE_URL -f migrations/001_create_tokens_table.sql
```

### Performance Issues

**Slow token lookups:**
```sql
-- Verify indexes exist
\d tokens

-- Rebuild indexes if needed
REINDEX TABLE tokens;
```

## Migration Path

### From No Database to PostgreSQL

1. **Start with database disabled** (no DATABASE_URL)
2. **Set up PostgreSQL** using docker-compose
3. **Configure DATABASE_URL** and restart broker
4. **Verify**: New tokens are saved to database
5. **Optional**: No migration needed - old tokens simply expire

### Database Backup

**Full backup:**
```bash
pg_dump $DATABASE_URL > broker_backup.sql
```

**Tokens only:**
```bash
pg_dump $DATABASE_URL -t tokens > tokens_backup.sql
```

**Restore:**
```bash
psql $DATABASE_URL < broker_backup.sql
```

## Example Usage

### Complete Setup Script

```bash
#!/bin/bash

# Start PostgreSQL
docker run -d \
  --name broker-postgres \
  -e POSTGRES_USER=broker \
  -e POSTGRES_PASSWORD=broker123 \
  -e POSTGRES_DB=broker_db \
  -p 5432:5432 \
  postgres:16-alpine

# Wait for PostgreSQL to be ready
until docker exec broker-postgres pg_isready -U broker; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

# Set environment
export DATABASE_URL="postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable"

# Start broker (auto-creates tables)
cd broker
go run main.go
```

### Testing Token Persistence

```bash
# 1. Generate a token
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 2. Check database
psql $DATABASE_URL -c "SELECT subject, expires_at, revoked FROM tokens;"

# 3. Revoke token (if revocation endpoint exists)
curl -X POST http://localhost:8081/api/v1/auth/revoke \
  -H "Authorization: Bearer TOKEN"

# 4. Verify revocation
psql $DATABASE_URL -c "SELECT subject, revoked FROM tokens WHERE token='TOKEN';"
```
