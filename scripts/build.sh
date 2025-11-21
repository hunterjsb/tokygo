#!/bin/bash

# Build script for GitHub Pages deployment
# This script generates the frontend with the correct API URL

set -e

echo "🏗️  Building frontend for deployment..."

# Use environment variables from command line, or load from .env, or use defaults
if [ -z "$API_URL" ] || [ -z "$MAPBOX_TOKEN" ]; then
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
    fi
fi

# Final fallback to defaults
API_URL=${API_URL:-"http://localhost:8080"}
MAPBOX_TOKEN=${MAPBOX_TOKEN:-""}
echo "📡 API URL: $API_URL"
echo "🗺️  Mapbox Token: ${MAPBOX_TOKEN:0:20}..."

# Create dist directory
DIST_DIR="dist"
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Copy frontend files to dist
echo "📦 Copying frontend files..."
cp -r frontend/* $DIST_DIR/

# Inject API_URL and MAPBOX_TOKEN into map.js
echo "⚙️  Injecting API_URL and MAPBOX_TOKEN into map.js..."
sed -i.bak "s|const API_BASE_URL = .*|const API_BASE_URL = \"$API_URL\";|g" $DIST_DIR/js/map.js
sed -i.bak "s|const MAPBOX_TOKEN = .*|const MAPBOX_TOKEN = \"$MAPBOX_TOKEN\";|g" $DIST_DIR/js/map.js
rm $DIST_DIR/js/map.js.bak

echo "✅ Build complete! Output in $DIST_DIR/"
echo ""
echo "To test locally:"
echo "  cd $DIST_DIR && python3 -m http.server 8000"
echo ""
echo "To deploy to GitHub Pages:"
echo "  Copy contents of $DIST_DIR to your gh-pages branch"
