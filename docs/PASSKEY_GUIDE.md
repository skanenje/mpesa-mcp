# M-Pesa Passkey Configuration Guide

## Understanding the Passkey

The passkey is a critical component for M-Pesa STK Push authentication. Understanding how it works is essential to avoid the common "Wrong credentials" error.

## How Password Encoding Works

According to M-Pesa documentation, the Password parameter must be:

```
Base64.encode(BusinessShortCode + Passkey + Timestamp)
```

### Important: Passkey Format

The passkey in your `.env` file should be the **RAW passkey string**, NOT base64 encoded.

❌ **WRONG** - Using already base64-encoded passkey:
```env
PASSKEY=jpDNpPOYza8lcQTJD5IlndJR4dyPWMbkeX8TdE6CZXZirqpqWwpJ73jrdBnsiVEw+Oi6XTXIAx/GAAr7zhWEjxr2h43MPDq7rIjOYpADjjKZrnNzqF1HV9tE+VS4yrvPstRXmygbUQFR67D6jI4wCo5/GiC9I544HggB/guxbAybanBst/rCVWJOnD2vTmH61/eE7wZAtz3nFVJzJTUgGZ8WaYzQoVicYwLcUWGgDAQvvkpvkLZ9mp0O+RZY8ZLmafuKH3SWaqe6DmLphpYP1KIxQx71mn2tLf5GEyH9/nTT0BPkyhK6l4ZbNggLN6skyw3179RS7YYW1M5s9M7tKQ==
```

✅ **CORRECT** - Using raw passkey:
```env
PASSKEY=bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919
```

## Sandbox vs Production

### Sandbox Passkey

For sandbox testing, use the standard test passkey:
```
bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919
```

This is available on the Daraja Developer Portal simulator page.

### Production Passkey

For production:
1. Complete the Go Live process on Daraja Portal
2. The production passkey will be sent to your developer email
3. Use this raw passkey (not base64 encoded) in your production `.env` file

## Common Errors

### Error: 500.001.1001 - Wrong credentials

This error occurs when:

1. **Double encoding**: Passkey is already base64 encoded in `.env` file
2. **Mismatched shortcode**: BusinessShortCode in password encoding doesn't match request body
3. **Timestamp mismatch**: Timestamp used in encoding doesn't match request body
4. **Invalid passkey**: Using wrong passkey for your environment (sandbox vs production)

### How Our Code Handles It

```go
// Generate timestamp
timestamp := time.Now().Format("20060102150405")

// Generate password: Base64(ShortCode + Passkey + Timestamp)
passwordStr := c.config.BusinessCode + c.config.Passkey + timestamp
password := base64.StdEncoding.EncodeToString([]byte(passwordStr))
```

The code:
1. Reads raw passkey from config
2. Concatenates: ShortCode + RawPasskey + Timestamp
3. Base64 encodes the result
4. Uses this as the Password parameter

## Verification

To verify your passkey is correct:

1. Check the format - should be a hex string (sandbox) or alphanumeric (production)
2. Should NOT contain base64 padding characters (=)
3. For sandbox, should match: `bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919`

## Example Configuration

### Complete Sandbox Setup

```env
# M-Pesa Daraja API Credentials
MPESA_CONSUMER_KEY=your_consumer_key_here
MPESA_CONSUMER_SECRET=your_consumer_secret_here

# Sandbox environment
BASE_URL=https://sandbox.safaricom.co.ke

# Sandbox test credentials
BUSINESS_SHORTCODE=174379
PASSKEY=bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919

# Your callback URL
CALLBACK_URL=https://your-domain.com/mpesa/callback

# Optional customization
ACCOUNT_REFERENCE=PaymentRef
```

### Complete Production Setup

```env
# M-Pesa Daraja API Credentials
MPESA_CONSUMER_KEY=your_production_consumer_key
MPESA_CONSUMER_SECRET=your_production_consumer_secret

# Production environment
BASE_URL=https://api.safaricom.co.ke

# Your live credentials (from Go Live email)
BUSINESS_SHORTCODE=your_paybill_or_till
PASSKEY=your_production_passkey_from_email

# Your production callback URL
CALLBACK_URL=https://your-production-domain.com/mpesa/callback

# Optional customization
ACCOUNT_REFERENCE=YourBusiness
```

## Testing

After updating your passkey, test with:

```bash
./scripts/test-stk-fixed.sh
```

Or manually:

```bash
# Start server
./mpesa-mcp

# In another terminal, test STK Push
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "stk_push",
      "arguments": {
        "amount": 1,
        "phone_number": "254708374149"
      }
    }
  }'
```

## References

- [M-Pesa Express API Documentation](https://developer.safaricom.co.ke/APIs/MpesaExpressSimulate)
- [Daraja Developer Portal](https://developer.safaricom.co.ke/)
