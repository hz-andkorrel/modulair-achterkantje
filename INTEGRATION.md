# Integration Architecture

## Overview

The broker now acts as a **reverse proxy** for registered plugins, routing requests based on the `base-api-route` field.

```
┌────────────────────────────────────────────────────────────────┐
│                         CLIENT                                  │
│              (User Portal, Mobile App, etc.)                   │
└─────────────────────────┬──────────────────────────────────────┘
                          │
                          │ HTTP Request
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      BROKER SERVICE                              │
│                    (localhost:8081)                              │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ 1. Receive request (e.g., /api/v1/albums)              │   │
│  └────────────────────────────────────────────────────────┘   │
│                          ▼                                      │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ 2. Look up plugin with base-api-route="/api/v1"        │   │
│  │    (uses longest prefix match)                         │   │
│  └────────────────────────────────────────────────────────┘   │
│                          ▼                                      │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ 3. Forward to plugin.Host                              │   │
│  │    http://localhost:8080/api/v1/albums                 │   │
│  └────────────────────────────────────────────────────────┘   │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          │ Reverse Proxy
          ┌───────────────┼───────────────┐
          │               │               │
          ▼               ▼               ▼
┌─────────────────┐ ┌─────────────┐ ┌──────────────┐
│  InternalAPI    │ │  Kiosk      │ │ Other        │
│  :8080          │ │  Plugin     │ │ Plugins      │
│  /api/v1/*      │ │  :9000      │ │              │
│                 │ │  /kiosk/*   │ │              │
└─────────────────┘ └─────────────┘ └──────────────┘
```

## Request Flow

### 1. Plugin Registration (Startup)

```sequence
InternalAPI → Broker: POST /api/v1/route
                      {
                        "slug": "internal-api",
                        "host": "http://localhost:8080",
                        "base-api-route": "/api/v1",
                        "enabled": true
                      }
Broker → Broker: Store in registry
Broker → Broker: Save to plugins.json
Broker → InternalAPI: 201 Created
```

### 2. Request Proxying

```sequence
Client → Broker: GET /api/v1/albums
Broker → Broker: Find plugin with base-api-route="/api/v1"
Broker → InternalAPI: GET http://localhost:8080/api/v1/albums
InternalAPI → Broker: 200 OK { "albums": [...] }
Broker → Client: 200 OK { "albums": [...] }
```

### 3. Route Matching Priority

The broker uses **longest prefix matching**:

| Request Path          | Registered Plugins             | Match            |
|-----------------------|--------------------------------|------------------|
| `/api/v1/albums`     | `/api/v1`, `/api`             | `/api/v1` (longer) |
| `/kiosk/status`      | `/kiosk`, `/api/v1`           | `/kiosk`         |
| `/unknown/path`      | `/api/v1`, `/kiosk`           | 404 Not Found    |

## Plugin Model

```go
type Plugin struct {
    Slug          string   // Unique identifier: "internal-api"
    Name          string   // Display name: "Hotel Internal API"
    Version       string   // Semantic version: "2.0.0"
    Description   string   // Brief description
    Host          string   // Target URL: "http://localhost:8080"
    BaseAPIRoute  string   // Route prefix: "/api/v1"
    SettingsRoute string   // Admin settings path (optional)
    APIRoutes     []string // Available endpoints (optional)
    Enabled       bool     // Active status
}
```

## Key Features

### 1. **Reverse Proxy**
- Uses Go's `httputil.ReverseProxy`
- Preserves headers and request body
- Forwards full request path (no path rewriting by default)

### 2. **Route Conflict Detection**
- Prevents duplicate `base-api-route` registrations
- Blocks registration on reserved broker routes (`/api/v1/route`, `/api/v1/status`)

### 3. **Persistent Storage**
- Saves plugin registrations to `data/plugins.json`
- Automatically reloads on broker restart
- Atomic file writes prevent corruption

### 4. **JWT Authentication**
- Required for plugin registration/updates/deletion
- Optional for proxied requests (handled by target plugin)

## Integration Example

InternalAPI integration is implemented in `../InternalAPI/internal/broker/register.go`.

### InternalAPI Integration

```go
// Add to InternalAPI main.go
func main() {
    // Register with broker on startup
    RegisterWithBroker() // Non-blocking, logs on failure
    
    // Start InternalAPI normally
    router.Run(":8080")
}
```

### Environment Variables

**InternalAPI:**
```bash
export BROKER_URL="http://localhost:8081"
export BROKER_AUTH_TOKEN="eyJhbGc..."
export HOST="localhost"
export PORT="8080"
```

**Broker:**
```bash
export BROKER_PORT="8081"
export PLUGINS_PERSIST_PATH="data/plugins.json"
```

## Testing

```bash
# Start broker
cd broker
go run main.go

# In another terminal, register InternalAPI
curl -X POST http://localhost:8081/api/v1/route \
  -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "internal-api",
    "name": "Hotel Internal API",
    "host": "http://localhost:8080",
    "base-api-route": "/api/v1",
    "enabled": true
  }'

# Test proxying
curl http://localhost:8081/api/v1/albums
# → Forwarded to http://localhost:8080/api/v1/albums
```

## Future Enhancements

- [ ] Health checks for registered plugins
- [ ] Load balancing across multiple plugin instances
- [ ] Circuit breaker for failing plugins
- [ ] Request/response transformation
- [ ] Rate limiting per plugin
- [ ] Metrics and observability (request counts, latency)
