#!/bin/bash

# Automate the entire testing flow:
# 1. Start Server
# 2. Get Session ID from SSE
# 3. Run Test Requests

# Ensure we are in the project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR/.."
echo "📂 Working directory: $(pwd)"

SERVER_PORT=8080
BASE_URL="http://localhost:$SERVER_PORT"

echo "🚀 Starting M-Pesa MCP Server..."
go run cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!
echo "   Server PID: $SERVER_PID"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping server..."
    kill $SERVER_PID
    rm -f server.log
}
trap cleanup EXIT

# Wait for server to be ready
echo "⏳ Waiting for server to start..."
sleep 5

# Connect to SSE and extract session ID
echo "📡 Connecting to SSE to get Session ID..."
# We use curl with -N (no buffer) and read the first few lines to find the session ID
SESSION_ID=$(curl -N -s "$BASE_URL/sse" | grep -m 1 "sessionid=" | sed -E 's/.*sessionid=([^[:space:]]+).*/\1/')

if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to get Session ID. Check server logs:"
    cat server.log
    exit 1
fi

echo "✅ Got Session ID: $SESSION_ID"
echo ""

# Run the requests script
./scripts/test-requests.sh "$SESSION_ID"

echo "⏳ Waiting 30s for M-Pesa callback..."
sleep 30
