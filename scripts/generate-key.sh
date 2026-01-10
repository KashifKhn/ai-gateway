#!/bin/bash

# Generate a random API key
# Usage: ./generate-key.sh [prefix]

PREFIX=${1:-"sk"}
RANDOM_PART=$(openssl rand -hex 24)
API_KEY="${PREFIX}-${RANDOM_PART}"

echo "Generated API Key:"
echo "$API_KEY"
echo ""
echo "Add to config/config.yaml under auth.keys:"
echo "  keys:"
echo "    - \"$API_KEY\""
echo ""
echo "Or set as environment variable:"
echo "  export AI_GATEWAY_API_KEY=\"$API_KEY\""
