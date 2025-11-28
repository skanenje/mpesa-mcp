#!/bin/bash

# ------------------------------------------------------------
# Direct M-PESA STK Push test using cURL (sandbox).
# This script follows the documentation provided by the user:
#   1) Generate an access token.
#   2) Build the password (Base64 of BusinessShortcode+Passkey+Timestamp).
#   3) POST the STK Push request.
# ------------------------------------------------------------

set -e

# Load environment variables (Consumer Key, Secret, Shortcode, Passkey, Callback URL)
source "$(dirname "$0")/../.env"

# Helper: Base64 encode without newline
b64() {
  echo -n "$1" | base64 | tr -d '\n'
}

# 1) Generate access token
# Encode credentials as Base64 for Basic Auth (remove any newlines)
AUTH_HEADER=$(echo -n "${MPESA_CONSUMER_KEY}:${MPESA_CONSUMER_SECRET}" | base64 | tr -d '\n')

echo "🔐 Requesting access token..."
TOKEN_RESPONSE=$(curl -s --http1.1 -X GET "${BASE_URL}/oauth/v1/generate?grant_type=client_credentials" \
  -H "Authorization: Basic ${AUTH_HEADER}")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "❌ Failed to obtain access token"
  echo "Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "✅ Access token obtained: ${ACCESS_TOKEN:0:20}..."

# 2) Build password
TIMESTAMP=$(date +"%Y%m%d%H%M%S")
PASSWORD=$(b64 "${BUSINESS_SHORTCODE}${PASSKEY}${TIMESTAMP}")

# 3) Build JSON payload using jq to avoid duplicate keys
PAYLOAD=$(jq -n \
  --arg bs "${BUSINESS_SHORTCODE}" \
  --arg pw "${PASSWORD}" \
  --arg ts "${TIMESTAMP}" \
  --argjson amt 1 \
  --arg pa "254708374149" \
  --arg pb "${BUSINESS_SHORTCODE}" \
  --arg pn "254708374149" \
  --arg cb "${CALLBACK_URL}" \
  --arg ar "TestPayment" \
  --arg td "Payment for goods" \
  '{BusinessShortCode:$bs,Password:$pw,Timestamp:$ts,TransactionType:"CustomerPayBillOnline",Amount:$amt,PartyA:$pa,PartyB:$pb,PhoneNumber:$pn,CallBackURL:$cb,AccountReference:$ar,TransactionDesc:$td}')

echo "🔎 JSON payload being sent:"
echo "$PAYLOAD"

# 4) Send STK Push request
STK_RESPONSE=$(curl -s --http1.1 -X POST "${BASE_URL}/mpesa/stkpush/v1/processrequest" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -d "$PAYLOAD")

# Show response (pretty‑print if jq is available)
if command -v jq > /dev/null 2>&1; then
  echo "📋 STK Push response:" && echo "$STK_RESPONSE" | jq '.'
else
  echo "📋 STK Push response:" && echo "$STK_RESPONSE"
fi

# Check for success
if echo "$STK_RESPONSE" | grep -q "CheckoutRequestID"; then
  echo "✅ STK Push initiated successfully"
else
  echo "⚠️  STK Push may have failed – see response above"
fi
