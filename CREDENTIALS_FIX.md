# Quick Fix: "Wrong credentials" Error

## The Problem

Getting error: `500.001.1001 - Wrong credentials` when calling STK Push

## The Solution

Your `.env` file had a **base64-encoded passkey** when it should be the **raw passkey**.

### What Changed

**Before (❌ Wrong):**
```env
PASSKEY=jpDNpPOYza8lcQTJD5IlndJR4dyPWMbkeX8TdE6CZXZirqpqWwpJ73jrdBnsiVEw+Oi6XTXIAx/GAAr7zhWEjxr2h43MPDq7rIjOYpADjjKZrnNzqF1HV9tE+VS4yrvPstRXmygbUQFR67D6jI4wCo5/GiC9I544HggB/guxbAybanBst/rCVWJOnD2vTmH61/eE7wZAtz3nFVJzJTUgGZ8WaYzQoVicYwLcUWGgDAQvvkpvkLZ9mp0O+RZY8ZLmafuKH3SWaqe6DmLphpYP1KIxQx71mn2tLf5GEyH9/nTT0BPkyhK6l4ZbNggLN6skyw3179RS7YYW1M5s9M7tKQ==
```

**After (✅ Correct):**
```env
PASSKEY=bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919
```

## Why This Matters

M-Pesa requires the password to be: `Base64(Shortcode + Passkey + Timestamp)`

Your code was doing: `Base64(Shortcode + AlreadyBase64EncodedPasskey + Timestamp)`

This created a **double-encoded password** that M-Pesa couldn't validate.

## Testing the Fix

```bash
# Rebuild
go build -o mpesa-mcp cmd/server/main.go

# Run server
./mpesa-mcp

# Test (in another terminal)
./scripts/test-stk-fixed.sh
```

## Expected Result

You should now see:
```
✅ Success. Request accepted for processing
MerchantRequestID: xxx-xxx-xxx
CheckoutRequestID: ws_CO_xxx
```

Instead of:
```
❌ 500.001.1001 - Wrong credentials
```

## Additional Changes

Also added support for optional parameters:
- `account_reference` - Shows on customer's phone (max 12 chars)
- `transaction_desc` - Transaction description (max 13 chars)

Example:
```json
{
  "name": "stk_push",
  "arguments": {
    "amount": 100,
    "phone_number": "254708374149",
    "account_reference": "ORDER123",
    "transaction_desc": "Payment"
  }
}
```

## Documentation

For more details, see:
- [Passkey Configuration Guide](docs/PASSKEY_GUIDE.md)
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
- [API Reference](docs/API_REFERENCE.md)

---

**Note**: This fix applies to both sandbox and production. For production, use the raw passkey sent to your email after Go Live.
