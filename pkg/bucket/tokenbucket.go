package bucket

import (
	"time"
)

type TokenBucket[K comparable] struct {
	max                   int
	refillIntervalSeconds int
	storage               map[K]*Bucket
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
