# Bible API Rate Limiting

This Bible API now includes Redis-based rate limiting to prevent abuse and ensure fair usage.

## Rate Limiting Configuration

- **Limit**: 5 requests per second per IP address
- **Block Duration**: 1 minute when limit is exceeded
- **Storage**: Redis with TTL for automatic cleanup

## Setup Instructions

### 1. Install Redis

#### Using Docker (Recommended)

```bash
# Start Redis and Bible API together
docker-compose up -d

# Or just Redis
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

#### Using Package Manager

```bash
# Ubuntu/Debian
sudo apt install redis-server

# macOS
brew install redis

# Start Redis
redis-server
```

### 2. Environment Variables

Set these environment variables if needed:

```bash
export REDIS_ADDR=localhost:6379    # Redis server address
export REDIS_PASSWORD=              # Redis password (if any)
```

### 3. Run the API

```bash
go run cmd/bible_api.go -dbPath ./data/kjv.db
```

## Rate Limiting Behavior

### Normal Operation

- Each request consumes 1 request from the allowance
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Maximum requests allowed per second
  - `X-RateLimit-Remaining`: Requests remaining in current window
  - `X-RateLimit-Reset`: Unix timestamp when limit resets

### When Limit Exceeded

- IP is blocked for 1 minute
- HTTP 429 (Too Many Requests) response
- JSON error message with retry information

### Example Headers

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset: 1640995200
```

## Testing Rate Limiting

Use the provided test script:

```bash
./test_rate_limit.sh
```

Or test manually:

```bash
# Send rapid requests
for i in {1..10}; do
  curl -s -w "Status: %{http_code}\n" \
    "http://localhost:8000/bible/random_verse?json=true"
done
```

## Admin Endpoints

Manage rate limiting through admin endpoints:

### Check IP Status

```bash
GET /admin/rate-limit/{ip}

# Example
curl "http://localhost:8000/admin/rate-limit/192.168.1.100"
```

### Block IP Manually

```bash
POST /admin/block-ip
Content-Type: application/json

{
  "ip": "192.168.1.100",
  "duration": "10m"
}

# Example
curl -X POST "http://localhost:8000/admin/block-ip" \
  -H "Content-Type: application/json" \
  -d '{"ip": "192.168.1.100", "duration": "10m"}'
```

### Unblock IP

```bash
DELETE /admin/unblock-ip/{ip}

# Example
curl -X DELETE "http://localhost:8000/admin/unblock-ip/192.168.1.100"
```

## Health Check

The health check endpoint bypasses rate limiting:

```bash
curl "http://localhost:8000/health"
```

## Redis Keys Used

The rate limiter uses these Redis key patterns:

- `rate_limit:{ip}` - Request counter with TTL
- `blocked:{ip}` - Block status with TTL

## Graceful Degradation

If Redis is unavailable:

- Requests are allowed through (no rate limiting)
- Error is logged
- Service remains operational

## Production Considerations

1. **Redis Persistence**: Configure Redis persistence for production
2. **Redis Clustering**: Use Redis Cluster for high availability
3. **Monitoring**: Monitor Redis performance and memory usage
4. **Authentication**: Add authentication to admin endpoints
5. **Logging**: Implement structured logging for rate limit events

## Configuration Options

You can modify rate limiting by changing values in `middleware/rate_limiter.go`:

```go
return &RateLimiter{
    client:   redisClient,
    limit:    5,           // Requests per second
    window:   time.Second, // Time window
    blockTTL: time.Minute, // Block duration
}
```

## Troubleshooting

### Redis Connection Issues

```bash
# Test Redis connection
redis-cli ping

# Check Redis logs
docker logs bible_api_redis
```

### Rate Limiting Not Working

1. Verify Redis is running and accessible
2. Check environment variables
3. Look for Redis connection errors in logs
4. Test with the provided test script

### Performance Issues

1. Monitor Redis memory usage
2. Check Redis connection pool settings
3. Consider Redis optimization for high traffic
