# Quick Start Guide

Get up and running with M-Pesa MCP server in 5 minutes!

> **⚠️ READ THIS FIRST:**  
> This server collects payments FROM customers TO your business account.  
> All money goes to YOUR M-Pesa business shortcode (configured in `.env`).  
> See [Understanding the Payment Flow](./README.md#-understanding-the-payment-flow-important) for details.

## Prerequisites

- Go 1.22+ installed
- Safaricom Daraja API credentials ([Get them here](https://developer.safaricom.co.ke/))
- Python 3.8+ (for examples)
- **A clear understanding of who pays who** (read the warning above!)

## Step 1: Setup (2 minutes)

```bash
# Clone the repository
git clone <your-repo-url>
cd mpesa-mcp

# Copy environment template
cp .env.example .env

# Edit .env with your Daraja credentials
nano .env  # or use your favorite editor
```

**Required credentials in `.env`:**
```bash
MPESA_CONSUMER_KEY=your_consumer_key_here
MPESA_CONSUMER_SECRET=your_consumer_secret_here
BASE_URL=https://sandbox.safaricom.co.ke
BUSINESS_SHORTCODE=174379
PASSKEY=your_passkey_here
CALLBACK_URL=https://your-callback-url.com/callback
ACCOUNT_REFERENCE=TestPayment
```

## Step 2: Run the Server (1 minute)

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

You should see:
```
🚀 M-Pesa MCP Server starting on :8080
📡 SSE endpoint: http://localhost:8080/sse
💬 Message endpoint: http://localhost:8080/message
❤️  Health check: http://localhost:8080/health
```

## Step 3: Test It (2 minutes)

### Option A: Using cURL

```bash
# In a new terminal, test health
curl http://localhost:8080/health

# Connect to SSE
curl -N http://localhost:8080/sse
# Note the session ID from the output

# Send a payment request (replace SESSION_ID)
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

### Option B: Using Python Example

```bash
# Install Python dependencies
cd examples
pip install -r requirements.txt

# Run the example
python python_client.py
```

## Step 4: Integrate with Your Agent

### Google ADK

```python
from examples.google_adk_agent import create_payment_agent

agent = create_payment_agent()
response = agent.run("Charge customer 254712345678 KES 1000")
print(response)
```

### Custom Integration

```python
from examples.python_client import MPesaMCPClient

client = MPesaMCPClient()
client.connect()

# Charge customer
result = client.stk_push("254712345678", 1000)
print(result)

# Generate QR code
result = client.generate_qr_code(
    merchant_name="My Store",
    amount=500,
    trx_code="BG",
    cp_identifier="123456"
)
print(result)
```

## Common Issues

### "Failed to load configuration"
**Problem:** Missing or invalid `.env` file  
**Solution:** Copy `.env.example` to `.env` and add your credentials

### "Invalid Access Token"
**Problem:** Wrong Daraja credentials  
**Solution:** Double-check your Consumer Key and Secret from Daraja portal

### "Connection refused"
**Problem:** Server not running  
**Solution:** Start the server with `go run cmd/server/main.go`

### "Invalid phone number"
**Problem:** Wrong phone format  
**Solution:** Use `254XXXXXXXXX` or `0XXXXXXXXX` (Kenyan numbers only)

## Next Steps

1. **Read the full README** - [README.md](./README.md)
2. **Explore examples** - [examples/README.md](./examples/README.md)
3. **Understand architecture** - [ARCHITECTURE.md](./ARCHITECTURE.md)
4. **Deploy to production** - See deployment guide in README

## Getting Help

- Check [Troubleshooting section](./README.md#-troubleshooting) in README
- Review [Daraja API Documentation](https://developer.safaricom.co.ke/Documentation)
- Open an issue on GitHub

## Production Checklist

Before going live:

- [ ] Get production Daraja credentials (not sandbox)
- [ ] Set up HTTPS with valid SSL certificate
- [ ] Implement callback endpoint for payment confirmations
- [ ] Add proper logging and monitoring
- [ ] Set up error alerting
- [ ] Test with real payments (small amounts first!)
- [ ] Configure firewall and security rules
- [ ] Set up backup and disaster recovery

---

**Ready to build?** Check out the [integration examples](./examples/README.md) for your framework!
