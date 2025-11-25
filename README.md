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

## 🔌 Integrating with Google ADK

To use this MCP server with Google ADK:

1. **Build the MCP server**:
   ```bash
   go build -o mpesa-mcp cmd/server/main.go
   ```

2. **In your Google ADK code**, connect to the MCP server:

```python
from google.adk import Agent
import subprocess

# Start MCP server process
mcp_process = subprocess.Popen(
    ["./mpesa-mcp"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
)

# Create agent with MCP connection
agent = Agent(
    name="mpesa-payment-agent",
    mcp_servers=[
        {
            "command": "./mpesa-mcp",
            "transport": "stdio"
        }
    ]
)

# Now your agent can use M-Pesa tools
response = agent.run("Charge customer 254712345678 KES 1000 for order #12345")
```

## 🛠️ Available Tools

### 1. `stk_push`
Initiates an STK Push payment request.

**Parameters:**
- `amount` (int): Amount in KES
- `phone_number` (string): Phone number in format `254XXXXXXXXX` or `0XXXXXXXXX`

**Example:**
```json
{
  "amount": 1000,
  "phone_number": "254712345678"
}
```

### 2. `generate_qr_code`
Generates an M-Pesa QR code.

**Parameters:**
- `merchant_name` (string): Business name
- `ref_no` (string): Transaction reference
- `amount` (int): Amount in KES
- `trx_code` (string): `BG` (Buy Goods), `PB` (Paybill), `SM` (Send Money), etc.
- `cp_identifier` (string): Till number, paybill, or phone number

**Example:**
```json
{
  "merchant_name": "My Store",
  "ref_no": "ORDER123",
  "amount": 500,
  "trx_code": "BG",
  "cp_identifier": "123456"
}
```

### 3. `get_token_status`
Check OAuth token status (useful for debugging).

## 📝 Testing

### Test STK Push

Using Claude Desktop or any MCP client:

```
Please use the stk_push tool to charge 254712345678 KES 100
```

### Test QR Code Generation

```
Generate a Buy Goods QR code for KES 500 at "My Store" with till number 123456
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