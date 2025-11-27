# M-Pesa Callback Implementation

This guide explains how the M-Pesa MCP server handles callbacks (webhooks) from Safaricom.

## Overview

When you initiate an STK Push, the transaction is asynchronous. M-Pesa processes the payment and then sends a notification (callback) to your server with the final status (Success or Failure).

## How it Works

1.  **Initiation**: You call `stk_push` with a `CallBackURL`.
2.  **Processing**: M-Pesa processes the request on the customer's phone.
3.  **Notification**: M-Pesa sends a POST request to your `CallBackURL` with the result.
4.  **Handling**: The server's `/callback` endpoint receives this JSON, parses it, and logs the result.

## JSON Payload Structure

M-Pesa sends a JSON payload with the following structure:

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
                        "Value": 1.00
                    },
                    {
                        "Name": "MpesaReceiptNumber",
                        "Value": "NLJ7RT61SV"
                    },
                    {
                        "Name": "TransactionDate",
                        "Value": 20191219102115
                    },
                    {
                        "Name": "PhoneNumber",
                        "Value": 254708374149
                    }
                ]
            }
        }
    }
}
```

### Key Fields

-   **ResultCode**: `0` means success. Any other number indicates failure (e.g., `1032` for cancelled).
-   **ResultDesc**: A text description of the result.
-   **CallbackMetadata**: Contains details like Amount, Receipt Number, and Date (only present on success).

## Local Development & Testing

Since M-Pesa cannot send requests to `localhost`, you must use a tunneling tool like **ngrok** to expose your local server.

### 1. Start ngrok

Expose port 8080:

```bash
ngrok http 8080
```

Copy the HTTPS URL (e.g., `https://1234-5678.ngrok-free.app`).

### 2. Configure Server

Update your `.env` file:

```env
CALLBACK_URL=https://1234-5678.ngrok-free.app/callback
```

### 3. Restart Server

Restart the MCP server to apply changes.

### 4. Verify

When you trigger an STK Push, watch your server logs. You should see:

```
Received M-Pesa Callback: { ... JSON payload ... }
✅ Transaction Successful! [MerchantID: ..., CheckoutID: ...]
   - Amount: 10
   - MpesaReceiptNumber: QWE123RTY
   ...
```
