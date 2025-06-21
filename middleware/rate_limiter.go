package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client   *redis.Client
	limit    int           // requests per second
	window   time.Duration // time window
	blockTTL time.Duration // how long to block when limit exceeded
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		client:   redisClient,
		limit:    5,           // 5 requests per second
		window:   time.Second, // 1 second window
		blockTTL: time.Minute, // 1 minute block
	}
}

func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in case of multiple
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		clientIP := rl.getClientIP(r)

		// Check if IP is currently blocked
		blockedKey := fmt.Sprintf("blocked:%s", clientIP)
		blocked, err := rl.client.Get(ctx, blockedKey).Result()
		if err != redis.Nil && blocked != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded. IP blocked for 1 minute.", "retry_after": 60}`))
			return
		}

		// Rate limiting key
		rateLimitKey := fmt.Sprintf("rate_limit:%s", clientIP)

		// Get current request count using pipeline for atomicity
		pipe := rl.client.Pipeline()
		incr := pipe.Incr(ctx, rateLimitKey)
		pipe.Expire(ctx, rateLimitKey, rl.window)
		_, err = pipe.Exec(ctx)

		if err != nil {
			// If Redis is down, allow the request but log the error
			fmt.Printf("Redis error: %v\n", err)
			next.ServeHTTP(w, r)
			return
		}

		requests := incr.Val()

		// Check if limit exceeded
		if requests > int64(rl.limit) {
			// Block the IP for the specified duration
			err = rl.client.Set(ctx, blockedKey, "1", rl.blockTTL).Err()
			if err != nil {
				fmt.Printf("Error setting block key: %v\n", err)
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.blockTTL).Unix(), 10))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded. IP blocked for 1 minute.", "retry_after": 60}`))
			return
		}

		// Add rate limit headers for successful requests
		remaining := rl.limit - int(requests)
		if remaining < 0 {
			remaining = 0
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.window).Unix(), 10))

		next.ServeHTTP(w, r)
	})
}

// Optional: Method to manually block an IP
func (rl *RateLimiter) BlockIP(ip string, duration time.Duration) error {
	ctx := context.Background()
	blockedKey := fmt.Sprintf("blocked:%s", ip)
	return rl.client.Set(ctx, blockedKey, "1", duration).Err()
}

// Optional: Method to unblock an IP
func (rl *RateLimiter) UnblockIP(ip string) error {
	ctx := context.Background()
	blockedKey := fmt.Sprintf("blocked:%s", ip)
	return rl.client.Del(ctx, blockedKey).Err()
}

// Optional: Method to get current rate limit status for an IP
func (rl *RateLimiter) GetRateLimitStatus(ip string) (requests int64, blocked bool, err error) {
	ctx := context.Background()

	// Check if blocked
	blockedKey := fmt.Sprintf("blocked:%s", ip)
	blockedVal, err := rl.client.Get(ctx, blockedKey).Result()
	if err == redis.Nil {
		blocked = false
	} else if err != nil {
		return 0, false, err
	} else {
		blocked = blockedVal != ""
	}

	// Get current request count
	rateLimitKey := fmt.Sprintf("rate_limit:%s", ip)
	requests, err = rl.client.Get(ctx, rateLimitKey).Int64()
	if err == redis.Nil {
		requests = 0
		err = nil
	}

	return requests, blocked, err
}
