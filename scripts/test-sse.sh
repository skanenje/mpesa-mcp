#!/bin/bash

# Test script for SSE MCP server

echo "=== Testing M-Pesa MCP Server with SSE ==="
echo ""

BASE_URL="http://localhost:8080"

# Test 1: Health check
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq .
echo ""

# Test 2: Connect to SSE (run in background)
echo "2. Connecting to SSE endpoint..."
echo "   (This will show the endpoint URL and keep connection open)"
echo ""
curl -N "$BASE_URL/sse" &
SSE_PID=$!

# Wait for connection to establish
sleep 2

# Extract session ID from SSE output (you'll need to manually get this)
echo ""
echo "3. To send requests, use the session ID from the endpoint event above"
echo "   Example:"
echo ""
echo "   SESSION_ID='session_xxxxx'"
echo "   curl -X POST \"$BASE_URL/message?session=\$SESSION_ID\" \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\",\"capabilities\":{},\"clientInfo\":{\"name\":\"test\",\"version\":\"1.0\"}}}'"
echo ""

# Keep SSE connection open
wait $SSE_PID
