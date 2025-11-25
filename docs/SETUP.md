# Setup Guide

Complete setup instructions for the M-Pesa MCP Server.

## Prerequisites

1. **Go 1.22+** installed
2. **Safaricom Daraja API credentials**:
   - Go to https://developer.safaricom.co.ke/
   - Create an account
   - Create a **"Lipa Na M-Pesa Sandbox"** app
   - Note down your Consumer Key and Consumer Secret

## Step 1: Get Daraja Credentials

1. Visit [Safaricom Daraja Portal](https://developer.safaricom.co.ke/)
2. Log in and go to "My Apps"
3. Click "Create Sandbox App"
4. **Select "Lipa Na M-Pesa Sandbox"**
5. After creation, you'll get:
   - Consumer Key
   - Consumer Secret
   - Passkey (found in "Test Credentials")

## Step 2: Project Setup

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

## Step 3: Project Structure

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

## Step 4: Configure Environment

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

## Step 5: Build and Run

```bash
# Run directly
go run cmd/server/main.go

# Or build and run
go build -o mpesa-mcp cmd/server/main.go
./mpesa-mcp
```

## Configuration Breakdown

In your `.env` file:

```bash
# YOUR business credentials (who receives the money)
BUSINESS_SHORTCODE=174379        # Your till/paybill number
PASSKEY=your_passkey             # Your business passkey

# API credentials (for authentication)
MPESA_CONSUMER_KEY=xxx           # From Daraja portal
MPESA_CONSUMER_SECRET=xxx        # From Daraja portal

# Where M-Pesa sends payment confirmations
CALLBACK_URL=https://yourdomain.com/callback

# What customers see on their phone
ACCOUNT_REFERENCE=YourBusinessName
```

**Key Points:**
- `BUSINESS_SHORTCODE` = YOUR business account (money goes here)
- `CALLBACK_URL` = YOUR server (to receive payment confirmations)
- Customer phone number = Specified in each payment request (money comes from here)

## Sandbox vs Production

| Aspect | Sandbox | Production |
|--------|---------|------------|
| Money | Fake (test only) | Real money |
| Shortcode | 174379 (shared test) | Your unique business number |
| Phone Numbers | Test numbers only | Real customer numbers |
| Verification | None required | KYC + Business registration |
| Callbacks | Optional | Required |
| URL | sandbox.safaricom.co.ke | api.safaricom.co.ke |

**⚠️ NEVER use production credentials in code examples or public repositories!**
