#!/bin/bash

echo "Starting Bible API with Rate Limiting..."

# Check if Redis is running
if ! command -v redis-cli &>/dev/null; then
    echo "Redis CLI not found. Installing Redis..."
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt update && sudo apt install redis-server -y
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install redis
    fi
fi

# Start Redis if not running
if ! redis-cli ping &>/dev/null; then
    echo "Starting Redis server..."
    redis-server --daemonize yes
    sleep 2
fi

# Verify Redis is running
if redis-cli ping &>/dev/null; then
    echo "✅ Redis is running"
else
    echo "❌ Failed to start Redis"
    exit 1
fi

# Build and start the API
echo "Building Bible API..."
go build -o bible_api ./cmd/bible_api.go

echo "Starting Bible API..."
echo "API will be available at: http://localhost:8000"
echo "Health check: http://localhost:8000/health"
echo "Rate limit: 5 requests per second per IP"
echo ""
echo "Press Ctrl+C to stop"

./bible_api -dbPath ./data/kjv.db
