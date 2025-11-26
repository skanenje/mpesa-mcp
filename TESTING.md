# Testing M-Pesa MCP Server

This guide explains how to use the test scripts to interact with your M-Pesa MCP server.

## Prerequisites

1. **Server must be running**: Make sure your server is running on port 8080
   ```bash
   go run cmd/server/main.go
   ```

2. **Two terminal windows**: You'll need two terminals - one for the SSE connection and one for sending requests

## Available Test Scripts

### 1. `scripts/test_client.sh` - Interactive Guide (Recommended for Beginners)

This script connects to the SSE endpoint and shows you example commands to run.

**Usage:**
```bash
cd /home/zedolph/project-zero/mpesa-mcp
./scripts/test_client.sh
```

**What it does:**
- Connects to the SSE endpoint
- Displays the session ID
- Shows example curl commands you can copy and paste
- Keeps the SSE connection alive

**Example output:**
```
📡 Step 1: Connecting to SSE endpoint...
event: endpoint
data: /message?session=session_1764146744985728944

Then, in another terminal, run one of these test commands:
...
```

### 2. `scripts/test-sse.sh` - Basic SSE Connection Test

This script establishes an SSE connection and shows you how to use it.

**Usage:**
```bash
cd /home/zedolph/project-zero/mpesa-mcp
./scripts/test-sse.sh
```

**What it does:**
- Tests the health endpoint
- Connects to SSE
- Shows example commands

### 3. `scripts/test-requests.sh` - Automated Request Testing

This script sends a series of test requests to an existing session.

**Usage:**
```bash
# First, get a session ID from test_client.sh or test-sse.sh
# Then run:
./scripts/test-requests.sh session_1764146744985728944
```

**What it does:**
- Sends an `initialize` request
- Lists available tools
- Calls the `get_token_status` tool

## Step-by-Step Testing Workflow

### Method 1: Using test_client.sh (Easiest)

**Terminal 1:**
```bash
cd /home/zedolph/project-zero/mpesa-mcp
./scripts/test_client.sh
```

Wait for the session ID to appear, then copy it.

**Terminal 2:**
```bash
# Replace SESSION_ID with the actual session ID from Terminal 1
curl -X POST "http://localhost:8080/message?session=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "tools/call",
    "params": {
      "name": "stk_push",
      "arguments": {
        "phone_number": "254723975141",
        "amount": 10
      }
    }
  }'
```

### Method 2: Using test-sse.sh + test-requests.sh

**Terminal 1:**
```bash
cd /home/zedolph/project-zero/mpesa-mcp
./scripts/test-sse.sh
```

Copy the session ID from the output.

**Terminal 2:**
```bash
cd /home/zedolph/project-zero/mpesa-mcp
./scripts/test-requests.sh session_1764146744985728944
```

## Available MCP Tools

### 1. STK Push
Initiate an M-Pesa payment request.

```bash
curl -X POST "http://localhost:8080/message?session=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "tools/call",
    "params": {
      "name": "stk_push",
      "arguments": {
        "phone_number": "254723975141",
        "amount": 10
      }
    }
  }'
```

### 2. Generate QR Code
Generate an M-Pesa QR code.

```bash
curl -X POST "http://localhost:8080/message?session=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "2",
    "method": "tools/call",
    "params": {
      "name": "generate_qr_code",
      "arguments": {
        "merchant_name": "Test Shop",
        "ref_no": "INV001",
        "amount": 100,
        "trx_code": "BG",
        "cp_identifier": "174379"
      }
    }
  }'
```

### 3. Get Token Status
Check the M-Pesa access token status.

```bash
curl -X POST "http://localhost:8080/message?session=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "3",
    "method": "tools/call",
    "params": {
      "name": "get_token_status",
      "arguments": {}
    }
  }'
```

## Monitoring Logs

With the new logging system, you'll see detailed output in your server terminal:

```
[SSE] Received message request from 127.0.0.1:xxxxx
[SSE] Looking for session: session_1764146744985728944
[SSE] Session found: session_1764146744985728944
[SSE] Parsed JSON-RPC request - Method: tools/call, ID: "1"
[Tool:stk_push] Called with params: map[amount:10 phone_number:254723975141]
[Tool:stk_push] Initiating STK Push - Amount: 10, Phone: 254723975141
[Tool:stk_push] Success - MerchantRequestID: xxx, CheckoutRequestID: xxx
```

## Common Issues

### "Session not found"
- Make sure you copied the correct session ID
- Ensure the SSE connection (Terminal 1) is still running
- Sessions expire when the SSE connection closes

### "Invalid JSON: json: cannot unmarshal number into Go struct field Request.ID"
- Make sure `id` is a **string** not a number: `"id": "1"` not `"id": 1`
- All test scripts have been updated with the correct format

### No response in Terminal 2
- Check Terminal 1 (SSE connection) - responses appear there as SSE events
- Check your server logs for errors

## Tips

1. **Keep Terminal 1 open**: The SSE connection must stay alive
2. **Watch server logs**: They show detailed information about each request
3. **Use unique IDs**: Each request should have a unique ID string
4. **Check .env file**: Make sure your M-Pesa credentials are configured
