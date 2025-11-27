#!/bin/bash

# Simple STK Push - maintains SSE connection properly
# Usage: ./send-stk.sh [amount] [phone]

SERVER_URL="http://localhost:8080"
PHONE="${2:-254723975141}"
AMOUNT="${1:-10}"

echo "💸 Sending STK Push"
echo "==================="
echo "Phone: $PHONE"
echo "Amount: $AMOUNT KES"
echo ""

# Check server
if ! curl -s "$SERVER_URL/health" --max-time 2 > /dev/null 2>&1; then
    echo "❌ Server not running"
    exit 1
fi

# Create a temporary file for SSE output
SSE_OUTPUT=$(mktemp)
SESSION_FILE=$(mktemp)

# Cleanup on exit
cleanup() {
    rm -f "$SSE_OUTPUT" "$SESSION_FILE"
    if [ ! -z "$SSE_PID" ]; then
        kill $SSE_PID 2>/dev/null
    fi
}
trap cleanup EXIT

# Start SSE connection in background and capture session ID
echo "📡 Connecting to server..."
curl -N -s "$SERVER_URL/sse" > "$SSE_OUTPUT" &
SSE_PID=$!

# Wait for session ID to appear
for i in {1..10}; do
    if grep -q "sessionid=" "$SSE_OUTPUT" 2>/dev/null; then
        SESSION_ID=$(grep -m 1 "sessionid=" "$SSE_OUTPUT" | sed -E 's/.*sessionid=([^[:space:]]+).*/\1/')
        break
    fi
    sleep 0.5
done

if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to get session ID"
    cat "$SSE_OUTPUT"
    exit 1
fi

echo "✅ Connected (Session: ${SESSION_ID:0:10}...)"
echo ""

# Give SSE connection a moment to stabilize
sleep 1

# Initialize
echo "🔧 Initializing..."
curl -s -X POST "$SERVER_URL/message?sessionid=$SESSION_ID" \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "send-stk", "version": "1.0.0"}
    }
  }' > /dev/null

sleep 1

# Send STK Push
echo "📱 Sending STK Push..."
RESPONSE=$(curl -s -X POST "$SERVER_URL/message?sessionid=$SESSION_ID" \
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
        \"account_reference\": \"Payment\",
        \"transaction_desc\": \"Test\"
      }
    }
  }")

echo ""
echo "📋 Response:"
echo "$RESPONSE"
echo ""

# Check for success
if echo "$RESPONSE" | grep -q "CheckoutRequestID"; then
    echo "✅ STK Push sent successfully!"
    echo "📱 Check your phone!"
    
    # Show ngrok info if available
    if curl -s http://localhost:4040/api/tunnels > /dev/null 2>&1; then
        NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | grep -o '"public_url":"https://[^"]*' | grep -o 'https://[^"]*' | head -1)
        echo ""
        echo "🔔 Callback URL: $NGROK_URL/callback"
        echo "📊 Monitor: http://localhost:4040"
    fi
elif echo "$RESPONSE" | grep -q "error"; then
    echo "❌ Error occurred"
else
    echo "⚠️  Check response above"
fi

echo ""
