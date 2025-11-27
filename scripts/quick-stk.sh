#!/bin/bash

# Quick STK Push - Automated session ID + STK push to your number
# Usage: ./quick-stk.sh [amount] [phone]

set -e

SERVER_URL="http://localhost:8080"
PHONE="${2:-254723975141}"  # Default to your number
AMOUNT="${1:-10}"           # Default to 10 KES

echo "💸 Quick STK Push"
echo "================"
echo ""

# Check if server is running
echo "🔍 Checking server..."
MAX_RETRIES=5
RETRY=0

while [ $RETRY -lt $MAX_RETRIES ]; do
    if curl -s "$SERVER_URL/health" --max-time 2 > /dev/null 2>&1; then
        echo "✅ Server is running"
        break
    fi
    RETRY=$((RETRY + 1))
    if [ $RETRY -lt $MAX_RETRIES ]; then
        echo "   Waiting for server... ($RETRY/$MAX_RETRIES)"
        sleep 1
    fi
done

if [ $RETRY -eq $MAX_RETRIES ]; then
    echo "❌ Server not responding on port 8080"
    echo "   Start it with: go run cmd/server/main.go"
    exit 1
fi
echo ""

# Step 1: Get session ID
echo "📡 Getting session ID..."
SESSION_ID=$(timeout 3 curl -N -s "$SERVER_URL/sse" | grep -m 1 "sessionid=" | sed -E 's/.*sessionid=([^[:space:]]+).*/\1/')

if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to get session ID"
    exit 1
fi

echo "✅ Session ID: $SESSION_ID"
echo ""

# Step 2: Initialize MCP
echo "🔧 Initializing MCP..."
INIT_RESPONSE=$(curl -s -X POST "$SERVER_URL/message?sessionid=$SESSION_ID" \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "quick-stk", "version": "1.0.0"}
    }
  }')

if echo "$INIT_RESPONSE" | grep -q "error"; then
    echo "❌ Initialization failed:"
    echo "$INIT_RESPONSE" | jq '.'
    exit 1
fi

echo "✅ MCP initialized"
echo ""

# Step 3: Send STK Push
echo "📱 Sending STK Push..."
echo "   Phone: $PHONE"
echo "   Amount: $AMOUNT KES"
echo ""

STK_RESPONSE=$(curl -s -X POST "$SERVER_URL/message?sessionid=$SESSION_ID" \
  -H 'Content-Type: application/json' \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": \"2\",
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"stk_push\",
      \"arguments\": {
        \"phone_number\": \"$PHONE\",
        \"amount\": $AMOUNT,
        \"account_reference\": \"QuickPay\",
        \"transaction_desc\": \"Test Payment\"
      }
    }
  }")

# Show raw response first
echo "📋 Raw Response:"
echo "$STK_RESPONSE"
echo ""

# Check response
if echo "$STK_RESPONSE" | grep -q "error"; then
    echo "❌ STK Push failed:"
    echo "$STK_RESPONSE" | jq '.' 2>/dev/null || echo "$STK_RESPONSE"
    exit 1
fi

# Extract and display result
if echo "$STK_RESPONSE" | grep -q "CheckoutRequestID"; then
    echo "✅ STK Push sent successfully!"
    echo ""
    echo "📱 Check your phone for the M-Pesa prompt!"
    echo ""
    
    # If ngrok is running, show callback info
    if curl -s http://localhost:4040/api/tunnels > /dev/null 2>&1; then
        NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | grep -o '"public_url":"https://[^"]*' | grep -o 'https://[^"]*' | head -1)
        echo "🔔 Callback will be sent to: $NGROK_URL/callback"
        echo "📊 Monitor at: http://localhost:4040"
    else
        echo "⚠️  Ngrok not running - callbacks won't work"
        echo "   Start ngrok: ngrok http 8080"
    fi
elif echo "$STK_RESPONSE" | grep -qi "success\|initiated"; then
    echo "✅ STK Push sent!"
    echo "📱 Check your phone for the M-Pesa prompt!"
else
    echo "⚠️  Check the response above for details"
fi

echo ""
