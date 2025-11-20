package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
	mu             sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Take attempts to take a token from the bucket
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	
	// Refill tokens based on elapsed time
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefillTime = now

	// Try to take a token
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// RateLimiter manages rate limits for multiple clients
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	
	maxTokens  float64
	refillRate float64
	
	// Cleanup old buckets periodically
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// NewRateLimiter creates a new rate limiter
// maxRequests: maximum requests allowed in the time window
// perDuration: time window (e.g., 1 minute)
func NewRateLimiter(maxRequests int, perDuration time.Duration) *RateLimiter {
	refillRate := float64(maxRequests) / perDuration.Seconds()
	
	return &RateLimiter{
		buckets:         make(map[string]*TokenBucket),
		maxTokens:       float64(maxRequests),
		refillRate:      refillRate,
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}
}

// Allow checks if a request should be allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = NewTokenBucket(rl.maxTokens, rl.refillRate)
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	// Periodic cleanup
	rl.cleanup()

	return bucket.Take()
}

// cleanup removes old unused buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	if now.Sub(rl.lastCleanup) < rl.cleanupInterval {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Remove buckets that haven't been used in the last cleanup interval
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if now.Sub(bucket.lastRefillTime) > rl.cleanupInterval {
			delete(rl.buckets, key)
		}
		bucket.mu.Unlock()
	}
	
	rl.lastCleanup = now
}

// IPRateLimit creates middleware for rate limiting by IP address
func IPRateLimit(maxRequests int, perDuration time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, perDuration)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
				"retry_after": int(perDuration.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimit creates middleware for rate limiting by authenticated user
func UserRateLimit(maxRequests int, perDuration time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, perDuration)

	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			// If no user, fall back to IP-based limiting
			userID = c.ClientIP()
		}

		key := userID.(string)
		
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
				"retry_after": int(perDuration.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
