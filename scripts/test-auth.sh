#!/bin/bash

# Simple script to test M-PESA authorization only

set -e

# Load environment variables
source "$(dirname "$0")/../.env"

echo "🔐 Testing M-PESA Authorization..."
echo "Base URL: ${BASE_URL}"
echo "Consumer Key: ${MPESA_CONSUMER_KEY:0:10}..."

# Encode credentials as Base64 for Basic Auth (remove any newlines)
AUTH_HEADER=$(echo -n "${MPESA_CONSUMER_KEY}:${MPESA_CONSUMER_SECRET}" | base64 | tr -d '\n')
echo "Auth Header (first 30 chars): ${AUTH_HEADER:0:30}..."
echo "Auth Header length: ${#AUTH_HEADER}"

# Test with verbose output and timeout
echo ""
echo "📡 Sending request..."
curl -v --max-time 10 --http1.1 \
  -X GET "${BASE_URL}/oauth/v1/generate?grant_type=client_credentials" \
  -H "Authorization: Basic ${AUTH_HEADER}"

echo ""
echo "✅ Request completed"
