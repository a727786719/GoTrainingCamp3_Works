package main

import "sync"

type RateLimiter interface {
	Metric
	Acquire() (bool, int32)
	Total() int32
}

type limiter struct {
	// 并发控制
	concurrency int32
	// 每个桶的并发控制
	bucketConcurrency int32
	latest            *bucket
	// buckets           map[*bucket]*counterValue
	buckets sync.Map
}

func NewLimiter(concurrency int) *limiter {
	return &limiter{
		concurrency: int32(concurrency),
		buckets:     sync.Map{},
	}
}

func (l *limiter) Init(windowMs int, numBuckets int) error {
	l.bucketConcurrency = l.concurrency / int32(numBuckets)
	return nil
}

func (l *limiter) Total() int32 {
	var total int32 = 0
	l.buckets.Range(func(key, value interface{}) bool {
		total += value.(*counterValue).GetValue()
		return true
	})
	return total
}

func (l *limiter) createLatest() (*bucket, *counterValue) {
	latest := l.latest

	if v, ok := l.buckets.Load(latest); ok {
		return latest, v.(*counterValue)
	}

	newValue := newCounterValue()
	l.buckets.Store(latest, newValue)
	return latest, newValue
}

func (l *limiter) moveNext(latest *bucket) {
	l.latest = latest
	_, val := l.createLatest()
	val.Reset()
}

func (l *limiter) Acquire() (bool, int32) {
	_, value := l.createLatest()
	newValue := value.Add(1)
	if newValue <= l.bucketConcurrency {
		return true, newValue
	}

	newValue = value.Add(-1)

	return false, newValue
}
