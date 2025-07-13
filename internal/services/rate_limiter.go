package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiter handles rate limiting using Redis
type RateLimiter struct {
	client *redis.Client
	ctx    context.Context
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisURL string) *RateLimiter {
	// Parse Redis URL (format: redis://host:port)
	var addr string
	if strings.HasPrefix(redisURL, "redis://") {
		addr = strings.TrimPrefix(redisURL, "redis://")
	} else {
		addr = redisURL
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &RateLimiter{
		client: client,
		ctx:    context.Background(),
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware(config RateLimitConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			clientIP := getClientIP(r)

			// Check rate limits
			allowed, remaining, resetTime, err := rl.checkRateLimit(clientIP, config)
			if err != nil {
				// If Redis is unavailable, allow the request (fail open)
				next(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				errorResponse := map[string]interface{}{
					"error":   "Too Many Requests",
					"code":    429,
					"message": "Rate limit exceeded. Please try again later.",
					"time":    time.Now().UTC().Format(time.RFC3339),
				}

				// Convert to JSON and send
				jsonResponse, _ := json.Marshal(errorResponse)
				w.Write(jsonResponse)
				return
			}

			next(w, r)
		}
	}
}

// checkRateLimit checks if the request is within rate limits
func (rl *RateLimiter) checkRateLimit(clientIP string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-time.Minute) // 1-minute sliding window

	// Create Redis key for this client and window
	key := fmt.Sprintf("rate_limit:%s:%d", clientIP, windowStart.Unix()/60)

	// Get current count
	count, err := rl.client.Get(rl.ctx, key).Int()
	if err == redis.Nil {
		// Key doesn't exist, start fresh
		count = 0
	} else if err != nil {
		return false, 0, 0, err
	}

	// Check if within limits
	if count >= config.RequestsPerMinute {
		// Calculate reset time (next minute)
		resetTime := windowStart.Add(time.Minute).Unix()
		return false, 0, resetTime, nil
	}

	// Increment counter
	pipe := rl.client.Pipeline()
	pipe.Incr(rl.ctx, key)
	pipe.Expire(rl.ctx, key, time.Minute) // Expire after 1 minute

	_, err = pipe.Exec(rl.ctx)
	if err != nil {
		return false, 0, 0, err
	}

	// Calculate remaining requests
	remaining := config.RequestsPerMinute - count - 1
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time
	resetTime := windowStart.Add(time.Minute).Unix()

	return true, remaining, resetTime, nil
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (for proxy/load balancer)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// Close closes the Redis connection
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}
