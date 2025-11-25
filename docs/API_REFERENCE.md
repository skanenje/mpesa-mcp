# API Reference

Complete reference for all available MCP tools and their parameters.

## Available Tools

### 1. `stk_push`

Initiates an STK Push payment request (sends payment prompt to customer's phone).

**Parameters:**
- `amount` (int, required): Amount in KES
- `phone_number` (string, required): Phone number in format `254XXXXXXXXX` or `0XXXXXXXXX`

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
- `merchant_name` (string, required): Business name
- `ref_no` (string, required): Transaction reference
- `amount` (int, required): Amount in KES
- `trx_code` (string, required): Transaction type
  - `BG` - Buy Goods (Till Number)
  - `PB` - Paybill
  - `SM` - Send Money (Phone Number)
  - `SB` - Send to Business
  - `WA` - Withdraw Cash
- `cp_identifier` (string, required): Till number, paybill number, or phone number

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

**Parameters:** None

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

## Understanding Daraja API

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

## Data Flow Example

```
1. AI Agent Request
   ↓
   POST /message?session=xyz
   {
     "method": "tools/call",
     "params": {
       "name": "stk_push",
       "arguments": {"amount": 1000, "phone_number": "0712345678"}
     }
   }

2. MCP Server Processing
   ↓
   - Validate input
   - Format phone: 0712345678 → 254712345678
   - Generate password: Base64(ShortCode + Passkey + Timestamp)
   - Get OAuth token (cached, auto-refreshed)

3. Daraja API Call
   ↓
   POST https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest
   Authorization: Bearer <token>
   {
     "BusinessShortCode": "174379",
     "Password": "MTc0Mzc5YmZiMjc5...",
     "Timestamp": "20241125143022",
     "TransactionType": "CustomerPayBillOnline",
     "Amount": 1000,
     "PartyA": "254712345678",
     "PartyB": "174379",
     "PhoneNumber": "254712345678",
     "CallBackURL": "https://your-callback.com/mpesa",
     "AccountReference": "Order12345",
     "TransactionDesc": "Payment for goods/services"
   }

4. Customer's Phone
   ↓
   - Receives M-Pesa push notification
   - "Enter PIN to pay KES 1000 to Order12345"
   - Customer enters PIN

5. Response to Agent
   ↓
   SSE Event: {
     "MerchantRequestID": "29115-34620561-1",
     "CheckoutRequestID": "ws_CO_191220191020363925",
     "ResponseCode": "0",
     "CustomerMessage": "Success. Request accepted for processing"
   }

6. Callback (Async)
   ↓
   POST https://your-callback.com/mpesa
   {
     "ResultCode": 0,
     "ResultDesc": "The service request is processed successfully.",
     "TransactionID": "OEI2AK4Q16"
   }
```
