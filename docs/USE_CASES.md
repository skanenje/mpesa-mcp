# Use Cases & Examples

Practical use cases and agent examples for the M-Pesa MCP Server.

## Understanding the Payment Flow

### Who Pays Who?

```
CUSTOMER (PartyA)  →  Pays Money  →  YOUR BUSINESS (PartyB)
   254712345678                         Your Shortcode/Till
```

**The money flow is ALWAYS:**
- **FROM**: Customer's M-Pesa account (the phone number you specify)
- **TO**: YOUR business account (configured in `.env` as `BUSINESS_SHORTCODE`)

**All payments go TO your business account, not to the agent or server operator.**

## Use Cases & Goals

### 1. E-commerce / Online Stores

**Goal**: Collect payment from customers for products/services

**Flow**: Customer orders → Agent requests payment → Customer pays YOUR business

**Example**:
```python
agent_prompt = """
You are an e-commerce checkout assistant. When a customer confirms their order:
1. Calculate the total amount
2. Use stk_push to request payment from their phone number
3. Inform them to check their phone and enter M-Pesa PIN
4. Confirm when payment is initiated
"""

# Example conversation:
# User: "I want to buy 2 shirts at KES 1500 each"
# Agent: "That's KES 3000 total. What's your M-Pesa number?"
# User: "0712345678"
# Agent: [calls stk_push] "Payment request sent! Check your phone to complete."
```

### 2. Service Providers (Salons, Clinics, etc.)

**Goal**: Collect payment for services rendered

**Flow**: Service completed → Agent requests payment → Customer pays YOUR business

**Example**: "Charge customer 254722334455 KES 2500 for haircut service"

### 3. Subscription Services

**Goal**: Collect recurring payments from subscribers

**Flow**: Subscription due → Agent requests payment → Customer pays YOUR business

**Example**:
```python
agent_prompt = """
You manage subscription payments. When a subscription is due:
1. Check customer's phone number
2. Send payment request via stk_push
3. Log the transaction
4. Send confirmation email
"""
```

### 4. Restaurants / Retail POS

**Goal**: Collect payment at point of sale

**Flow**: Bill ready → Generate QR or STK Push → Customer pays YOUR business

**Example**:
```python
agent_prompt = """
You are a point-of-sale assistant. When a customer is ready to pay:
1. Calculate the bill total
2. Generate a QR code using generate_qr_code
3. Display the QR code for customer to scan
"""

# Example:
# Staff: "Customer bill is KES 2500"
# Agent: [generates QR code] "QR code ready! Customer can scan to pay."
```

### 5. Utility Bills / Invoices

**Goal**: Collect payment for bills and invoices

**Flow**: Invoice sent → Agent requests payment → Customer pays YOUR business

**Example**: "Send payment request to 254744556677 for invoice #INV-2024-001"

## What This Server Does NOT Do

❌ **Does NOT**: Transfer money between arbitrary accounts  
❌ **Does NOT**: Send money from your business to customers (that's B2C, different API)  
❌ **Does NOT**: Handle peer-to-peer transfers  
❌ **Does NOT**: Store or hold money (it just initiates payment requests)

## Example Payment Flow

```
1. Customer orders product on your website
   └─> Order #12345, Amount: KES 1500, Customer: 254712345678

2. Your AI agent calls stk_push tool
   └─> Agent: "Charge 254712345678 KES 1500 for order #12345"

3. MCP Server sends request to M-Pesa
   └─> FROM: 254712345678 (customer)
   └─> TO: 174379 (YOUR business shortcode)
   └─> AMOUNT: KES 1500

4. Customer receives prompt on phone
   └─> "Pay KES 1500 to YourBusinessName for order #12345"
   └─> Customer enters M-Pesa PIN

5. M-Pesa processes payment
   └─> Deducts KES 1500 from customer's M-Pesa account
   └─> Credits KES 1500 to YOUR business M-Pesa account

6. M-Pesa sends confirmation to your callback URL
   └─> POST https://yourdomain.com/callback
   └─> { "ResultCode": 0, "TransactionID": "OEI2AK4Q16" }

7. Your system confirms order and delivers product
```

## Customer Support Agent Example

```python
agent_prompt = """
You help customers with M-Pesa payments. You can:
- Check token status if payments are failing
- Resend payment requests
- Generate new QR codes if needed
- Explain payment steps to customers
"""
```
