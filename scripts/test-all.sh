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

# Check if port is already in use and kill it
echo "🔍 Checking if port $SERVER_PORT is in use..."
if lsof -Pi :$SERVER_PORT -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo "⚠️  Port $SERVER_PORT is in use. Killing existing process..."
    lsof -ti:$SERVER_PORT | xargs kill -9 2>/dev/null || true
    sleep 2
fi

echo "🚀 Starting M-Pesa MCP Server..."
PORT=$SERVER_PORT go run cmd/server/main.go > server.log 2>&1 &
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
# We use curl with -N (no buffer) and timeout to read the first few lines to find the session ID
SESSION_ID=$(timeout 5 curl -N -s "$BASE_URL/sse" | grep -m 1 "sessionid=" | sed -E 's/.*sessionid=([^[:space:]]+).*/\1/')

if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to get Session ID. Check server logs:"
    cat server.log
    exit 1
fi

echo "✅ Got Session ID: $SESSION_ID"
echo ""

# Run the requests script
./scripts/test-requests.sh "$SESSION_ID"

echo ""
echo "✅ Test completed!"
echo ""
echo "📝 To test M-Pesa callbacks:"
echo "   1. Start ngrok: ngrok http 8080"
echo "   2. Update .env with: CALLBACK_URL=https://your-ngrok-url.ngrok-free.app/callback"
echo "   3. Restart server and trigger an STK push"
echo "   4. Monitor ngrok dashboard at http://localhost:4040"
echo ""
