package ratelimit

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/cabewaldrop/liminal/pkg/middleware/matcher"
)

type Bucket struct {
	mu         *sync.Mutex
	MaxTokens  float64
	Tokens     float64
	LastFilled time.Time
	RefillRate float64 // Number of tokens to refill per second
}

type Limiter interface {
	Allow() bool
}

// Limit calculates whether there are sufficient tokens to
// allow the request.
func (b *Bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Fill the bucket
	now := time.Now()
	elapsed := now.Sub(b.LastFilled)
	amount := elapsed.Seconds() * b.RefillRate
	b.Tokens = math.Min(b.Tokens+amount, b.MaxTokens)
	b.LastFilled = now

	// Evaluate the condition
	if b.Tokens >= 1 {
		b.Tokens = math.Min(b.Tokens-1, 0)
		return true
	}

	return false
}

func newBucket(maxTokens float64, refillRate float64) *Bucket {
	return &Bucket{
		mu:         &sync.Mutex{},
		MaxTokens:  maxTokens,
		Tokens:     maxTokens,
		LastFilled: time.Now(),
		RefillRate: refillRate,
	}
}

func newDefaultBucket() *Bucket {
	return newBucket(10, 1)
}

type RateLimit struct {
	mu      sync.Mutex
	buckets map[string]*Bucket
	matcher matcher.Matcher
	next    http.Handler
}

func NewRateLimiter(strategy string, next http.Handler) *RateLimit {
	return &RateLimit{
		mu:      sync.Mutex{},
		buckets: make(map[string]*Bucket),
		matcher: matcher.NewMatcher(strategy),
		next:    next,
	}
}

func (rl *RateLimit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key, err := rl.matcher.Match(r)
	if err != nil {
		http.Error(w, "Unable to parse the source of the request", http.StatusInternalServerError)
		return
	}

	rl.mu.Lock()
	bucket, ok := rl.buckets[key]
	if !ok {
		bucket = newDefaultBucket()
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	if !bucket.Allow() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	rl.next.ServeHTTP(w, r)
}
