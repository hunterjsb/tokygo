#!/bin/bash

# Start the tokygo server

set -e

cd /opt/tokygo

echo "Starting tokygo server..."

# Create logs directory if it doesn't exist
mkdir -p logs

# Load environment variables if .env exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set default port if not specified
export PORT=${PORT:-8080}

# Start the server in the background
nohup ./tokygo > logs/server.log 2>&1 &

# Save the PID
echo $! > tokygo.pid

echo "Server started with PID $(cat tokygo.pid)"

# Wait a moment to ensure it starts
sleep 2

# Check if process is running
if pgrep -f "tokygo" > /dev/null; then
    echo "Server is running"
else
    echo "Failed to start server"
    exit 1
fi

exit 0
