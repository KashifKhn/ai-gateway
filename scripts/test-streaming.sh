#!/bin/bash

API_KEY="sk-5c2bec00-eddd-11f0-995e-cb11c89f1256"
BASE_URL="https://llm.kashifkhan.dev"

echo "=== Testing Streaming ==="
echo ""
echo "Sending request with stream=true..."
echo ""

curl -X POST "$BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini",
    "messages": [{"role": "user", "content": "Write a short poem about AI in 4 lines"}],
    "stream": true
  }' \
  --no-buffer

echo ""
echo ""
echo "=== Done ===" 
