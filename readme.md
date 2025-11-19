# Broker Service

A modular broker service with plugin registration and reverse proxy capabilities.

## Features

- **Plugin Registration API**: Services can register themselves to receive routed traffic
- **Reverse Proxy**: Automatically forwards requests to registered plugins based on route matching
- **JWT Authentication**: Secure plugin registration and optional endpoint authentication
- **Token Revocation**: Revoke individual or all user tokens with database persistence
- **Automatic Cleanup**: Background job removes expired tokens every hour
- **Graceful Shutdown**: Properly closes database connections and completes in-flight requests
- **Request Logging**: Structured logging of all requests with status, duration, and client IP
- **Status Endpoint**: Health check with optional user information
- **Persistent Storage**: Plugin registrations are saved to disk and reloaded on startup

## Architecture

The broker acts as a central gateway that:
1. **Accepts plugin registrations** via `POST /api/v1/route`
2. **Routes incoming requests** to the appropriate plugin based on `base-api-route`
3. **Forwards requests** using a reverse proxy to the plugin's `host` URL

```
┌──────────────┐
│ User/Client  │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│  Broker Service      │
│  (localhost:8081)    │
│                      │
│  - Plugin Registry   │
│  - Reverse Proxy     │
│  - JWT Auth          │
└──────┬───────────────┘
       │
       ├─────────────────┐
       ▼                 ▼
┌─────────────┐   ┌─────────────┐
│ InternalAPI │   │  Other      │
│ :8080       │   │  Plugins    │
│ /api/v1/*   │   │  /plugin/*  │
└─────────────┘   └─────────────┘
```

## API Endpoints

### Plugin Management (requires JWT authentication)

#### `POST /api/v1/route` - Register a plugin

Plugins register themselves with the broker to receive traffic.

**Request Body:**
```json
{
  "slug": "internal-api",
  "name": "Hotel Internal API",
  "version": "2.0.0",
  "description": "Gateway for user portal and admin services",
  "category": "gateway",
  "host": "http://localhost:8080",
  "base-api-route": "/api/v1",
  "settings-route": "/admin/system/stats",
  "api-routes": ["/api/v1/albums", "/api/auth/login", "/health"],
  "enabled": true
}
```

**Response:**
```json
{
  "message": "plugin registered",
  "plugin": { ... }
}
```

#### `GET /api/v1/routes` - List all registered plugins

**Response:**
```json
{
  "plugins": [
    {
      "slug": "internal-api",
      "name": "Hotel Internal API",
      "category": "gateway",
      "host": "http://localhost:8080",
      "base-api-route": "/api/v1",
      "enabled": true
    }
  ]
}
```

#### `GET /api/v1/routes/categories` - List all plugin categories

**Response:**
```json
{
  "categories": [
    {
      "category": "gateway",
      "count": 1
    },
    {
      "category": "user-interface",
      "count": 2
    }
  ],
  "total": 2
}
```

#### `GET /api/v1/routes/category/:category` - List plugins by category

**Response:**
```json
{
  "category": "user-interface",
  "count": 2,
  "plugins": [
    {
      "slug": "kiosk",
      "name": "Kiosk Plug-in",
      "category": "user-interface",
      "host": "http://localhost:8080",
      "base-api-route": "/kiosk",
      "enabled": true
    }
  ]
}
```

#### `PUT /api/v1/route/:slug` - Update a plugin (requires JWT)

#### `DELETE /api/v1/route/:slug` - Delete a plugin (requires JWT)

### Authentication Endpoints

#### `POST /api/v1/auth/revoke` - Revoke current token (requires JWT)

Revokes the token used in the request.

**Response:**
```json
{
  "message": "token revoked successfully"
}
```

#### `POST /api/v1/auth/revoke-all` - Revoke all user tokens (requires JWT)

Revokes all tokens for the current user.

**Response:**
```json
{
  "message": "all tokens revoked successfully"
}
```

### Admin Endpoints

#### `POST /api/v1/admin/cleanup-tokens` - Cleanup expired tokens (requires JWT)

Manually triggers cleanup of expired tokens from the database.

**Response:**
```json
{
  "message": "cleanup completed",
  "deleted": 42
}
```

### Status Endpoint

#### `GET /api/v1/status` - Service status

**Without authentication:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-17T10:44:00Z",
  "version": "1.0.0",
  "authenticated": false
}
```

**With JWT token:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-17T10:44:00Z",
  "version": "1.0.0",
  "authenticated": true,
  "user": {
    "id": "user-id-here",
    "email": "admin@example.com"
  }
}
```

### Reverse Proxy

All requests that don't match broker-specific routes are forwarded to registered plugins:

```bash
# Request to broker
curl http://localhost:8081/api/v1/albums

# Broker finds plugin with base-api-route="/api/v1"
# Forwards to: http://localhost:8080/api/v1/albums
```

## Integration with InternalAPI

See `examples/register-with-broker.go` for a complete example of how plugins should register with the broker.

### Environment Variables for InternalAPI

Add these to your InternalAPI service:

- `BROKER_URL`: URL of the broker service (default: `http://localhost:8081`)
- `BROKER_AUTH_TOKEN`: JWT token for authenticating with the broker

### Registration on Startup

Plugins should register themselves when they start:

```go
func RegisterWithBroker() error {
    registration := PluginRegistration{
        Slug:         "internal-api",
        Name:         "Hotel Internal API",
        Host:         "http://localhost:8080",
        BaseAPIRoute: "/api/v1",
        Enabled:      true,
    }
    
    // POST to broker's /api/v1/route endpoint
    // See examples/register-with-broker.go for full code
}
```

## Configuration

### Environment Variables

**Server Settings:**
- `BROKER_PORT`: Service port (default: `8081`)
- `APP_VERSION`: Application version (default: `1.0.0`)

**JWT Settings:**
- `JWT_EXPIRY`: Token validity duration (default: `10m`)
- `JWT_ISSUER`: JWT issuer name (default: `broker-service`)

**Plugin Persistence:**
- `PLUGINS_PERSIST_PATH`: Plugin storage file path (default: `data/plugins.json`)

**Database Settings:**
- `DATABASE_URL`: PostgreSQL connection string (optional)
  - Format: `postgres://username:password@host:port/database?sslmode=disable`
  - Leave empty to disable token persistence to database
  - Example: `postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable`

### Database Setup

The broker can persist JWT tokens to PostgreSQL for:
- Token revocation
- Audit trail
- Cross-instance token validation
- User session management

**1. Start PostgreSQL (using Docker):**
```bash
docker-compose up -d postgres
```

**2. Configure DATABASE_URL:**
```bash
export DATABASE_URL="postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable"
```

**3. Start the broker:**
The broker will automatically create the required tables on startup.

**Database Schema:**
- `tokens` table: Stores all issued JWT tokens
  - `id`: Primary key
  - `token`: Full JWT token string
  - `subject`: User identifier
  - `issued_at`: Token issue timestamp
  - `expires_at`: Token expiration timestamp
  - `revoked`: Revocation status
  - `created_at`: Record creation timestamp

**Manual Migration:**
If you prefer to run migrations manually:
```bash
psql $DATABASE_URL -f migrations/001_create_tokens_table.sql
psql $DATABASE_URL -f migrations/002_cleanup_function.sql
```

## Getting Started

### Option 1: Local Development (without database)

```bash
cd broker
go mod download
go run main.go
```

### Option 2: With PostgreSQL Database

**Using Docker Compose:**
```bash
# Start both PostgreSQL and broker
docker-compose up -d

# View logs
docker-compose logs -f broker

# Stop services
docker-compose down
```

**Manual Setup:**
```bash
# 1. Start PostgreSQL
docker run -d \
  --name broker-postgres \
  -e POSTGRES_USER=broker \
  -e POSTGRES_PASSWORD=broker123 \
  -e POSTGRES_DB=broker_db \
  -p 5432:5432 \
  postgres:16-alpine

# 2. Configure environment
export DATABASE_URL="postgres://broker:broker123@localhost:5432/broker_db?sslmode=disable"

# 3. Start broker
cd broker
go run main.go
```

### Start InternalAPI and Register

```bash
# Terminal 1: Start InternalAPI
cd ../InternalAPI
export BROKER_URL="http://localhost:8081"
export BROKER_AUTH_TOKEN="your-jwt-token"
go run main.go

# Terminal 2: Register InternalAPI with broker
go run ../modulair-achterkantje/examples/register-with-broker.go
```

### Test the Integration

```bash
# Request via broker - gets proxied to InternalAPI
curl http://localhost:8081/api/v1/albums \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Direct request to InternalAPI (bypasses broker)
curl http://localhost:8080/api/v1/albums \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Development

### Run Tests

```bash
cd broker
go test ./...
```

### Build

```bash
cd broker
go build -o broker.exe .
```

### Graceful Shutdown

The broker handles shutdown signals gracefully:

```bash
# Press Ctrl+C to trigger graceful shutdown
# Or send SIGTERM signal:
# kill -SIGTERM <pid>
```

**Shutdown process:**
1. Stops accepting new connections
2. Completes in-flight requests (up to 30 seconds)
3. Stops background cleanup job
4. Closes database connections
5. Exits cleanly

**Docker:**
```bash
docker-compose down  # Sends SIGTERM, triggers graceful shutdown
```

## Logging

The broker logs all requests with detailed information:

**Request Log Format:**
```
[REQUEST] 200 |     1.234ms |      127.0.0.1 | GET     /api/v1/status
[REQUEST] 201 |    12.456ms |      127.0.0.1 | POST    /api/v1/route
[PROXY] Forwarding to plugin 'internal-api' at http://localhost:8080
[REQUEST] 200 |   123.789ms |      127.0.0.1 | GET     /api/v1/albums
```

**Log includes:**
- HTTP status code
- Request duration
- Client IP address
- HTTP method and path
- Proxy forwarding details
- Error messages (if any)

**Example output:**
```
2025-11-19T10:30:15 Broker service starting on :8081
2025-11-19T10:30:20 [REQUEST] 200 |      2.145ms |   192.168.1.100 | GET     /api/v1/status
2025-11-19T10:30:25 [PROXY] Forwarding to plugin 'kiosk' at http://localhost:9000
2025-11-19T10:30:25 [REQUEST] 200 |     45.678ms |   192.168.1.100 | GET     /kiosk/welcome
2025-11-19T10:30:30 [PROXY ERROR] Plugin 'internal-api' failed: dial tcp: connection refused
2025-11-19T10:30:30 [REQUEST] 502 |     10.234ms |   192.168.1.100 | GET     /api/v1/albums
```
