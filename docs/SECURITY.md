# Security Best Practices

Security guidelines for production deployment of the M-Pesa MCP Server.

## Production Deployment Security

### 1. Credential Management

```bash
# ❌ NEVER do this
MPESA_CONSUMER_KEY=abc123  # Hardcoded in code

# ✅ DO this
# Use environment variables
export MPESA_CONSUMER_KEY=$(vault read secret/mpesa/key)

# Or use secret management services
# - AWS Secrets Manager
# - Google Secret Manager
# - HashiCorp Vault
# - Azure Key Vault
```

### 2. Callback Verification

When M-Pesa sends payment confirmations to your callback URL, verify them:

```go
// Example callback handler (you need to implement this)
func handleCallback(w http.ResponseWriter, r *http.Request) {
    // 1. Verify request is from M-Pesa (check IP whitelist)
    allowedIPs := []string{"196.201.214.200", "196.201.214.206"}
    if !isAllowedIP(r.RemoteAddr, allowedIPs) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // 2. Parse callback data
    var callback CallbackData
    if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 3. Verify transaction in your database
    order := getOrderByCheckoutRequestID(callback.CheckoutRequestID)
    if order == nil {
        log.Printf("Unknown transaction: %s", callback.CheckoutRequestID)
        return
    }
    
    // 4. Update order status
    if callback.ResultCode == 0 {
        order.Status = "paid"
        order.TransactionID = callback.TransactionID
        saveOrder(order)
        
        // 5. Fulfill order (send product, activate service, etc.)
        fulfillOrder(order)
    } else {
        order.Status = "failed"
        order.FailureReason = callback.ResultDesc
        saveOrder(order)
    }
    
    // 6. Respond to M-Pesa
    w.WriteHeader(http.StatusOK)
}
```

### 3. Input Validation

```go
// Always validate inputs before processing
func validatePaymentRequest(phone string, amount int) error {
    // Validate phone number format
    if !regexp.MustCompile(`^254[0-9]{9}$`).MatchString(phone) {
        return fmt.Errorf("invalid phone number format")
    }
    
    // Validate amount
    if amount < 1 || amount > 150000 {
        return fmt.Errorf("amount must be between 1 and 150,000 KES")
    }
    
    return nil
}
```

### 4. Rate Limiting

Prevent abuse by limiting requests:

```go
// Example using golang.org/x/time/rate
limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 req/sec

func paymentHandler(w http.ResponseWriter, r *http.Request) {
    if !limiter.Allow() {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    // Process payment...
}
```

### 5. Logging & Monitoring

```go
// Log all payment requests (but NOT sensitive data)
log.Printf("Payment request: OrderID=%s, Amount=%d, Phone=%s, Status=%s",
    order.ID,
    amount,
    maskPhoneNumber(phone), // 254712***678
    status,
)

// ❌ NEVER log
// - Full phone numbers (mask them)
// - Consumer keys/secrets
// - Access tokens
// - Customer PINs (you never have these anyway)
```

### 6. HTTPS Only

```go
// Redirect HTTP to HTTPS in production
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
    if r.TLS == nil {
        url := "https://" + r.Host + r.RequestURI
        http.Redirect(w, r, url, http.StatusMovedPermanently)
        return
    }
    // Continue with normal handler
}
```

### 7. Transaction Reconciliation

```sql
-- Daily reconciliation query
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_requests,
    SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
    SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END) as total_amount
FROM payment_requests
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

## Compliance Requirements

### For Production Use in Kenya:

1. **Business Registration**
   - Must have a registered business (sole proprietorship, limited company, etc.)
   - Business registration certificate required

2. **KYC (Know Your Customer)**
   - Submit business documents to Safaricom
   - Provide ID/passport of business owner
   - Proof of business address

3. **Data Protection**
   - Comply with Kenya Data Protection Act 2019
   - Have a privacy policy
   - Secure customer data
   - Allow customers to request data deletion

4. **Financial Regulations**
   - Keep transaction records for 7 years
   - Report suspicious transactions
   - Comply with anti-money laundering (AML) regulations

5. **Tax Compliance**
   - Register for KRA PIN
   - File tax returns
   - Issue receipts for all transactions

## Security Checklist

Before going live:

- [ ] All credentials stored in secure secret management system
- [ ] HTTPS enabled with valid SSL certificate
- [ ] Callback endpoint implemented and secured
- [ ] IP whitelisting for M-Pesa callbacks
- [ ] Input validation on all endpoints
- [ ] Rate limiting implemented
- [ ] Logging configured (without sensitive data)
- [ ] Monitoring and alerting set up
- [ ] Transaction reconciliation process in place
- [ ] Backup and disaster recovery plan
- [ ] Privacy policy published
- [ ] Terms of service published
- [ ] Refund policy documented
- [ ] Customer support process established
- [ ] Incident response plan documented
- [ ] Regular security audits scheduled

## Common Security Mistakes to Avoid

❌ **Storing credentials in code**
```go
// NEVER do this
const consumerKey = "abc123xyz"
```

❌ **Not verifying callbacks**
```go
// NEVER do this - accept any callback
func handleCallback(w http.ResponseWriter, r *http.Request) {
    // Process without verification
}
```

❌ **Logging sensitive data**
```go
// NEVER do this
log.Printf("Payment: phone=%s, token=%s", phone, token)
```

❌ **No transaction tracking**
```go
// NEVER do this - fire and forget
mpesa.STKPush(phone, amount)
// No record of what was sent
```

❌ **Trusting user input**
```go
// NEVER do this
amount := r.FormValue("amount") // Could be negative, huge, or non-numeric
mpesa.STKPush(phone, amount)
```

## Incident Response

If you suspect a security breach:

1. **Immediately**: Rotate all API credentials
2. **Investigate**: Check logs for suspicious activity
3. **Notify**: Inform affected customers if data was compromised
4. **Report**: Contact Safaricom if M-Pesa credentials compromised
5. **Document**: Record what happened and how you responded
6. **Improve**: Update security measures to prevent recurrence
