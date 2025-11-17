# Broker Service

A modular broker service with plugin registration and reverse proxy capabilities.

## Features

- **Plugin Registration API**: Services can register themselves to receive routed traffic
- **Reverse Proxy**: Automatically forwards requests to registered plugins based on route matching
- **JWT Authentication**: Secure plugin registration and optional endpoint authentication
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
      "host": "http://localhost:8080",
      "base-api-route": "/api/v1",
      "enabled": true
    }
  ]
}
```

#### `PUT /api/v1/route/:slug` - Update a plugin (requires JWT)

#### `DELETE /api/v1/route/:slug` - Delete a plugin (requires JWT)

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

- `BROKER_PORT`: Service port (default: `8081`)
- `JWT_EXPIRY`: Token validity duration (default: `10m`)
- `JWT_ISSUER`: JWT issuer name (default: `broker-service`)
- `PLUGINS_PERSIST_PATH`: Plugin storage file path (default: `data/plugins.json`)
- `APP_VERSION`: Application version (default: `1.0.0`)

## Getting Started

### Start the Broker

```bash
cd broker
go mod download
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
