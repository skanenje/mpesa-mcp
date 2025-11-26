#!/bin/bash

# M-Pesa MCP Test Client
# This script demonstrates how to properly test the MCP server

set -e

SERVER_URL="http://localhost:8080"
SESSION_ID=""

echo "🧪 M-Pesa MCP Test Client"
echo "========================="
echo ""

# Step 1: Connect to SSE endpoint to get session ID
echo "📡 Step 1: Connecting to SSE endpoint..."
echo "Running: curl -N $SERVER_URL/sse"
echo ""
echo "⚠️  Keep this terminal open and look for the session ID in the output"
echo "    It will look like: event: endpoint"
echo "                       data: /message?session=session_XXXXXXXXXXXXX"
echo ""
echo "Then, in another terminal, run one of these test commands:"
echo ""
echo "Test STK Push:"
echo "curl -X POST \"$SERVER_URL/message?session=SESSION_ID_HERE\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"jsonrpc\": \"2.0\","
echo "    \"id\": \"1\","
echo "    \"method\": \"tools/call\","
echo "    \"params\": {"
echo "      \"name\": \"stk_push\","
echo "      \"arguments\": {"
echo "        \"phone_number\": \"254723975141\","
echo "        \"amount\": 10"
echo "      }"
echo "    }"
echo "  }'"
echo ""
echo "Test Token Status:"
echo "curl -X POST \"$SERVER_URL/message?session=SESSION_ID_HERE\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"jsonrpc\": \"2.0\","
echo "    \"id\": \"2\","
echo "    \"method\": \"tools/call\","
echo "    \"params\": {"
echo "      \"name\": \"get_token_status\","
echo "      \"arguments\": {}"
echo "    }"
echo "  }'"
echo ""
echo "Starting SSE connection..."
echo "========================="
echo ""

# Connect to SSE
curl -N "$SERVER_URL/sse"
