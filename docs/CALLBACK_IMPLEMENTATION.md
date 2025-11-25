# Implementing M-Pesa Callback Handler

This guide explains how to implement a secure callback handler to receive payment confirmations from M-Pesa.

## Why You Need a Callback Handler

When you initiate an STK Push:
1. Customer receives prompt on their phone
2. Customer enters PIN (or cancels)
3. M-Pesa processes the payment
4. **M-Pesa sends result to your callback URL** ← You need to handle this!

Without a callback handler, you won't know if the payment succeeded or failed.

## Callback Flow

```
1. You initiate STK Push
   └─> M-Pesa sends prompt to customer

2. Customer enters PIN
   └─> M-Pesa processes payment

3. M-Pesa sends callback to your server
   └─> POST https://yourdomain.com/callback
   └─> Contains: success/failure, transaction ID, amount, etc.

4. Your server processes callback
   └─> Update order status
   └─> Fulfill order (if successful)
   └─> Notify customer
```

## Callback Data Structure

M-Pesa sends this JSON to your callback URL:

```json
{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "29115-34620561-1",
      "CheckoutRequestID": "ws_CO_191220191020363925",
      "ResultCode": 0,
      "ResultDesc": "The service request is processed successfully.",
      "CallbackMetadata": {
        "Item": [
          {
            "Name": "Amount",
            "Value": 1000
          },
          {
            "Name": "MpesaReceiptNumber",
            "Value": "OEI2AK4Q16"
          },
          {
            "Name": "TransactionDate",
            "Value": 20191219102115
          },
          {
            "Name": "PhoneNumber",
            "Value": 254712345678
          }
        ]
      }
    }
  }
}
```

**Result Codes:**
- `0` = Success
- `1032` = Cancelled by user
- `1037` = Timeout (user didn't respond)
- `1` = Insufficient funds
- `2001` = Wrong PIN

## Implementation Example

### Step 1: Add Callback Types

```go
// internal/mpesa/types.go

type CallbackRequest struct {
    Body struct {
        STKCallback STKCallback `json:"stkCallback"`
    } `json:"Body"`
}

type STKCallback struct {
    MerchantRequestID string           `json:"MerchantRequestID"`
    CheckoutRequestID string           `json:"CheckoutRequestID"`
    ResultCode        int              `json:"ResultCode"`
    ResultDesc        string           `json:"ResultDesc"`
    CallbackMetadata  CallbackMetadata `json:"CallbackMetadata"`
}

type CallbackMetadata struct {
    Item []CallbackItem `json:"Item"`
}

type CallbackItem struct {
    Name  string      `json:"Name"`
    Value interface{} `json:"Value"`
}

// Helper to extract values from callback metadata
func (c *STKCallback) GetAmount() float64 {
    for _, item := range c.CallbackMetadata.Item {
        if item.Name == "Amount" {
            if v, ok := item.Value.(float64); ok {
                return v
            }
        }
    }
    return 0
}

func (c *STKCallback) GetTransactionID() string {
    for _, item := range c.CallbackMetadata.Item {
        if item.Name == "MpesaReceiptNumber" {
            if v, ok := item.Value.(string); ok {
                return v
            }
        }
    }
    return ""
}

func (c *STKCallback) GetPhoneNumber() string {
    for _, item := range c.CallbackMetadata.Item {
        if item.Name == "PhoneNumber" {
            if v, ok := item.Value.(float64); ok {
                return fmt.Sprintf("%.0f", v)
            }
        }
    }
    return ""
}
```

### Step 2: Create Callback Handler

```go
// internal/mpesa/callback.go

package mpesa

import (
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "strings"
)

// M-Pesa callback IP addresses (whitelist these)
var mpesaIPs = []string{
    "196.201.214.200",
    "196.201.214.206",
    "196.201.213.114",
    "196.201.214.207",
    "196.201.214.208",
}

// CallbackHandler handles M-Pesa payment callbacks
type CallbackHandler struct {
    onSuccess func(callback *STKCallback) error
    onFailure func(callback *STKCallback) error
}

// NewCallbackHandler creates a new callback handler
func NewCallbackHandler(
    onSuccess func(*STKCallback) error,
    onFailure func(*STKCallback) error,
) *CallbackHandler {
    return &CallbackHandler{
        onSuccess: onSuccess,
        onFailure: onFailure,
    }
}

// HandleCallback processes M-Pesa callbacks
func (h *CallbackHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    // 1. Verify request method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 2. Verify IP address (optional but recommended)
    if !h.isAllowedIP(r.RemoteAddr) {
        log.Printf("Callback from unauthorized IP: %s", r.RemoteAddr)
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // 3. Parse callback data
    var callbackReq CallbackRequest
    if err := json.NewDecoder(r.Body).Decode(&callbackReq); err != nil {
        log.Printf("Failed to parse callback: %v", err)
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    callback := &callbackReq.Body.STKCallback

    // 4. Log callback (without sensitive data)
    log.Printf("Callback received: CheckoutRequestID=%s, ResultCode=%d",
        callback.CheckoutRequestID,
        callback.ResultCode,
    )

    // 5. Process based on result code
    if callback.ResultCode == 0 {
        // Success
        if h.onSuccess != nil {
            if err := h.onSuccess(callback); err != nil {
                log.Printf("Error processing successful callback: %v", err)
                http.Error(w, "Internal error", http.StatusInternalServerError)
                return
            }
        }
    } else {
        // Failure
        if h.onFailure != nil {
            if err := h.onFailure(callback); err != nil {
                log.Printf("Error processing failed callback: %v", err)
                http.Error(w, "Internal error", http.StatusInternalServerError)
                return
            }
        }
    }

    // 6. Respond to M-Pesa
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "ResultCode": "0",
        "ResultDesc": "Accepted",
    })
}

// isAllowedIP checks if the request is from M-Pesa servers
func (h *CallbackHandler) isAllowedIP(remoteAddr string) bool {
    // Extract IP from address (remove port)
    ip, _, err := net.SplitHostPort(remoteAddr)
    if err != nil {
        ip = remoteAddr
    }

    // Check against whitelist
    for _, allowedIP := range mpesaIPs {
        if ip == allowedIP {
            return true
        }
    }

    // In development, allow localhost
    if strings.HasPrefix(ip, "127.0.0.1") || strings.HasPrefix(ip, "::1") {
        return true
    }

    return false
}
```

### Step 3: Integrate with Main Server

```go
// cmd/server/main.go

func main() {
    // ... existing setup ...

    // Create callback handler
    callbackHandler := mpesa.NewCallbackHandler(
        // On success
        func(callback *mpesa.STKCallback) error {
            log.Printf("Payment successful!")
            log.Printf("  Transaction ID: %s", callback.GetTransactionID())
            log.Printf("  Amount: %.2f", callback.GetAmount())
            log.Printf("  Phone: %s", callback.GetPhoneNumber())
            log.Printf("  Checkout Request ID: %s", callback.CheckoutRequestID)

            // TODO: Update your database
            // db.UpdateOrderStatus(callback.CheckoutRequestID, "paid")
            
            // TODO: Fulfill order
            // fulfillOrder(callback.CheckoutRequestID)
            
            // TODO: Send confirmation to customer
            // sendSMS(callback.GetPhoneNumber(), "Payment received!")

            return nil
        },
        // On failure
        func(callback *mpesa.STKCallback) error {
            log.Printf("Payment failed!")
            log.Printf("  Result Code: %d", callback.ResultCode)
            log.Printf("  Result Desc: %s", callback.ResultDesc)
            log.Printf("  Checkout Request ID: %s", callback.CheckoutRequestID)

            // TODO: Update your database
            // db.UpdateOrderStatus(callback.CheckoutRequestID, "failed")
            
            // TODO: Notify customer
            // sendSMS(phone, "Payment failed. Please try again.")

            return nil
        },
    )

    // Add callback route
    mux.HandleFunc("/mpesa/callback", callbackHandler.HandleCallback)

    // ... rest of server setup ...
}
```

### Step 4: Update Environment Variables

```bash
# .env
CALLBACK_URL=https://yourdomain.com/mpesa/callback
```

## Testing Callbacks Locally

### Option 1: ngrok (Recommended)

```bash
# 1. Install ngrok
# Download from https://ngrok.com/

# 2. Start your server
go run cmd/server/main.go

# 3. Start ngrok
ngrok http 8080

# 4. Copy the HTTPS URL
# Example: https://abc123.ngrok.io

# 5. Update .env
CALLBACK_URL=https://abc123.ngrok.io/mpesa/callback

# 6. Restart server and test
```

### Option 2: webhook.site

```bash
# 1. Go to https://webhook.site/
# 2. Copy your unique URL
# 3. Update .env
CALLBACK_URL=https://webhook.site/your-unique-id

# 4. Test payment
# 5. View callback data on webhook.site
```

### Option 3: Local Testing Script

```bash
# Test callback locally
curl -X POST http://localhost:8080/mpesa/callback \
  -H "Content-Type: application/json" \
  -d '{
    "Body": {
      "stkCallback": {
        "MerchantRequestID": "29115-34620561-1",
        "CheckoutRequestID": "ws_CO_191220191020363925",
        "ResultCode": 0,
        "ResultDesc": "The service request is processed successfully.",
        "CallbackMetadata": {
          "Item": [
            {"Name": "Amount", "Value": 1000},
            {"Name": "MpesaReceiptNumber", "Value": "TEST123"},
            {"Name": "TransactionDate", "Value": 20191219102115},
            {"Name": "PhoneNumber", "Value": 254712345678}
          ]
        }
      }
    }
  }'
```

## Database Schema Example

```sql
CREATE TABLE payment_requests (
    id SERIAL PRIMARY KEY,
    checkout_request_id VARCHAR(255) UNIQUE NOT NULL,
    merchant_request_id VARCHAR(255),
    phone_number VARCHAR(20) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    account_reference VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending',
    transaction_id VARCHAR(255),
    result_code INT,
    result_desc TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_checkout_request_id (checkout_request_id),
    INDEX idx_phone_number (phone_number),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

## Production Considerations

### 1. Idempotency
M-Pesa may send the same callback multiple times. Handle this:

```go
func (h *CallbackHandler) onSuccess(callback *STKCallback) error {
    // Check if already processed
    exists, err := db.CheckTransactionExists(callback.CheckoutRequestID)
    if err != nil {
        return err
    }
    if exists {
        log.Printf("Callback already processed: %s", callback.CheckoutRequestID)
        return nil // Already handled
    }
    
    // Process transaction...
}
```

### 2. Async Processing
Don't block the callback response:

```go
func (h *CallbackHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    // Parse callback
    var callback CallbackRequest
    json.NewDecoder(r.Body).Decode(&callback)
    
    // Respond immediately
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"ResultCode": "0"})
    
    // Process asynchronously
    go h.processCallback(&callback.Body.STKCallback)
}
```

### 3. Retry Logic
If processing fails, retry:

```go
func (h *CallbackHandler) processCallback(callback *STKCallback) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := h.onSuccess(callback)
        if err == nil {
            return // Success
        }
        log.Printf("Retry %d/%d failed: %v", i+1, maxRetries, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    log.Printf("Failed to process callback after %d retries", maxRetries)
}
```

### 4. Monitoring
Track callback metrics:

```go
// Prometheus metrics example
var (
    callbacksReceived = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mpesa_callbacks_total",
            Help: "Total number of M-Pesa callbacks received",
        },
        []string{"result_code"},
    )
)

func (h *CallbackHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    // ... parse callback ...
    
    callbacksReceived.WithLabelValues(
        fmt.Sprintf("%d", callback.ResultCode),
    ).Inc()
    
    // ... process callback ...
}
```

## Troubleshooting

### Callback Not Received
1. Check callback URL is publicly accessible (HTTPS)
2. Verify firewall allows incoming requests
3. Check M-Pesa IP whitelist
4. Test with ngrok or webhook.site
5. Check server logs for errors

### Duplicate Callbacks
- Implement idempotency checks
- Use `CheckoutRequestID` as unique identifier
- Store processed callbacks in database

### Callback Timeout
- Respond to M-Pesa within 30 seconds
- Process heavy operations asynchronously
- Use message queues for complex workflows

---

**Next Steps:**
1. Implement the callback handler
2. Test with sandbox
3. Set up monitoring
4. Deploy to production
5. Monitor and optimize
