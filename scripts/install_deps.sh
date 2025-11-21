#!/bin/bash

# Install dependencies for tokygo

set -e

# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

cd /opt/tokygo

echo "Installing Go dependencies..."

# Download Go module dependencies
go mod download

echo "Dependencies installed successfully"

exit 0
