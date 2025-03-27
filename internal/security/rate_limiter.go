package security

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Rate represents a rate limit defined as count per duration
type Rate struct {
	Count    int
	Duration time.Duration
}

// RateLimitKey is the key type for rate limiting (e.g., user ID, blueprint ID)
type RateLimitKey string

// RateLimitType represents different types of rate limits
type RateLimitType string

const (
	RateLimitTypeUser             RateLimitType = "user"           // Limit per user
	RateLimitTypeBlueprint        RateLimitType = "blueprint"      // Limit per blueprint
	RateLimitTypeUserAndBlueprint RateLimitType = "user_blueprint" // Limit per user and blueprint
	RateLimitTypeAPI              RateLimitType = "api"            // Limit for API calls
	RateLimitTypeGlobal           RateLimitType = "global"         // Global rate limit
)

// TokenBucket implements the token bucket algorithm for rate limiting
type TokenBucket struct {
	capacity   int
	tokens     float64
	lastRefill time.Time
	refillRate float64 // tokens per second
	mutex      sync.Mutex
}

// NewTokenBucket creates a new token bucket with specified capacity and refill rate
func NewTokenBucket(capacity int, refillPeriod time.Duration) *TokenBucket {
	refillRate := float64(capacity) / refillPeriod.Seconds()
	return &TokenBucket{
		capacity:   capacity,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
		refillRate: refillRate,
	}
}

// TakeToken attempts to take a token from the bucket
// Returns true if successful, false if no tokens are available
func (tb *TokenBucket) TakeToken() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = tb.tokens + (elapsed * tb.refillRate)
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	tb.lastRefill = now

	// Check if we have enough tokens
	if tb.tokens < 1 {
		return false
	}

	// Take a token
	tb.tokens--
	return true
}

// AvailableTokens returns the number of available tokens
func (tb *TokenBucket) AvailableTokens() float64 {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tokens := tb.tokens + (elapsed * tb.refillRate)
	if tokens > float64(tb.capacity) {
		tokens = float64(tb.capacity)
	}

	return tokens
}

// RateLimiter handles rate limiting for blueprints and users
type RateLimiter struct {
	limits  map[RateLimitType]Rate
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with default limits
func NewRateLimiter() *RateLimiter {
	limiter := &RateLimiter{
		limits: map[RateLimitType]Rate{
			RateLimitTypeUser: {
				Count:    100,
				Duration: time.Hour,
			},
			RateLimitTypeBlueprint: {
				Count:    30,
				Duration: time.Hour,
			},
			RateLimitTypeUserAndBlueprint: {
				Count:    20,
				Duration: time.Minute * 10,
			},
			RateLimitTypeAPI: {
				Count:    1000,
				Duration: time.Hour,
			},
			RateLimitTypeGlobal: {
				Count:    10000,
				Duration: time.Hour,
			},
		},
		buckets: make(map[string]*TokenBucket),
	}

	return limiter
}

// SetRateLimit sets a rate limit for a specific type
func (rl *RateLimiter) SetRateLimit(limitType RateLimitType, count int, duration time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.limits[limitType] = Rate{
		Count:    count,
		Duration: duration,
	}

	// Update existing buckets of this type
	for key, bucket := range rl.buckets {
		if keyParts := splitRateLimitKey(key); keyParts[0] == string(limitType) {
			refillRate := float64(count) / duration.Seconds()

			// Create a new bucket with the updated rate
			rl.buckets[key] = &TokenBucket{
				capacity:   count,
				tokens:     bucket.tokens, // Preserve current tokens
				lastRefill: bucket.lastRefill,
				refillRate: refillRate,
			}
		}
	}
}

// createBucketKey creates a key for storing token buckets
func createBucketKey(limitType RateLimitType, key RateLimitKey) string {
	return fmt.Sprintf("%s:%s", limitType, key)
}

// splitRateLimitKey splits a bucket key into its components
func splitRateLimitKey(key string) []string {
	result := make([]string, 0, 2)
	for i, c := range key {
		if c == ':' {
			result = append(result, key[:i], key[i+1:])
			return result
		}
	}
	return []string{key}
}

// getBucket gets or creates a token bucket for a key and limit type
func (rl *RateLimiter) getBucket(limitType RateLimitType, key RateLimitKey) *TokenBucket {
	bucketKey := createBucketKey(limitType, key)

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if bucket, exists := rl.buckets[bucketKey]; exists {
		return bucket
	}

	// Create new bucket if it doesn't exist
	rate, exists := rl.limits[limitType]
	if !exists {
		// Use default rate if not configured
		rate = Rate{
			Count:    100,
			Duration: time.Hour,
		}
	}

	bucket := NewTokenBucket(rate.Count, rate.Duration)
	rl.buckets[bucketKey] = bucket
	return bucket
}

// Allow checks if a request should be allowed based on rate limits
// Returns true if allowed, false if rate limited
func (rl *RateLimiter) Allow(limitType RateLimitType, key RateLimitKey) bool {
	bucket := rl.getBucket(limitType, key)
	return bucket.TakeToken()
}

// AllowN checks if a request for n tokens should be allowed
// Used for operations that may consume multiple "quota units"
func (rl *RateLimiter) AllowN(limitType RateLimitType, key RateLimitKey, n int) bool {
	bucket := rl.getBucket(limitType, key)

	// Check if we have enough tokens
	if bucket.AvailableTokens() < float64(n) {
		return false
	}

	// Take tokens one by one
	for i := 0; i < n; i++ {
		if !bucket.TakeToken() {
			return false
		}
	}

	return true
}

// GetRemainingTokens returns the number of tokens remaining for a key
func (rl *RateLimiter) GetRemainingTokens(limitType RateLimitType, key RateLimitKey) float64 {
	bucket := rl.getBucket(limitType, key)
	return bucket.AvailableTokens()
}

// Reset resets rate limiting for a specific key
func (rl *RateLimiter) Reset(limitType RateLimitType, key RateLimitKey) {
	bucketKey := createBucketKey(limitType, key)

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rate, exists := rl.limits[limitType]
	if !exists {
		rate = Rate{
			Count:    100,
			Duration: time.Hour,
		}
	}

	rl.buckets[bucketKey] = NewTokenBucket(rate.Count, rate.Duration)
}

// CreateUserBlueprintKey creates a combined key for user and blueprint
func CreateUserBlueprintKey(userID, blueprintID string) RateLimitKey {
	return RateLimitKey(fmt.Sprintf("%s:%s", userID, blueprintID))
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	LimitType RateLimitType
	Key       RateLimitKey
	Limit     Rate
	Remaining float64
	Reset     time.Duration
}

// Error returns the error message
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("Rate limit exceeded for %s:%s. Limit: %d per %v. Remaining: %.2f. Try again in %v",
		e.LimitType, e.Key, e.Limit.Count, e.Limit.Duration, e.Remaining, e.Reset)
}

// CheckRateLimit checks if a request is allowed and returns an error if not
func (rl *RateLimiter) CheckRateLimit(limitType RateLimitType, key RateLimitKey) error {
	if !rl.Allow(limitType, key) {
		bucket := rl.getBucket(limitType, key)
		remaining := bucket.AvailableTokens()

		rate := rl.limits[limitType]

		// Calculate reset time (approximate)
		tokensNeeded := 1.0 - remaining
		if tokensNeeded < 0 {
			tokensNeeded = 0
		}

		resetSeconds := (tokensNeeded / bucket.refillRate)
		resetDuration := time.Duration(resetSeconds * float64(time.Second))

		return &RateLimitError{
			LimitType: limitType,
			Key:       key,
			Limit:     rate,
			Remaining: remaining,
			Reset:     resetDuration,
		}
	}

	return nil
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}
