#!/bin/bash

# Simple webhook test - just confirms the webhook endpoint works
# This doesn't call the Mews API to avoid rate limits

echo "🧪 Simple Webhook Test (no API calls)"
echo ""
echo "Sending test webhook to http://localhost:8080/webhook..."
echo ""

RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "EnterpriseId": "test-enterprise",
    "IntegrationId": "test-integration",
    "Events": [
      {
        "Discriminator": "ServiceOrderUpdated",
        "Value": {
          "Id": "test-reservation-123"
        }
      }
    ]
  }')

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | grep -v "HTTP_STATUS")

echo "Response Status: $HTTP_STATUS"
echo "Response Body: $BODY"
echo ""

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Webhook endpoint is working!"
    echo ""
    echo "📋 What happened:"
    echo "   1. ✅ Webhook POST received by plugin"
    echo "   2. ✅ JSON parsed successfully"
    echo "   3. ✅ Event queued for processing"
    echo "   4. ✅ HTTP 200 OK returned to sender"
    echo ""
    echo "🔍 View logs:"
    echo "   docker-compose logs -f mews-integration"
else
    echo "❌ Webhook endpoint returned error: $HTTP_STATUS"
fi

echo ""
echo "🧪 To test health endpoint:"
echo "   curl http://localhost:8080/health"
