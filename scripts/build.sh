#!/bin/bash

# Build script for GitHub Pages deployment
# This script generates the frontend with the correct API URL

set -e

echo "🏗️  Building frontend for deployment..."

# Use environment variable from command line, or load from .env, or use default
if [ -z "$API_URL" ]; then
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
    fi
fi

# Final fallback to default
API_URL=${API_URL:-"http://localhost:8080"}
echo "📡 API URL: $API_URL"

# Create dist directory
DIST_DIR="dist"
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Copy frontend files to dist
echo "📦 Copying frontend files..."
cp -r frontend/* $DIST_DIR/

# Inject API_URL into map.js
echo "⚙️  Injecting API_URL into map.js..."
sed -i.bak "s|const API_BASE_URL = .*|const API_BASE_URL = \"$API_URL\";|g" $DIST_DIR/js/map.js
rm $DIST_DIR/js/map.js.bak

echo "✅ Build complete! Output in $DIST_DIR/"
echo ""
echo "To test locally:"
echo "  cd $DIST_DIR && python3 -m http.server 8000"
echo ""
echo "To deploy to GitHub Pages:"
echo "  Copy contents of $DIST_DIR to your gh-pages branch"
