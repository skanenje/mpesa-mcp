# M-Pesa MCP Integration Examples

This directory contains practical examples of integrating the M-Pesa MCP server with various AI agent frameworks.

## Prerequisites

1. **Start the M-Pesa MCP server:**
   ```bash
   cd ..
   go run cmd/server/main.go
   ```

2. **Install Python dependencies:**
   ```bash
   pip install requests sseclient-py
   ```

## Examples

### 1. Python Client (`python_client.py`)

Basic Python client demonstrating direct integration with the MCP server.

**Features:**
- SSE connection management
- STK Push payments
- QR code generation
- Token status checking

**Run:**
```bash
python python_client.py
```

**Output:**
```
============================================================
M-Pesa MCP Client - Python Example
============================================================

1. Checking server health...
✓ Server is healthy

2. Connecting to MCP server...
✓ Connected! Session: session_1234567890

3. Checking OAuth token status...
  Result: {
    "has_token": true,
    "expires_at": "2024-11-25 15:30:00",
    "is_valid": true
  }

4. Initiating STK Push payment...
  Result: {
    "status": "accepted",
    "request_id": 1
  }

5. Generating QR code...
  Result: {
    "status": "accepted",
    "request_id": 2
  }
```

---

### 2. Google ADK Agent (`google_adk_agent.py`)

AI agent built with Google ADK that can process M-Pesa payments.

**Features:**
- Natural language payment processing
- Automatic phone number formatting
- Conversational payment assistance
- QR code generation

**Install Google ADK:**
```bash
pip install google-adk
```

**Run:**
```bash
python google_adk_agent.py
```

**Example Conversations:**
```
User: Charge customer 254712345678 KES 1000 for order #12345
Agent: I'll initiate an M-Pesa payment request for KES 1000 to 254712345678...
      [Calls charge_customer tool]
      Payment request sent! The customer will receive a prompt on their phone.

User: Generate a QR code for KES 500 at My Store with till 123456
Agent: I'll generate a QR code for your store...
      [Calls generate_payment_qr tool]
      QR code generated! Customers can scan it to pay KES 500.
```

---

### 3. LangChain Integration (`langchain_agent.py`)

Coming soon! Integration with LangChain for building payment-enabled chatbots.

---

### 4. Claude Desktop Integration

Add to your Claude Desktop config:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

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
You: Charge customer 254712345678 KES 1000
Claude: [Uses stk_push tool automatically]
```

---

## Testing with cURL

### 1. Connect to SSE
```bash
curl -N http://localhost:8080/sse
```

Output:
```
event: endpoint
data: /message?session=session_1732543210123

event: ping
data: 
```

### 2. Send STK Push Request
```bash
SESSION_ID="session_1732543210123"  # Use actual session from above

curl -X POST "http://localhost:8080/message?session=$SESSION_ID" \
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

### 3. Generate QR Code
```bash
curl -X POST "http://localhost:8080/message?session=$SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "generate_qr_code",
      "arguments": {
        "merchant_name": "Test Store",
        "ref_no": "ORDER123",
        "amount": 500,
        "trx_code": "BG",
        "cp_identifier": "123456"
      }
    }
  }'
```

---

## Use Cases

### E-commerce Checkout
```python
# Customer completes order
order_total = calculate_cart_total()
customer_phone = get_customer_phone()

# Request payment
mpesa.charge_customer(customer_phone, order_total, f"Order #{order_id}")
```

### Restaurant POS
```python
# Generate QR for table payment
table_bill = calculate_table_bill(table_number)
qr_code = mpesa.generate_payment_qr(
    merchant_name="My Restaurant",
    amount=table_bill,
    till_number="123456",
    reference=f"TABLE_{table_number}"
)
display_qr_code(qr_code)
```

### Subscription Billing
```python
# Monthly subscription charge
for subscriber in get_due_subscriptions():
    mpesa.charge_customer(
        subscriber.phone,
        subscriber.plan_amount,
        f"Subscription - {subscriber.plan_name}"
    )
```

---

## Troubleshooting

### Server Not Running
```
Error: Connection refused
```
**Solution:** Start the MCP server first:
```bash
cd .. && go run cmd/server/main.go
```

### Invalid Credentials
```
Error: Invalid Access Token
```
**Solution:** Check your `.env` file has correct Daraja credentials

### Phone Number Format
```
Error: Invalid phone number
```
**Solution:** Use format `254XXXXXXXXX` or `0XXXXXXXXX` (Kenyan numbers only)

### Callback Not Received
For production, ensure:
- Callback URL is publicly accessible (HTTPS)
- Firewall allows incoming requests
- Endpoint validates M-Pesa callback signature

---

## Next Steps

1. **Add Error Handling:** Implement retry logic and better error messages
2. **Add Logging:** Track all payment requests and responses
3. **Add Database:** Store payment records for reconciliation
4. **Add Webhooks:** Implement callback handler for payment confirmations
5. **Add Tests:** Write unit and integration tests
6. **Add Monitoring:** Set up alerts for failed payments

---

## Resources

- [M-Pesa Daraja API Docs](https://developer.safaricom.co.ke/Documentation)
- [MCP Protocol Spec](https://spec.modelcontextprotocol.io/)
- [Google ADK Docs](https://github.com/google/adk)
- [LangChain Docs](https://python.langchain.com/)

---

## Contributing

Have a cool integration example? Submit a PR!

1. Fork the repository
2. Create your example file
3. Add documentation
4. Test thoroughly
5. Submit pull request
