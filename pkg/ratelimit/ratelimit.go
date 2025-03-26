package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

type TokenBucket[K comparable] struct {
	max                   int
	refillIntervalSeconds int
	storage               map[K]*Bucket
	mu                    sync.Mutex
}

type Bucket struct {
	Count      int
	RefilledAt int64
}

func NewTokenBucket[K comparable](max int, refillIntervalSeconds int) *TokenBucket[K] {
	return &TokenBucket[K]{
		max:                   max,
		refillIntervalSeconds: refillIntervalSeconds,
		storage:               make(map[K]*Bucket),
	}
}

func (tb *TokenBucket[K]) Check(key K, cost int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	bucket, exists := tb.storage[key]
	if !exists {
		return true
	}

	now := time.Now().UnixMilli()
	refill := (now - bucket.RefilledAt) / int64(tb.refillIntervalSeconds*1000)

	if refill > 0 {
		newCount := bucket.Count + int(refill)
		newCount = min(newCount, tb.max)
		return newCount >= cost
	}
	return bucket.Count >= cost
}

func (tb *TokenBucket[K]) Consume(key K, cost int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now().UnixMilli()
	bucket, exists := tb.storage[key]

	if !exists {
		newBucket := &Bucket{
			Count:      tb.max - cost,
			RefilledAt: now,
		}
		tb.storage[key] = newBucket
		return true
	}

	refill := (now - bucket.RefilledAt) / int64(tb.refillIntervalSeconds*1000)
	if refill > 0 {
		bucket.Count += int(refill)
		if bucket.Count > tb.max {
			bucket.Count = tb.max
		}
		bucket.RefilledAt = now
	}

	if bucket.Count < cost {
		return false
	}

	bucket.Count -= cost
	return true
}

var globalBucket = NewTokenBucket[string](100, 1)

func GlobalGETRateLimit(r *http.Request) bool {
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		return true
	}
	return globalBucket.Consume(clientIP, 1)
}

func GlobalPOSTRateLimit(r *http.Request) bool {
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		return true
	}
	return globalBucket.Consume(clientIP, 3)
}
