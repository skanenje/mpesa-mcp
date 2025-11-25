#!/bin/bash

# Manual test requests for MCP server
# Usage: ./test-requests.sh <session_id>

if [ -z "$1" ]; then
    echo "Usage: $0 <session_id>"
    echo "Get session_id from the SSE endpoint event"
    exit 1
fi

SESSION_ID=$1
BASE_URL="http://localhost:8080"

echo "=== Testing MCP Requests with Session: $SESSION_ID ==="
echo ""

# Test 1: Initialize
echo "1. Sending initialize request..."
curl -X POST "$BASE_URL/message?session=$SESSION_ID" \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'
echo -e "\n"

sleep 1

# Test 2: List tools
echo "2. Sending tools/list request..."
curl -X POST "$BASE_URL/message?session=$SESSION_ID" \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
echo -e "\n"

sleep 1

# Test 3: Get token status
echo "3. Calling get_token_status tool..."
curl -X POST "$BASE_URL/message?session=$SESSION_ID" \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "get_token_status",
      "arguments": {}
    }
  }'
echo -e "\n"

echo ""
echo "Check the SSE connection terminal for responses!"
