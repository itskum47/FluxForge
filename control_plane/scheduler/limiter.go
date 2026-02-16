package scheduler

import (
	"sync"
	"time"

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

// Reserve checks permission and returns a delay if limit is exceeded.
func (l *TokenBucketLimiter) Reserve(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(l.r, l.b)
		l.limiters[key] = limiter
	}

	r := limiter.Reserve()
	delay := r.Delay()
	if delay > 0 {
		r.Cancel() // We are just checking, so cancel the reservation
		return false, delay
	}
	return true, 0
}

// EnsureLimiter guarantees a limiter exists for the key (used for health init)
func (l *TokenBucketLimiter) EnsureLimiter(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.limiters[key]; !exists {
		l.limiters[key] = rate.NewLimiter(l.r, l.b)
	}
}

// ReducedRateLimiter is a wrapper that can enforce a stricter limit for failure domains.
type DynamicLimiter struct {
	limiter *TokenBucketLimiter
}
