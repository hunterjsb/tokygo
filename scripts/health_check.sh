#!/bin/bash

# Health check for tokygo server

set -e

# Get port from environment or use default
PORT=${PORT:-8080}

echo "Checking health of tokygo server on port $PORT..."

# Try to hit the health endpoint
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/health || echo "000")

if [ "$RESPONSE" = "200" ]; then
    echo "Health check passed - Server is healthy"

    # Also verify the response body
    HEALTH_DATA=$(curl -s http://localhost:$PORT/health)
    echo "Health response: $HEALTH_DATA"

    exit 0
else
    echo "Health check failed - HTTP status: $RESPONSE"
    exit 1
fi
