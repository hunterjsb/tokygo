#!/bin/bash

# Build script for GitHub Pages deployment
# This script builds the React + Vite frontend with the correct API URL

set -e

echo "🏗️  Building frontend for deployment..."

# Navigate to frontend directory
cd frontend

# Load from root .env if not set
if [ -f ../.env ]; then
    export $(cat ../.env | grep -v '^#' | xargs)
fi

# Use environment variables or defaults
API_URL=${API_URL:-"http://localhost:8080"}
BASE_PATH=${BASE_PATH:-/}
echo "📡 API URL: $API_URL"
echo "🔗 BASE_PATH: $BASE_PATH"

# Create .env file for Vite build
echo "⚙️  Creating .env file for Vite..."
cat > .env << EOF
VITE_API_BASE_URL=$API_URL
VITE_BASE_PATH=$BASE_PATH
EOF

echo "📄 Generated .env file:"
cat .env

# Export VITE_BASE_PATH so vite.config.ts can read it
export VITE_BASE_PATH=$BASE_PATH

# Install dependencies
echo "📦 Installing dependencies..."
npm ci

# Build the frontend
echo "🔨 Building React app..."
npm run build

# Move dist to project root for GitHub Pages
echo "📤 Preparing dist for GitHub Pages..."
cd ..
rm -rf dist
cp -r frontend/dist ./dist

echo "✅ Build complete! Output in ./dist/"
echo ""
echo "To test locally:"
echo "  cd dist && python3 -m http.server 8000"
echo ""
echo "Note: The built app uses free CARTO map tiles and expects the backend API at $API_URL"
