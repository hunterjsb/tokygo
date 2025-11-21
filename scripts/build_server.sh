#!/bin/bash

# Build the tokygo server binary

set -e

cd /opt/tokygo

echo "Building tokygo server..."

# Build the Go binary
go build -o tokygo cmd/server/main.go

# Make sure the binary is executable
chmod +x tokygo

echo "Build completed successfully"

exit 0
