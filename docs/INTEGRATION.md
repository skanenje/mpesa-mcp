# Integration Guide

How to integrate the M-Pesa MCP Server with various AI agent frameworks.

## Overview

This MCP server can be integrated with various AI agent frameworks. The server runs as an HTTP service with Server-Sent Events (SSE) for real-time communication.

## Server Endpoints

```bash
# Start the server
./mpesa-mcp

# Server will start on http://localhost:8080
# Endpoints:
# - GET  /sse      - SSE connection endpoint
# - POST /message  - Send JSON-RPC messages
# - GET  /health   - Health check
```

## Option 1: SSE Transport (HTTP-based) - Recommended

### Python Client

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

result = client.call_tool("stk_push", {
    "amount": 100,
    "phone_number": "254712345678"
})
print(result)
```

## Option 2: Google ADK Integration

### Using SSE Transport

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

## Option 3: LangChain Integration

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
        # Connect to SSE
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

## Option 4: Claude Desktop Integration

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

## Option 5: Node.js/TypeScript Integration

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

## Option 6: Custom REST API Wrapper

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

## Testing Your Integration

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
