#!/bin/bash

# Test STK Push with fixed credentials

echo "🧪 Testing STK Push with corrected passkey..."
echo ""

# Start server in background
echo "Starting server..."
./mpesa-mcp > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test STK Push
echo "Sending STK Push request..."
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "stk_push",
      "arguments": {
        "amount": 1,
        "phone_number": "254708374149",
        "account_reference": "TestPayment",
        "transaction_desc": "Test"
      }
    }
  }' | jq .

echo ""
echo "Check server.log for detailed output"

# Kill server
kill $SERVER_PID 2>/dev/null
