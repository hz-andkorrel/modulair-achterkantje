#!/bin/bash

# Test webhook with a real reservation ID from Mews demo API
# This simulates Mews sending a ServiceOrderUpdated webhook

echo "🧪 Testing Mews Webhook Integration..."
echo ""

# Send webhook for a real reservation from Mews demo
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "EnterpriseId": "851df8c8-90f2-4c4a-8e01-a4fc46b25178",
    "IntegrationId": "c8bee838-7fb1-4f4e-8fac-ac87008b2f90",
    "Events": [
      {
        "Discriminator": "ServiceOrderUpdated",
        "Value": {
          "Id": "54285"
        }
      },
      {
        "Discriminator": "ServiceOrderUpdated",
        "Value": {
          "Id": "54598"
        }
      }
    ]
  }'

echo ""
echo ""
echo "✅ Webhook sent! Check logs:"
echo "   docker-compose logs -f mews-integration"
echo ""
echo "📊 Check Redis events:"
echo "   docker exec -it hub_bus redis-cli"
echo "   SUBSCRIBE hotel.events"
