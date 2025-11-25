# M-Pesa MCP Server (Go Implementation)

A Model Context Protocol (MCP) server for integrating M-Pesa payments with AI agents, built in Go.

## 🎯 What This Does

This MCP server allows AI agents (like those built with Google ADK) to:
- Initiate M-Pesa STK Push payments (prompt customers to pay from their phones)
- Generate M-Pesa QR codes for payment
- Check authentication token status

## 📋 Prerequisites

1. **Go 1.22+** installed
2. **Safaricom Daraja API credentials**:
   - Go to https://developer.safaricom.co.ke/
   - Create an account
   - Create a **"Lipa Na M-Pesa Sandbox"** app (the one you have selected in your screenshot)
   - Note down your Consumer Key and Consumer Secret

## 🚀 Setup Instructions

### Step 1: Get Daraja Credentials

1. Visit [Safaricom Daraja Portal](https://developer.safaricom.co.ke/)
2. Log in and go to "My Apps"
3. Click "Create Sandbox App"
4. **Select "Lipa Na M-Pesa Sandbox"** (as shown in your second screenshot)
5. After creation, you'll get:
   - Consumer Key
   - Consumer Secret
   - Passkey (found in "Test Credentials")

### Step 2: Project Setup

```bash
# Clone or create your project directory
mkdir mpesa-mcp
cd mpesa-mcp

# Initialize Go module
go mod init github.com/yourusername/mpesa-mcp

# Install dependencies
go get github.com/modelcontextprotocol/go-sdk
go get github.com/joho/godotenv
```

### Step 3: Project Structure

The project follows Go best practices with clear separation of concerns:

```
mpesa-mcp/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── mpesa/
│   │   ├── client.go            # M-Pesa API client
│   │   ├── auth.go              # OAuth authentication
│   │   ├── stk_push.go          # STK Push operations
│   │   ├── qr_code.go           # QR code generation
│   │   └── types.go             # Shared data types
│   ├── mcp/
│   │   ├── server.go            # MCP server setup
│   │   ├── tools.go             # MCP tool handlers
│   │   └── prompts.go           # MCP prompt handlers
│   └── utils/
│       └── phone.go             # Utility functions
├── .env                         # Your credentials (DO NOT COMMIT)
├── .env.example                 # Example environment file
├── go.mod                       # Go module file
├── go.sum                       # Dependency checksums
└── README.md                    # This file
```

### Step 4: Configure Environment

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```bash
MPESA_CONSUMER_KEY=your_actual_consumer_key
MPESA_CONSUMER_SECRET=your_actual_consumer_secret
BASE_URL=https://sandbox.safaricom.co.ke
BUSINESS_SHORTCODE=174379
PASSKEY=your_test_passkey
CALLBACK_URL=https://your-callback-url.com/callback
ACCOUNT_REFERENCE=TestPayment
```

**Important Notes:**
- For sandbox: `BUSINESS_SHORTCODE` is usually `174379`
- Get `PASSKEY` from "Test Credentials" page in Daraja portal
- For `CALLBACK_URL`: use ngrok or webhook.site for testing
- **Never commit `.env` file to git!** Add it to `.gitignore`

### Step 5: Build and Run

```bash
# Run directly
go run cmd/server/main.go

# Or build and run
go build -o mpesa-mcp cmd/server/main.go
./mpesa-mcp
```

## 🔌 Integrating with Agentic Systems

This MCP server can be integrated with various AI agent frameworks. Below are examples for popular platforms:

### Option 1: SSE Transport (HTTP-based) - Recommended for Production

The server runs as an HTTP service with Server-Sent Events for real-time communication.

#### Start the Server:
```bash
# Build the server
go build -o mpesa-mcp cmd/server/main.go

# Run the server
./mpesa-mcp

# Server will start on http://localhost:8080
# Endpoints:
# - GET  /sse      - SSE connection endpoint
# - POST /message  - Send JSON-RPC messages
# - GET  /health   - Health check
```

#### Integration Example (Python):
```python
import requests
import json
import sseclient

class MPesaMCPClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.session_id = None
        self.message_endpoint = None
        
    def connect(self):
        """Establish SSE connection and get session ID"""
        response = requests.get(f"{self.base_url}/sse", stream=True)
        client = sseclient.SSEClient(response)
        
        for event in client.events():
            if event.event == "endpoint":
                self.message_endpoint = f"{self.base_url}{event.data}"
                # Extract session ID from endpoint
                self.session_id = event.data.split("session=")[1]
                print(f"Connected! Session: {self.session_id}")
                break
                
    def call_tool(self, tool_name, arguments):
        """Call an MCP tool"""
        message = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        response = requests.post(self.message_endpoint, json=message)
        return response.json()

# Usage
client = MPesaMCPClient()
client.connect()

# Initiate STK Push
result = client.call_tool("stk_push", {
    "amount": 100,
    "phone_number": "254712345678"
})
print(result)
```

---

### Option 2: Google ADK Integration

#### Method A: Using SSE Transport (Recommended)
```python
from google.adk import Agent
import requests

class MPesaTools:
    def __init__(self, mcp_url="http://localhost:8080"):
        self.mcp_url = mcp_url
        self.session_id = None
        self._connect()
    
    def _connect(self):
        # Connect to SSE endpoint
        response = requests.get(f"{self.mcp_url}/sse", stream=True)
        # Parse session from endpoint event
        # (implementation similar to above)
        pass
    
    def stk_push(self, phone_number: str, amount: int):
        """Initiate M-Pesa STK Push payment"""
        message = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "stk_push",
                "arguments": {
                    "phone_number": phone_number,
                    "amount": amount
                }
            }
        }
        endpoint = f"{self.mcp_url}/message?session={self.session_id}"
        response = requests.post(endpoint, json=message)
        return response.json()

# Create agent with M-Pesa tools
mpesa = MPesaTools()

agent = Agent(
    name="payment-assistant",
    model="gemini-2.0-flash-exp",
    tools=[mpesa.stk_push],
    instruction="""You are a payment assistant that helps process M-Pesa payments.
    When a user requests a payment, use the stk_push tool to initiate it."""
)

# Use the agent
response = agent.run("Charge customer 254712345678 KES 1000 for order #12345")
print(response)
```

#### Method B: Using STDIO Transport (Alternative)
If you prefer stdio transport, modify `cmd/server/main.go` to support it:

```python
from google.adk import Agent
import subprocess

# Start MCP server in stdio mode
agent = Agent(
    name="mpesa-payment-agent",
    model="gemini-2.0-flash-exp",
    mcp_servers=[
        {
            "command": "./mpesa-mcp",
            "args": ["--transport=stdio"],  # You'll need to add this flag
            "env": {
                "MPESA_CONSUMER_KEY": "your_key",
                "MPESA_CONSUMER_SECRET": "your_secret"
            }
        }
    ]
)

# Agent automatically discovers and uses M-Pesa tools
response = agent.run("Charge customer 254712345678 KES 1000")
```

---

### Option 3: LangChain Integration

```python
from langchain.agents import AgentExecutor, create_openai_functions_agent
from langchain.tools import Tool
from langchain_openai import ChatOpenAI
import requests

class MPesaMCPClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.session_id = None
        self._connect()
    
    def _connect(self):
        # Connect to SSE (implementation from above)
        pass
    
    def stk_push(self, phone_number: str, amount: int) -> str:
        """Initiate M-Pesa STK Push payment"""
        message = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "stk_push",
                "arguments": {
                    "phone_number": phone_number,
                    "amount": amount
                }
            }
        }
        endpoint = f"{self.base_url}/message?session={self.session_id}"
        response = requests.post(endpoint, json=message)
        return str(response.json())

# Create LangChain tools
mpesa_client = MPesaMCPClient()

tools = [
    Tool(
        name="mpesa_stk_push",
        func=lambda x: mpesa_client.stk_push(
            phone_number=x.split(",")[0].strip(),
            amount=int(x.split(",")[1].strip())
        ),
        description="Initiate M-Pesa STK Push payment. Input: 'phone_number, amount'"
    )
]

# Create agent
llm = ChatOpenAI(model="gpt-4")
agent = create_openai_functions_agent(llm, tools)
agent_executor = AgentExecutor(agent=agent, tools=tools)

# Use the agent
result = agent_executor.invoke({
    "input": "Charge customer 254712345678 KES 1000"
})
print(result)
```

---

### Option 4: Claude Desktop Integration

Add to your Claude Desktop MCP config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "mpesa": {
      "command": "/path/to/mpesa-mcp",
      "args": [],
      "env": {
        "MPESA_CONSUMER_KEY": "your_key",
        "MPESA_CONSUMER_SECRET": "your_secret",
        "BASE_URL": "https://sandbox.safaricom.co.ke",
        "BUSINESS_SHORTCODE": "174379",
        "PASSKEY": "your_passkey",
        "CALLBACK_URL": "https://your-callback.com/mpesa",
        "ACCOUNT_REFERENCE": "Payment"
      }
    }
  }
}
```

Then in Claude Desktop:
```
User: Charge customer 254712345678 KES 1000 for order #12345

Claude: I'll initiate an M-Pesa STK Push payment for that order.
[Uses stk_push tool automatically]
```

---

### Option 5: Custom REST API Wrapper

If you want a simpler REST API, create a wrapper:

```python
from flask import Flask, request, jsonify
import requests

app = Flask(__name__)
mcp_client = MPesaMCPClient("http://localhost:8080")
mcp_client.connect()

@app.route('/api/payment/stk-push', methods=['POST'])
def stk_push():
    data = request.json
    result = mcp_client.call_tool("stk_push", {
        "phone_number": data['phone_number'],
        "amount": data['amount']
    })
    return jsonify(result)

@app.route('/api/payment/qr-code', methods=['POST'])
def generate_qr():
    data = request.json
    result = mcp_client.call_tool("generate_qr_code", data)
    return jsonify(result)

if __name__ == '__main__':
    app.run(port=5000)
```

Now any system can use simple REST calls:
```bash
curl -X POST http://localhost:5000/api/payment/stk-push \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "254712345678", "amount": 1000}'
```

---

### Option 6: Node.js/TypeScript Integration

```typescript
import axios from 'axios';
import EventSource from 'eventsource';

class MPesaMCPClient {
  private baseUrl: string;
  private sessionId: string | null = null;
  private messageEndpoint: string | null = null;

  constructor(baseUrl: string = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
  }

  async connect(): Promise<void> {
    return new Promise((resolve) => {
      const eventSource = new EventSource(`${this.baseUrl}/sse`);
      
      eventSource.addEventListener('endpoint', (event) => {
        this.messageEndpoint = `${this.baseUrl}${event.data}`;
        this.sessionId = event.data.split('session=')[1];
        console.log(`Connected! Session: ${this.sessionId}`);
        resolve();
      });
    });
  }

  async callTool(toolName: string, args: any): Promise<any> {
    const message = {
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/call',
      params: {
        name: toolName,
        arguments: args
      }
    };

    const response = await axios.post(this.messageEndpoint!, message);
    return response.data;
  }

  async stkPush(phoneNumber: string, amount: number) {
    return this.callTool('stk_push', {
      phone_number: phoneNumber,
      amount: amount
    });
  }
}

// Usage
const client = new MPesaMCPClient();
await client.connect();

const result = await client.stkPush('254712345678', 1000);
console.log(result);
```

---

### Testing Your Integration

```bash
# 1. Start the MCP server
./mpesa-mcp

# 2. Test SSE connection
curl -N http://localhost:8080/sse

# 3. Test health endpoint
curl http://localhost:8080/health

# 4. Test STK Push (replace SESSION_ID with actual session from SSE)
curl -X POST "http://localhost:8080/message?session=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "stk_push",
      "arguments": {
        "phone_number": "254712345678",
        "amount": 100
      }
    }
  }'
```

## 🛠️ Available Tools

### 1. `stk_push`
Initiates an STK Push payment request (sends payment prompt to customer's phone).

**Parameters:**
- `amount` (int): Amount in KES
- `phone_number` (string): Phone number in format `254XXXXXXXXX` or `0XXXXXXXXX`

**JSON-RPC Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "stk_push",
    "arguments": {
      "amount": 1000,
      "phone_number": "254712345678"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{
      "type": "text",
      "text": "STK Push initiated successfully!\n\nMerchant Request ID: 29115-34620561-1\nCheckout Request ID: ws_CO_191220191020363925\nCustomer Message: Success. Request accepted for processing"
    }],
    "MerchantRequestID": "29115-34620561-1",
    "CheckoutRequestID": "ws_CO_191220191020363925",
    "ResponseCode": "0",
    "ResponseDescription": "Success. Request accepted for processing",
    "CustomerMessage": "Success. Request accepted for processing"
  }
}
```

---

### 2. `generate_qr_code`
Generates an M-Pesa QR code that customers can scan to pay.

**Parameters:**
- `merchant_name` (string): Business name
- `ref_no` (string): Transaction reference
- `amount` (int): Amount in KES
- `trx_code` (string): Transaction type
  - `BG` - Buy Goods (Till Number)
  - `PB` - Paybill
  - `SM` - Send Money (Phone Number)
  - `SB` - Send to Business
  - `WA` - Withdraw Cash
- `cp_identifier` (string): Till number, paybill number, or phone number

**JSON-RPC Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "generate_qr_code",
    "arguments": {
      "merchant_name": "My Store",
      "ref_no": "ORDER123",
      "amount": 500,
      "trx_code": "BG",
      "cp_identifier": "123456"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{
      "type": "text",
      "text": "QR Code generated successfully!\n\nRequest ID: QRCode-123456789\nQR Code (Base64): iVBORw0KGgoAAAANSUhEUgAA..."
    }],
    "ResponseCode": "00",
    "RequestID": "QRCode-123456789",
    "ResponseDescription": "The service request is processed successfully.",
    "QRCode": "iVBORw0KGgoAAAANSUhEUgAA..."
  }
}
```

---

### 3. `get_token_status`
Check OAuth token status (useful for debugging authentication issues).

**JSON-RPC Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "get_token_status",
    "arguments": {}
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [{
      "type": "text",
      "text": "Token Status:\n- Has Token: true\n- Expires At: 2024-11-25 15:30:00\n- Is Valid: true"
    }],
    "has_token": true,
    "expires_at": "2024-11-25 15:30:00",
    "is_valid": true
  }
}
```

## � Practical Use Cases

### E-commerce Checkout Agent
```python
# Agent that processes orders and collects payment
agent_prompt = """
You are an e-commerce checkout assistant. When a customer confirms their order:
1. Calculate the total amount
2. Use stk_push to request payment from their phone number
3. Inform them to check their phone and enter M-Pesa PIN
4. Confirm when payment is initiated
"""

# Example conversation:
# User: "I want to buy 2 shirts at KES 1500 each"
# Agent: "That's KES 3000 total. What's your M-Pesa number?"
# User: "0712345678"
# Agent: [calls stk_push] "Payment request sent! Check your phone to complete."
```

### Restaurant/Retail POS Agent
```python
# Agent that generates QR codes for in-person payments
agent_prompt = """
You are a point-of-sale assistant. When a customer is ready to pay:
1. Calculate the bill total
2. Generate a QR code using generate_qr_code
3. Display the QR code for customer to scan
"""

# Example:
# Staff: "Customer bill is KES 2500"
# Agent: [generates QR code] "QR code ready! Customer can scan to pay."
```

### Subscription Payment Agent
```python
# Agent that handles recurring payments
agent_prompt = """
You manage subscription payments. When a subscription is due:
1. Check customer's phone number
2. Send payment request via stk_push
3. Log the transaction
4. Send confirmation email
"""
```

### Customer Support Agent
```python
# Agent that helps with payment issues
agent_prompt = """
You help customers with M-Pesa payments. You can:
- Check token status if payments are failing
- Resend payment requests
- Generate new QR codes if needed
- Explain payment steps to customers
"""
```

---

## 📝 Testing

### Test with cURL

```bash
# 1. Start the server
./mpesa-mcp

# 2. Connect to SSE (in another terminal)
curl -N http://localhost:8080/sse
# Note the session ID from the endpoint event

# 3. Test STK Push
curl -X POST "http://localhost:8080/message?session=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "stk_push",
      "arguments": {
        "phone_number": "254712345678",
        "amount": 100
      }
    }
  }'

# 4. Test QR Code Generation
curl -X POST "http://localhost:8080/message?session=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "generate_qr_code",
      "arguments": {
        "merchant_name": "Test Store",
        "ref_no": "TEST123",
        "amount": 500,
        "trx_code": "BG",
        "cp_identifier": "123456"
      }
    }
  }'
```

### Test with AI Agents

#### Using Claude Desktop or any MCP client:

```
Please use the stk_push tool to charge 254712345678 KES 100
```

#### Using Google ADK:

```python
response = agent.run("Charge customer 254712345678 KES 100 for order #12345")
```

#### Using LangChain:

```python
result = agent_executor.invoke({
    "input": "Generate a Buy Goods QR code for KES 500 at My Store with till 123456"
})
```

## 🔍 Understanding the Daraja API

### Authentication Flow
1. Server starts → Gets OAuth token using Consumer Key/Secret
2. Token valid for ~60 minutes
3. Server auto-refreshes token every 50 minutes
4. Token used in `Authorization: Bearer <token>` header for all API calls

### STK Push Flow
1. Your agent calls `stk_push` tool
2. Server sends request to Daraja API
3. Customer gets push notification on their phone
4. Customer enters M-Pesa PIN
5. Daraja sends callback to your CALLBACK_URL
6. You verify payment status

### Phone Number Format
- Kenyan numbers must start with `254` (country code)
- Example: `254712345678` (NOT `+254` or `0712345678`)
- Server automatically formats numbers starting with `0`

## 🔐 Security Best Practices

1. **Never commit credentials** - Use `.env` and add to `.gitignore`
2. **Sandbox vs Production**:
   - Use sandbox for development/testing
   - Switch to production URL only when ready
   - Production requires KYC verification
3. **Callback URL**: In production, use HTTPS and verify callbacks
4. **Token Management**: Server handles this automatically

## 🐛 Troubleshooting

### "Invalid Access Token"
- Token might have expired
- Check `get_token_status` tool
- Verify Consumer Key/Secret are correct
- Ensure BASE_URL is correct (sandbox vs production)

### "Invalid Phone Number"
- Must be Kenyan number (254...)
- Remove spaces and special characters
- Server auto-formats `07XX` to `2547XX`

### "Callback Not Received"
- For local testing, use ngrok: `ngrok http 8080`
- Update CALLBACK_URL in `.env`
- Check firewall/network settings

### "Build Errors"
- Ensure Go 1.22+ installed: `go version`
- Run `go mod tidy` to sync dependencies
- Check import paths in code

## 📚 Additional Resources

- [Safaricom Daraja API Docs](https://developer.safaricom.co.ke/Documentation)
- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Go MCP SDK Documentation](https://github.com/modelcontextprotocol/go-sdk)

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Test thoroughly with sandbox
4. Submit a pull request

## 📄 License

MIT License - feel free to use in your projects!

## ⚠️ Disclaimer

This is for educational purposes. Always:
- Test thoroughly in sandbox before production
- Follow Safaricom's terms of service
- Implement proper error handling
- Secure your credentials
- Validate all payments server-side