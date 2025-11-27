#!/bin/bash

# Debug STK Push - shows exactly what's being sent

source .env

echo "🔍 STK Push Debug Info"
echo "======================"
echo ""
echo "Configuration:"
echo "  Consumer Key: ${MPESA_CONSUMER_KEY:0:20}..."
echo "  Base URL: $BASE_URL"
echo "  Shortcode: $BUSINESS_SHORTCODE"
echo "  Passkey (first 50 chars): ${PASSKEY:0:50}..."
echo "  Callback URL: $CALLBACK_URL"
echo ""

# Test OAuth first
echo "1️⃣  Testing OAuth Token..."
TOKEN_RESPONSE=$(curl -s "$BASE_URL/oauth/v1/generate?grant_type=client_credentials" \
  -u "$MPESA_CONSUMER_KEY:$MPESA_CONSUMER_SECRET")

if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
    echo "✅ OAuth token obtained successfully"
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
    echo "   Token: ${ACCESS_TOKEN:0:30}..."
else
    echo "❌ OAuth failed:"
    echo "$TOKEN_RESPONSE"
    exit 1
fi

echo ""
echo "2️⃣  Testing STK Push..."

# Generate timestamp and password (same logic as Go code)
TIMESTAMP=$(date +"%Y%m%d%H%M%S")
PASSWORD_STR="${BUSINESS_SHORTCODE}${PASSKEY}${TIMESTAMP}"
PASSWORD=$(echo -n "$PASSWORD_STR" | base64)

echo "   Timestamp: $TIMESTAMP"
echo "   Password (first 50 chars): ${PASSWORD:0:50}..."
echo ""

# Make STK Push request
STK_RESPONSE=$(curl -s -X POST "$BASE_URL/mpesa/stkpush/v1/processrequest" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"BusinessShortCode\": \"$BUSINESS_SHORTCODE\",
    \"Password\": \"$PASSWORD\",
    \"Timestamp\": \"$TIMESTAMP\",
    \"TransactionType\": \"CustomerPayBillOnline\",
    \"Amount\": 1,
    \"PartyA\": \"254708374149\",
    \"PartyB\": \"$BUSINESS_SHORTCODE\",
    \"PhoneNumber\": \"254708374149\",
    \"CallBackURL\": \"$CALLBACK_URL\",
    \"AccountReference\": \"Test\",
    \"TransactionDesc\": \"Test Payment\"
  }")

echo "📋 STK Push Response:"
echo "$STK_RESPONSE" | jq '.' 2>/dev/null || echo "$STK_RESPONSE"
echo ""

if echo "$STK_RESPONSE" | grep -q "CheckoutRequestID"; then
    echo "✅ STK Push successful!"
elif echo "$STK_RESPONSE" | grep -q "500.001.1001"; then
    echo "❌ Wrong credentials error"
    echo ""
    echo "💡 This means:"
    echo "   - Your OAuth token is valid (we got past that)"
    echo "   - But your PASSKEY or BUSINESS_SHORTCODE is wrong for STK Push"
    echo ""
    echo "🔧 To fix:"
    echo "   1. Go to https://developer.safaricom.co.ke/"
    echo "   2. Select your app"
    echo "   3. Go to APIs tab"
    echo "   4. Click 'Lipa Na M-Pesa Online'"
    echo "   5. Get the 'Lipa Na M-Pesa Online Passkey' from Test Credentials"
    echo "   6. Update PASSKEY in your .env file"
else
    echo "⚠️  Other error - check response above"
fi
