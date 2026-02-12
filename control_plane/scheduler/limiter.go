package scheduler

import (
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	Allow(key string) bool
}

// TokenBucketLimiter implements RateLimiter using token buckets.
type TokenBucketLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	r        rate.Limit
	b        int
}

// NewTokenBucketLimiter creates a new limiter with rate r tokens per second and burst b.
// Using generic rate.Limit for flexibility.
func NewTokenBucketLimiter(r float64, b int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(r),
		b:        b,
	}
}

// Allow checks if the key is allowed to proceed.
func (l *TokenBucketLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(l.r, l.b)
		l.limiters[key] = limiter
	}

	return limiter.Allow()
}

// ReducedRateLimiter is a wrapper that can enforce a stricter limit for failure domains.
type DynamicLimiter struct {
	limiter *TokenBucketLimiter
}
