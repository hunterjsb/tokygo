#!/bin/bash

# Build the tokygo server binary

set -e

# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

cd /opt/tokygo

echo "Building tokygo server..."

# Build the Go binary
go build -o tokygo cmd/server/main.go

# Make sure the binary is executable
chmod +x tokygo

echo "Build completed successfully"

exit 0
