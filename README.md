# M-Pesa MCP Server (Go Implementation)

A Model Context Protocol (MCP) server for integrating M-Pesa payments with AI agents, built in Go.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **⚠️ IMPORTANT: This handles real money transactions!**  
> Before using in production, read [Understanding the Payment Flow](./docs/PAYMENT_FLOW.md) and review [Security Best Practices](./docs/SECURITY.md). Always test thoroughly in sandbox mode first.

## 🎯 What This Does

This MCP server allows AI agents (like those built with Google ADK, LangChain, Claude Desktop) to:
- Initiate M-Pesa STK Push payments (prompt customers to pay from their phones)
- Generate M-Pesa QR codes for payment
- Check authentication token status
- **[NEW]** Handle M-Pesa payment callbacks (webhooks)

**New to this project?** → Start with the [Quick Start Guide](./QUICKSTART.md) (5 minutes)

**Want to integrate?** → Check out [Integration Examples](./examples/README.md)

## 💰 Understanding the Payment Flow

This is critical to understand before using this server:

```
CUSTOMER (PartyA)  →  Pays Money  →  YOUR BUSINESS (PartyB)
   254712345678                         Your Shortcode/Till
```

**The money flow is ALWAYS:**
- **FROM**: Customer's M-Pesa account (the phone number you specify)
- **TO**: YOUR business account (configured in `.env` as `BUSINESS_SHORTCODE`)

**All payments go TO your business account, not to the agent or server operator.**

📖 **Read the full guide**: [Payment Flow & Use Cases](./docs/PAYMENT_FLOW.md)

## 📚 Documentation

### Getting Started
- [Setup Guide](./docs/SETUP.md) - Detailed installation and configuration
- [Payment Flow Guide](./docs/PAYMENT_FLOW.md) - Understanding who pays who

### Integration
- [Integration Guide](./docs/INTEGRATION.md) - Connect with AI frameworks (Google ADK, LangChain, Claude, etc.)
- [API Reference](./docs/API_REFERENCE.md) - Complete tool documentation
- [Integration Examples](./examples/README.md) - Working code examples

### Deployment & Operations
- [Deployment Guide](./docs/DEPLOYMENT.md) - Docker, Kubernetes, Cloud platforms
- [Security Best Practices](./docs/SECURITY.md) - Production security & compliance
- [Troubleshooting](./docs/TROUBLESHOOTING.md) - Common issues and solutions

### Additional Resources
- [Use Cases & Examples](./docs/USE_CASES.md) - Real-world scenarios
- [Callback Implementation](./docs/CALLBACK_IMPLEMENTATION.md) - Handle payment confirmations
- [Architecture](./docs/ARCHITECTURE.md) - System design and structure
- [Contributing](./docs/CONTRIBUTING.md) - How to contribute

## 🚀 Quick Start

### Prerequisites

1. **Go 1.22+** installed
2. **Safaricom Daraja API credentials** - Get them at https://daraja.safaricom.co.ke/

### Installation

```bash
# Clone the repository
git clone https://github.com/skanenje/mpesa-mcp.git
cd mpesa-mcp

# Configure environment
cp .env.example .env
# Edit .env with your Daraja credentials
# ⚠️ IMPORTANT: Use RAW passkey, not base64 encoded!
# See docs/PASSKEY_GUIDE.md for details

# Build and run
go build -o mpesa-mcp cmd/server/main.go
./mpesa-mcp
```

Server will start on `http://localhost:8080`

**Full setup instructions**: [Setup Guide](./docs/SETUP.md)  
**Passkey configuration**: [Passkey Guide](./docs/PASSKEY_GUIDE.md)

## 🛠️ Available Tools

### 1. `stk_push`
Initiates an STK Push payment request (sends payment prompt to customer's phone).

```json
{
  "name": "stk_push",
  "arguments": {
    "amount": 1000,
    "phone_number": "254712345678"
  }
}
```

### 2. `generate_qr_code`
Generates an M-Pesa QR code that customers can scan to pay.

```json
{
  "name": "generate_qr_code",
  "arguments": {
    "merchant_name": "My Store",
    "ref_no": "ORDER123",
    "amount": 500,
    "trx_code": "BG",
    "cp_identifier": "123456"
  }
}
```

### 3. `get_token_status`
Check OAuth token status (useful for debugging authentication issues).

**Full API documentation**: [API Reference](./docs/API_REFERENCE.md)

## 🔌 Integration Examples

### Python Client

```python
from mpesa_mcp_client import MPesaMCPClient

client = MPesaMCPClient()
client.connect()

result = client.call_tool("stk_push", {
    "amount": 100,
    "phone_number": "254712345678"
})
```

### Google ADK Agent

```python
from google.adk import Agent

agent = Agent(
    name="payment-assistant",
    model="gemini-2.0-flash-exp",
    tools=[mpesa.stk_push],
    instruction="You help process M-Pesa payments."
)

response = agent.run("Charge customer 254712345678 KES 1000")
```

**More examples**: [Integration Guide](./docs/INTEGRATION.md) | [Examples Directory](./examples/)

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│     AI Agent Layer                      │
│  (Google ADK, LangChain, Claude, etc.)  │
└────────────┬────────────────────────────┘
             │ JSON-RPC over SSE/HTTP
┌────────────▼────────────────────────────┐
│     M-Pesa MCP Server (Go)              │
│  - HTTP Server (Port 8080)              │
│  - MCP Tool Handlers                    │
│  - M-Pesa Client                        │
└────────────┬────────────────────────────┘
             │ HTTPS + Bearer Token
┌────────────▼────────────────────────────┐
│     Safaricom Daraja API                │
└────────────┬────────────────────────────┘
             │ Push Notification
┌────────────▼────────────────────────────┐
│     Customer's Phone                    │
└─────────────────────────────────────────┘
```

**Detailed architecture**: [ARCHITECTURE.md](./ARCHITECTURE.md)

## 📝 Testing

```bash
# Start the server
./mpesa-mcp

# Test health endpoint
curl http://localhost:8080/health

# Test STK Push
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

**More testing examples**: [Troubleshooting Guide](./docs/TROUBLESHOOTING.md)

## 🚀 Deployment

### Docker

```bash
docker build -t mpesa-mcp .
docker run -p 8080:8080 \
  -e MPESA_CONSUMER_KEY=your_key \
  -e MPESA_CONSUMER_SECRET=your_secret \
  mpesa-mcp
```

### Cloud Platforms

Supports deployment to:
- Google Cloud Run
- AWS ECS/Fargate
- Heroku
- Kubernetes

**Full deployment guide**: [Deployment Guide](./docs/DEPLOYMENT.md)

## 🔐 Security

Before going to production:

- [ ] Use production Daraja credentials (not sandbox)
- [ ] Set up HTTPS with valid SSL certificate
- [ ] Implement callback endpoint verification
- [ ] Enable rate limiting
- [ ] Configure logging and monitoring
- [ ] Complete KYC verification with Safaricom

**Full security checklist**: [Security Best Practices](./docs/SECURITY.md)

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

**Areas we need help:**
- Adding more M-Pesa API operations (B2C, B2B, etc.)
- Writing tests (unit and integration)
- Creating integration examples for more frameworks
- Improving documentation

## 📄 License

MIT License - feel free to use in your projects!

## 📚 Additional Resources

- [Safaricom Daraja API Docs](https://developer.safaricom.co.ke/Documentation)
- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Go MCP SDK Documentation](https://github.com/modelcontextprotocol/go-sdk)
- [Google ADK Documentation](https://github.com/google/adk)

## ⚠️ Disclaimer

This is for educational purposes. Always:
- Test thoroughly in sandbox before production
- Follow Safaricom's terms of service
- Implement proper error handling
- Secure your credentials
- Validate all payments server-side

---

**Questions?** Check the [Troubleshooting Guide](./docs/TROUBLESHOOTING.md) or open an issue.
