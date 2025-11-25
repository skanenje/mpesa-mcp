# Troubleshooting Guide

Common issues and solutions for the M-Pesa MCP Server.

## Common Issues

### "Invalid Access Token"

**Symptoms**: API calls fail with authentication errors

**Solutions**:
- Token might have expired
- Check `get_token_status` tool to verify token status
- Verify Consumer Key/Secret are correct in `.env`
- Ensure BASE_URL is correct (sandbox vs production)
- Check if credentials are for the correct environment

### "Invalid Phone Number"

**Symptoms**: STK Push fails with phone number validation error

**Solutions**:
- Must be Kenyan number starting with `254`
- Remove spaces and special characters
- Server auto-formats `07XX` to `2547XX`
- Valid format: `254712345678` (12 digits total)
- Invalid formats: `+254712345678`, `0712345678`, `712345678`

### "Callback Not Received"

**Symptoms**: Payment initiated but no confirmation received

**Solutions**:
- For local testing, use ngrok: `ngrok http 8080`
- Update CALLBACK_URL in `.env` with ngrok URL
- Check firewall/network settings
- Verify callback endpoint is publicly accessible
- Check M-Pesa IP whitelist (production only)
- Ensure callback endpoint returns 200 OK

### "Build Errors"

**Symptoms**: Go compilation fails

**Solutions**:
- Ensure Go 1.22+ installed: `go version`
- Run `go mod tidy` to sync dependencies
- Check import paths in code
- Delete `go.sum` and run `go mod download`
- Verify all files are in correct directories

### "Connection Refused"

**Symptoms**: Cannot connect to MCP server

**Solutions**:
- Verify server is running: `ps aux | grep mpesa-mcp`
- Check if port 8080 is available: `lsof -i :8080`
- Check firewall rules
- Verify server started without errors in logs

### "Rate Limit Exceeded"

**Symptoms**: Too many requests error from Daraja API

**Solutions**:
- Implement exponential backoff
- Cache OAuth tokens (server does this automatically)
- Reduce request frequency
- Contact Safaricom for rate limit increase (production)

### "Transaction Timeout"

**Symptoms**: Customer doesn't complete payment in time

**Solutions**:
- Default timeout is 60 seconds
- Customer must enter PIN within this time
- Implement retry mechanism
- Send reminder to customer
- Check if customer's phone is on and has network

## Testing Issues

### Cannot Test with Real Phone

**Solution**: Use Safaricom test credentials and test phone numbers provided in Daraja portal

### Sandbox Returns Errors

**Solutions**:
- Verify using sandbox URL: `https://sandbox.safaricom.co.ke`
- Use test shortcode: `174379`
- Use test passkey from Daraja portal
- Check test phone numbers are valid

## Debugging Tips

### Enable Verbose Logging

```go
// Add to your code
log.SetFlags(log.LstdFlags | log.Lshortfile)
log.Println("Debug info here")
```

### Check Token Status

```bash
curl -X POST "http://localhost:8080/message?session=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get_token_status",
      "arguments": {}
    }
  }'
```

### Test Connectivity

```bash
# Test Daraja API connectivity
curl https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials \
  -u "CONSUMER_KEY:CONSUMER_SECRET"
```

### Monitor Logs

```bash
# Follow server logs
tail -f /var/log/mpesa-mcp.log

# Or if running in terminal
./mpesa-mcp 2>&1 | tee mpesa-mcp.log
```

## Getting Help

If you're still stuck:

1. Check [Safaricom Daraja Documentation](https://developer.safaricom.co.ke/Documentation)
2. Review [examples/](../examples/) directory
3. Open an issue on GitHub with:
   - Error message (remove sensitive data)
   - Steps to reproduce
   - Environment details (Go version, OS, etc.)
   - Relevant logs
