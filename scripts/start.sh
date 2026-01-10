#!/bin/bash

# AI Gateway Start Script
# This script starts both OpenCode server and the AI Gateway

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting AI Gateway Stack...${NC}"
echo ""

# Configuration
OPENCODE_PORT=${OPENCODE_PORT:-3001}
GATEWAY_PORT=${GATEWAY_PORT:-8080}
OPENCODE_MODEL=${OPENCODE_MODEL:-"opencode/big-pickle"}

# Check if opencode is installed
if ! command -v opencode &> /dev/null; then
    echo -e "${RED}Error: opencode is not installed${NC}"
    echo "Install it with: curl -fsSL https://opencode.ai/install | bash"
    exit 1
fi

# Check if gateway binary exists
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATEWAY_BIN="$SCRIPT_DIR/../ai-gateway"

if [ ! -f "$GATEWAY_BIN" ]; then
    echo -e "${YELLOW}Gateway binary not found. Building...${NC}"
    cd "$SCRIPT_DIR/.."
    go build -o ai-gateway ./cmd/server
    echo -e "${GREEN}Build complete!${NC}"
fi

# Start OpenCode server in background
echo -e "${GREEN}Starting OpenCode server on port $OPENCODE_PORT...${NC}"
opencode serve --port $OPENCODE_PORT &
OPENCODE_PID=$!

# Wait for OpenCode to start
sleep 2

# Check if OpenCode is running
if ! kill -0 $OPENCODE_PID 2>/dev/null; then
    echo -e "${RED}Error: OpenCode failed to start${NC}"
    exit 1
fi

echo -e "${GREEN}OpenCode started (PID: $OPENCODE_PID)${NC}"

# Trap to clean up on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down...${NC}"
    kill $OPENCODE_PID 2>/dev/null || true
    wait $OPENCODE_PID 2>/dev/null || true
    echo -e "${GREEN}Cleanup complete${NC}"
}
trap cleanup EXIT

# Start AI Gateway
echo -e "${GREEN}Starting AI Gateway on port $GATEWAY_PORT...${NC}"
cd "$SCRIPT_DIR/.."
CONFIG_PATH=config/config.yaml ./ai-gateway
