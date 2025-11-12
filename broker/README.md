# Broker Service

Een eenvoudige broker service met een status endpoint die optioneel authenticatie ondersteunt.

## Features

- **Status Endpoint**: `GET /api/v1/status`
  - Retourneert service status, timestamp en versie
  - Als een geldig JWT token wordt meegegeven, toont het ook gebruikersinformatie
  - Zonder token werkt het endpoint ook gewoon

## Gebruik

### Status zonder authenticatie

```bash
curl http://localhost:8081/api/v1/status
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-11-12T10:44:00Z",
  "version": "1.0.0",
  "authenticated": false
}
```

### Status met authenticatie

Eerst een token verkrijgen van de auth-service:

```bash
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"email":"admin@example.com","password":"admin123"}'
$token = $loginResponse.access_token
```

Dan status opvragen met het token:

```bash
curl -H "Authorization: Bearer $token" http://localhost:8081/api/v1/status
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-11-12T10:44:00Z",
  "version": "1.0.0",
  "authenticated": true,
  "user": {
    "id": "user-id-here",
    "email": "admin@example.com"
  }
}
```

## Configuratie

Omgevingsvariabelen:

- `BROKER_PORT`: Poort waarop de service draait (default: 8081)
- `JWT_EXPIRY`: Hoe lang tokens geldig zijn (default: 10m)
- `JWT_ISSUER`: JWT issuer naam (default: broker-service)

## Starten

```bash
cd broker
go mod download
go run main.go
```
