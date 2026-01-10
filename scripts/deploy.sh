#!/bin/bash
set -e

if [ ! -f .env ]; then
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "Edit .env with your settings, then run this script again."
    exit 1
fi

source .env

if [ -z "$API_KEY" ] || [ "$API_KEY" = "sk-your-secret-api-key-here" ]; then
    echo "ERROR: Set a real API_KEY in .env"
    exit 1
fi

echo "Starting OpenCode server..."
if ! pgrep -f "opencode serve" > /dev/null; then
    nohup opencode serve --port ${OPENCODE_PORT:-3001} > /tmp/opencode.log 2>&1 &
    sleep 3
fi

echo "Building and starting AI Gateway..."
docker compose up -d --build

echo ""
echo "Deployed!"
echo "  Local:  http://localhost:8090"
echo "  Domain: https://${DOMAIN}"
echo ""
echo "Test: curl -H 'Authorization: Bearer YOUR_KEY' https://${DOMAIN}/health"
