#!/bin/bash

echo "Testing Rate Limiting with Logging..."
echo "Watch the server logs to see rate limiting events."
echo ""

API_URL="http://localhost:8000/bible/random_verse"

echo "Starting rate limit test..."
echo "Sending 8 requests rapidly (limit is 5 per second)"
echo ""

for i in {1..8}; do
    echo "Request $i..."
    curl -s -w "HTTP Status: %{http_code} | Rate Limit Remaining: %{header_x-ratelimit-remaining}\n" \
        "$API_URL?json=true" >/dev/null
    sleep 0.1
done

echo ""
echo "Test completed. Check server logs for:"
echo "  - Warning logs when approaching limit"
echo "  - Error logs when IP is blocked"
echo "  - Blocked IP access attempts"
echo ""
echo "Expected log messages:"
echo "  Rate limiter: IP 127.0.0.1 approaching limit (4/5 requests used)"
echo "  Rate limiter: IP 127.0.0.1 exceeded limit (6 requests), blocking for 1m0s"
echo "  Rate limiter: Blocked IP 127.0.0.1 attempted access (still blocked)"
