#!/bin/bash

# Test M-Pesa credentials by getting an access token

set -e

echo "🔐 Testing M-Pesa Credentials"
echo "=============================="
echo ""

# Load .env
if [ ! -f ".env" ]; then
    echo "❌ .env file not found"
    exit 1
fi

source .env

# Check if credentials are set
if [ -z "$MPESA_CONSUMER_KEY" ] || [ "$MPESA_CONSUMER_KEY" = "your_consumer_key_here" ]; then
    echo "❌ MPESA_CONSUMER_KEY not configured"
    exit 1
fi

if [ -z "$MPESA_CONSUMER_SECRET" ]; then
    echo "❌ MPESA_CONSUMER_SECRET not configured"
    exit 1
fi

echo "📋 Configuration:"
echo "   Consumer Key: ${MPESA_CONSUMER_KEY:0:20}..."
echo "   Base URL: $BASE_URL"
echo "   Shortcode: $BUSINESS_SHORTCODE"
echo ""

# Try to get access token
echo "🔑 Requesting access token..."
RESPONSE=$(curl -s "$BASE_URL/oauth/v1/generate?grant_type=client_credentials" \
  -u "$MPESA_CONSUMER_KEY:$MPESA_CONSUMER_SECRET")

echo "📋 Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
echo ""

# Check if we got a token
if echo "$RESPONSE" | grep -q "access_token"; then
    echo "✅ Credentials are VALID!"
    echo "   Your M-Pesa API credentials are working correctly."
elif echo "$RESPONSE" | grep -q "Invalid"; then
    echo "❌ Credentials are INVALID!"
    echo ""
    echo "📝 To fix this:"
    echo "   1. Go to https://developer.safaricom.co.ke/"
    echo "   2. Log in and go to 'My Apps'"
    echo "   3. Get fresh Consumer Key and Consumer Secret"
    echo "   4. Update your .env file"
else
    echo "⚠️  Unexpected response - check above"
fi

echo ""
