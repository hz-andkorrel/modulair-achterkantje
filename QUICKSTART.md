# Quick Start Guide

## JWT Token Persistence with PostgreSQL

### Setup (3 steps)

**1. Start PostgreSQL using Docker Compose:**
```bash
docker-compose up -d postgres
```

**2. Set environment variable:**
```bash
# Windows PowerShell
$env:DATABASE_URL="postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable"

# Linux/macOS
export DATABASE_URL="postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable"
```

**3. Start the broker:**
```bash
cd broker
go run main.go
```

You should see:
```
Database connected and initialized
Broker service starting on :8081
```

### Testing Token Storage

**1. Generate a token (requires auth endpoint - see InternalAPI):**
```bash
# Example: Login to get a token
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

**2. Check the database:**
```bash
# Connect to PostgreSQL
docker exec -it broker-postgres psql -U broker -d broker_db

# View stored tokens
SELECT subject, issued_at, expires_at, revoked FROM tokens;

# Count tokens by user
SELECT subject, COUNT(*) FROM tokens GROUP BY subject;

# Exit
\q
```

**3. Token validation checks revocation:**
When you use a token, the broker automatically checks if it's been revoked in the database.

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env`:
```env
DATABASE_URL=postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable
BROKER_PORT=8081
JWT_EXPIRY=10m
```

Load environment (PowerShell):
```powershell
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^#][^=]+)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
```

### Docker Compose (Everything Together)

Start both PostgreSQL and the broker:
```bash
docker-compose up -d
```

Check logs:
```bash
docker-compose logs -f broker
```

Stop everything:
```bash
docker-compose down
```

### Disable Database

To run without database persistence:
```bash
# Remove or comment out DATABASE_URL
# $env:DATABASE_URL=""

cd broker
go run main.go
```

Output:
```
Database disabled - tokens will not be persisted
```

### Production Configuration

For production, use environment variables:
```bash
# Use strong credentials
DATABASE_URL=postgres://prod_user:strong_password@db.example.com:5432/broker_prod?sslmode=require

# Enable SSL
# ?sslmode=require or ?sslmode=verify-full
```

### Database Operations

**Cleanup expired tokens:**
```sql
docker exec -it broker-postgres psql -U broker -d broker_db \
  -c "SELECT cleanup_expired_tokens();"
```

**View all tokens:**
```sql
docker exec -it broker-postgres psql -U broker -d broker_db \
  -c "SELECT * FROM tokens ORDER BY created_at DESC LIMIT 10;"
```

**Revoke a specific token (manual):**
```sql
docker exec -it broker-postgres psql -U broker -d broker_db \
  -c "UPDATE tokens SET revoked = TRUE WHERE token = 'YOUR_TOKEN_HERE';"
```

### Troubleshooting

**Connection refused:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connectivity
docker exec broker-postgres pg_isready -U broker
```

**Authentication failed:**
```bash
# Verify credentials in docker-compose.yml match DATABASE_URL
cat docker-compose.yml | grep POSTGRES
echo $DATABASE_URL
```

**Tables not created:**
```bash
# Manual migration
docker exec -i broker-postgres psql -U broker -d broker_db < migrations/001_create_tokens_table.sql
```

For more details, see [DATABASE.md](DATABASE.md)
