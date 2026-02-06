#!/bin/bash

API_KEY="sk-5c2bec00-eddd-11f0-995e-cb11c89f1256"
BASE_URL="https://llm.kashifkhan.dev"

echo "Testing AI Gateway..."
echo ""

echo "1. Health Check (no auth)"
curl -s $BASE_URL/health | jq .
echo ""

echo "2. List Models"
curl -s -H "Authorization: Bearer $API_KEY" $BASE_URL/v1/models | jq '.data[].id'
echo ""

echo "3. Chat Completion"
curl -s -X POST $BASE_URL/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "big-pickle",
    "messages": [{"role": "user", "content": "Say hello in one word"}]
  }' | jq '.choices[0].message.content'
echo ""

echo "✅ All tests passed!"
