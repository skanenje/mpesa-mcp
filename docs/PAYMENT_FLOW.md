# M-Pesa Payment Flow Explained

## Simple Overview

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  CUSTOMER   │  Pays   │   M-PESA    │ Credits │YOUR BUSINESS│
│             │ ──────> │   SYSTEM    │ ──────> │  ACCOUNT    │
│ 254712...   │ KES 100 │             │ KES 100 │  (174379)   │
└─────────────┘         └─────────────┘         └─────────────┘
```

**Key Point:** Money flows FROM customer TO your business. Always.

---

## Detailed Flow

### 1. Customer Places Order

```
Customer: "I want to buy a shirt for KES 1500"
         ↓
Your Website/App: Creates order #12345
         ↓
AI Agent: "I need to collect payment"
```

### 2. Payment Request Initiated

```
AI Agent → MCP Server → M-Pesa API
         ↓
Request: {
  FROM: 254712345678 (customer)
  TO: 174379 (your business)
  AMOUNT: KES 1500
  REFERENCE: "Order #12345"
}
```

### 3. Customer Receives Prompt

```
Customer's Phone:
┌─────────────────────────────┐
│ M-PESA                      │
│                             │
│ Pay KES 1,500.00           │
│ to YourBusinessName         │
│ for Order #12345            │
│                             │
│ Enter PIN: [____]           │
│                             │
│ [Cancel]        [OK]        │
└─────────────────────────────┘
```

### 4. Customer Authorizes Payment

```
Customer enters M-Pesa PIN
         ↓
M-Pesa verifies:
  ✓ PIN is correct
  ✓ Customer has sufficient balance
  ✓ Transaction is valid
         ↓
M-Pesa processes payment
```

### 5. Money Transfer

```
Customer's M-Pesa Account:
  Balance: KES 5,000
  - KES 1,500 (payment)
  = KES 3,500 (new balance)

Your Business M-Pesa Account:
  Balance: KES 10,000
  + KES 1,500 (received)
  = KES 11,500 (new balance)
```

### 6. Confirmation Sent

```
M-Pesa → Your Callback URL
         ↓
POST https://yourdomain.com/callback
{
  "ResultCode": 0,
  "TransactionID": "OEI2AK4Q16",
  "Amount": 1500,
  "PhoneNumber": 254712345678
}
         ↓
Your Server:
  ✓ Updates order status to "paid"
  ✓ Sends confirmation email
  ✓ Triggers order fulfillment
```

### 7. Customer Receives Receipt

```
Customer's Phone (SMS):
┌─────────────────────────────┐
│ OEI2AK4Q16 Confirmed.       │
│ You have paid KES 1,500.00  │
│ to YourBusinessName         │
│ on 25/11/24 at 2:30 PM      │
│ New M-PESA balance is       │
│ KES 3,500.00                │
└─────────────────────────────┘
```

---

## Who Controls What?

### You Control:
- ✅ Your business shortcode (where money goes)
- ✅ When to request payment
- ✅ How much to charge
- ✅ What the payment is for
- ✅ Your callback URL
- ✅ Order fulfillment

### Customer Controls:
- ✅ Whether to approve or cancel
- ✅ Their M-Pesa PIN (you never see this)
- ✅ Their M-Pesa balance

### M-Pesa Controls:
- ✅ Payment processing
- ✅ Security and fraud detection
- ✅ Transaction limits
- ✅ When callbacks are sent

### You DON'T Control:
- ❌ Customer's M-Pesa account
- ❌ Customer's PIN
- ❌ M-Pesa's processing time
- ❌ Whether customer has sufficient balance

---

## Common Scenarios

### Scenario 1: E-commerce Purchase

```
1. Customer adds items to cart: KES 2,500
2. Customer clicks "Checkout"
3. Customer enters phone: 254722334455
4. AI Agent: "Charge 254722334455 KES 2500 for order #789"
5. MCP Server → M-Pesa: STK Push request
6. Customer receives prompt, enters PIN
7. Payment successful
8. Your system ships the product
```

**Money Flow:** Customer (254722334455) → Your Business

---

### Scenario 2: Restaurant Bill

```
1. Customer finishes meal
2. Waiter: "Table 5, bill is KES 3,200"
3. AI Agent generates QR code
4. Customer scans QR with M-Pesa app
5. Customer confirms payment
6. Payment successful
7. Receipt printed
```

**Money Flow:** Customer → Your Restaurant

---

### Scenario 3: Subscription Renewal

```
1. Subscription due date arrives
2. AI Agent: "Charge 254733445566 KES 999 for Premium Plan"
3. MCP Server → M-Pesa: STK Push request
4. Customer receives prompt
5. Customer enters PIN
6. Payment successful
7. Subscription renewed for another month
```

**Money Flow:** Subscriber → Your Business

---

### Scenario 4: Failed Payment

```
1. AI Agent requests payment
2. Customer receives prompt
3. Customer cancels OR has insufficient funds
4. M-Pesa sends failure callback
5. Your system:
   - Marks order as "payment failed"
   - Sends retry link to customer
   - Keeps order in cart for 24 hours
```

**Money Flow:** None (payment failed)

---

## Security & Trust

### How M-Pesa Protects Everyone:

**For Customers:**
- PIN required for every transaction
- Transaction limits (max KES 150,000 per transaction)
- SMS confirmation for every payment
- Can dispute fraudulent transactions
- Regulated by Central Bank of Kenya

**For Businesses:**
- KYC verification required
- Business must be registered
- Transaction records maintained
- Chargeback protection
- Fraud detection systems

**For You (Developer):**
- OAuth authentication required
- HTTPS only
- IP whitelisting available
- Callback verification
- Sandbox for testing

---

## What Can Go Wrong?

### Customer Side:
- ❌ Customer cancels payment → Order remains unpaid
- ❌ Insufficient balance → Payment fails
- ❌ Wrong PIN entered 3 times → Account locked temporarily
- ❌ Customer doesn't respond → Timeout after 60 seconds

### Your Side:
- ❌ Wrong business shortcode → Money goes to wrong account
- ❌ Callback URL down → You don't receive confirmation
- ❌ No callback handler → Can't track payment status
- ❌ Wrong credentials → Authentication fails

### M-Pesa Side:
- ❌ System downtime → Payment delayed
- ❌ Network issues → Timeout
- ❌ Fraud detection → Transaction blocked

---

## Best Practices

### 1. Always Track Requests
```go
// Store every payment request
db.SavePaymentRequest(PaymentRequest{
    CheckoutRequestID: response.CheckoutRequestID,
    PhoneNumber: phone,
    Amount: amount,
    Status: "pending",
    CreatedAt: time.Now(),
})
```

### 2. Implement Callback Handler
```go
// Receive payment confirmations
func handleCallback(callback *STKCallback) {
    if callback.ResultCode == 0 {
        // Payment successful
        db.UpdateOrderStatus(callback.CheckoutRequestID, "paid")
        fulfillOrder(callback.CheckoutRequestID)
    } else {
        // Payment failed
        db.UpdateOrderStatus(callback.CheckoutRequestID, "failed")
        notifyCustomer(callback.CheckoutRequestID)
    }
}
```

### 3. Reconcile Daily
```sql
-- Check for missing callbacks
SELECT * FROM payment_requests
WHERE status = 'pending'
AND created_at < NOW() - INTERVAL '1 hour';
```

### 4. Handle Failures Gracefully
```go
if paymentFailed {
    // Don't just give up!
    sendRetryLink(customer)
    keepOrderInCart(24 * time.Hour)
    notifySupport(order)
}
```

### 5. Test Everything
```bash
# Test in sandbox first
BUSINESS_SHORTCODE=174379  # Sandbox shortcode
BASE_URL=https://sandbox.safaricom.co.ke

# Then move to production
BUSINESS_SHORTCODE=your_real_shortcode
BASE_URL=https://api.safaricom.co.ke
```

---

## Summary

**Remember:**
1. Money flows FROM customer TO your business
2. Customer must approve every payment
3. You need a callback handler to know if payment succeeded
4. Always test in sandbox first
5. Keep transaction records
6. Handle failures gracefully
7. Comply with regulations

**Questions?**
- Read the [full README](../README.md)
- Check [Security Best Practices](../README.md#-security-best-practices)
- See [Callback Implementation Guide](../CALLBACK_IMPLEMENTATION.md)
- Review [Troubleshooting](../README.md#-troubleshooting)

---

**Ready to implement?** Start with the [Quick Start Guide](../QUICKSTART.md)!
