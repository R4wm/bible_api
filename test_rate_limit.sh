#!/bin/bash

# Test script for rate limiting
echo "Testing Bible API Rate Limiting..."
echo "Sending 10 requests rapidly to test rate limiting (limit: 5 req/sec)"
echo

API_URL="http://localhost:8000/bible/random_verse"

for i in {1..10}; do
    echo "Request $i:"
    response=$(curl -s -w "HTTP_CODE:%{http_code}\nRESPONSE_TIME:%{time_total}s\nRATE_LIMIT:%{header_x-ratelimit-remaining}\n" \
        -H "Content-Type: application/json" \
        "$API_URL?json=true" 2>/dev/null)

    # Extract just the status code and rate limit info
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    rate_limit=$(echo "$response" | grep "RATE_LIMIT:" | cut -d: -f2)
    response_time=$(echo "$response" | grep "RESPONSE_TIME:" | cut -d: -f2)

    echo "  Status: $http_code | Remaining: $rate_limit | Time: $response_time"

    if [ "$http_code" = "429" ]; then
        echo "  ⚠️  Rate limited!"
    elif [ "$http_code" = "200" ]; then
        echo "  ✅ Success"
    else
        echo "  ❌ Error"
    fi
    echo

    # Small delay to see the sliding window effect
    sleep 0.1
done

echo "Test completed. If working correctly, you should see 429 errors after the 5th request."
echo "The IP should be blocked for 1 minute after exceeding the rate limit."
