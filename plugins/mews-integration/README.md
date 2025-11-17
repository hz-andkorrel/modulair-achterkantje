# Mews Integration Plugin

This plugin integrates with the Mews PMS API using **webhooks** to receive real-time guest check-in events and publish them to your Redis event bus.

## What It Does

1. **Listens for Mews webhooks** on HTTP endpoint `/webhook`
2. **Receives real-time events** when reservations are updated (2-5 min delay)
3. **Fetches reservation details** from Mews API when needed
4. **Publishes Redis events** for checked-in guests only
5. **Other plugins listen** to these events and react (unlock doors, send emails, create invoices, etc.)

## Architecture

```
Mews Webhook → HTTP Server (this plugin) → Redis Pub/Sub → Your Other Plugins
               (receives POST)              (publishes)      (listeners)
```

**Event Flow:**
1. Guest checks in at Mews
2. Mews sends webhook POST to `http://your-server:8080/webhook`
3. Plugin receives `ServiceOrderUpdated` event
4. Plugin fetches reservation details from Mews API
5. If state is "Started" (checked in), publishes to Redis
6. Key activation and other plugins react immediately

## Configuration

### 1. Get Mews API Credentials

1. Log into your Mews account
2. Go to Settings → Integrations → Mews Connector API
3. Create a new Access Token
4. Copy your `ClientToken` and `AccessToken`

### 2. Set Environment Variables

Create a `.env` file in this directory (see `.env.example`):

```bash
MEWS_CLIENT_TOKEN=your_client_token_here
MEWS_ACCESS_TOKEN=your_access_token_here
MEWS_PLATFORM_ADDRESS=https://api.mews-demo.com  # or your production URL
MEWS_CLIENT_NAME=HotelIntegration 1.0.0
REDIS_ADDR=hub_bus:6379
WEBHOOK_PORT=8080
```

### 3. Expose Webhook Endpoint Publicly

Mews needs to reach your webhook endpoint. Options:

**Option A: ngrok (for testing)**
```bash
ngrok http 8080
# Copy the https URL, e.g., https://abc123.ngrok.io
```

**Option B: Cloud server (production)**
- Deploy to Azure, AWS, DigitalOcean, etc.
- Use public IP or domain
- Example: `https://your-hotel.com:8080/webhook`

**Option C: Port forwarding (local network)**
- Forward port 8080 on your router to your server
- Use dynamic DNS service

### 4. Configure Webhook in Mews

1. Log into Mews Dashboard
2. Go to **Settings** → **Integrations** → **Webhooks**
3. Create new webhook subscription:
   - **URL**: `https://your-public-url:8080/webhook`
   - **Event**: `ServiceOrderUpdated` (for reservations)
   - **Enterprise**: Select your property
4. Save and test

### 5. Add to docker-compose.yml

Already added! The service exposes port 8080:

```yaml
  mews-integration:
    build: ./plugins/mews-integration
    container_name: mews_integration
    env_file:
      - ./plugins/mews-integration/.env
    ports:
      - "8080:8080"  # Webhook endpoint
    networks:
      - private_network
      - public_network
    depends_on:
      eventbus:
        condition: service_healthy
    restart: always
```

## Events Published

## How It Works

1. **HTTP Server**: Listens on port 8080 for incoming webhooks
2. **Webhook Receives POST**: `/webhook` endpoint receives Mews events
3. **Async Processing**: Processes `ServiceOrderUpdated` events in background
4. **Fetch Details**: Calls Mews API to get full reservation details by ID
5. **Filter State**: Only processes reservations with state "Started" (checked in)
6. **Publish to Redis**: Sends CheckInEvent to Redis event bus
7. **Quick Response**: Returns HTTP 200 OK to Mews immediately (prevents timeout)

## Webhook Flow

**1. Mews sends webhook:**
```json
{
  "EnterpriseId": "851df8c8-...",
  "IntegrationId": "c8bee838-...",
  "Events": [
    {
      "Discriminator": "ServiceOrderUpdated",
      "Value": { "Id": "reservation-uuid" }
    }
  ]
}
```

**2. Plugin publishes to Redis:**

### Channel: `hotel.events` and `hotel.events.guest.checked_in`

```json
{
  "source": "mews",
  "event_type": "guest.checked_in",
  "reservation_id": "abc123-...",
  "reservation_number": "R-2025-001",
  "room_id": "room-305-uuid",
  "guest_account_id": "guest-uuid",
  "check_in_time": "2025-11-17T14:00:00Z",
  "state": "Started",
  "timestamp": "2025-11-17T14:05:30Z"
}
```

## Example: Key Activation Plugin Listening

```go
// Listen for Mews check-in events
subscription := redis.Subscribe(ctx, "hotel.events.guest.checked_in")

for msg := range subscription.Channel() {
    var event CheckInEvent
    json.Unmarshal([]byte(msg.Payload), &event)
    
    if event.Source == "mews" && event.EventType == "guest.checked_in" {
        fmt.Printf("🔓 Unlocking room %s\n", event.RoomID)
        // Unlock door logic here
    }
}
```

## How It Works

1. **Polling Mechanism**: Checks Mews API every 30 seconds
2. **Incremental Updates**: Only fetches reservations updated since last check
3. **State Filter**: Only gets reservations with state "Started" (checked in)
4. **Redis Publishing**: Each check-in is published as a Redis event
5. **Decoupled**: Other plugins don't need to know about Mews

## Reservation States in Mews

- `Inquired` - Not confirmed
- `Confirmed` - Confirmed reservation
- **`Started`** - **Checked in** ← This plugin monitors this state
- `Processed` - Checked out
- `Canceled` - Canceled

## Testing

### 1. Start the plugin:
```bash
docker-compose up mews-integration
```

### 2. Check the logs:
```bash
docker-compose logs -f mews-integration
```

You should see:
```
🏨 Mews Integration Plugin Started
Polling Mews API for check-ins...
--------------------------------------------------
📊 Found 1 checked-in reservation(s)
✅ Published check-in event for reservation R-2025-001 (Room: room-305-uuid)
```

### 3. Listen to Redis events:
```bash
docker exec -it hub_bus redis-cli
SUBSCRIBE hotel.events.guest.checked_in
```

## Production Considerations

1. **Error Handling**: Plugin logs errors but continues running
2. **Rate Limiting**: Polls every 30 seconds (adjustable)
3. **Lookback Window**: Initially checks last 24 hours on startup
4. **Restart Safe**: Tracks last checked time to avoid duplicates
5. **Health**: Validates Redis connection on startup

## Extending

To monitor other reservation states:

```go
States: []string{"Started", "Processed"},  // Add "Processed" for check-outs
```

To change poll interval:

```go
pollInterval := 60 * time.Second  // Poll every 60 seconds instead
```

## Troubleshooting

**Problem**: No events being published

**Solution**:
1. Check Mews API credentials are correct
2. Verify network connectivity to Mews API
3. Check Redis is running: `docker-compose ps`
4. View logs: `docker-compose logs mews-integration`

**Problem**: Duplicate events

**Solution**: Restart the plugin - it will reset the lookback window

## Related Plugins

- **key-activation**: Listens for check-ins, unlocks doors
- **billing**: Listens for check-ins, creates invoices
- **email**: Listens for check-ins, sends welcome emails
