#!/bin/bash

# Install dependencies for tokygo

set -e

cd /opt/tokygo

echo "Installing Go dependencies..."

# Download Go module dependencies
go mod download

echo "Dependencies installed successfully"

exit 0
