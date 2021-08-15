package main

import "sync"

type Counter interface {
	Metric
	Add(delta int) (newValue int)
	Total() int32
}

type counter struct {
	latest  *bucket
	buckets sync.Map
}

func NewCounter() *counter {
	return &counter{}
}

func (c *counter) Init(windowMs int, numBuckets int) error {
	return nil
}

func (c *counter) createLatest() (*bucket, *counterValue) {
	latest := c.latest

	if v, ok := c.buckets.Load(latest); ok {
		return latest, v.(*counterValue)
	}

	newValue := newCounterValue()
	c.buckets.Store(latest, newValue)
	return latest, newValue
}

func (c *counter) moveNext(latest *bucket) {
	c.latest = latest
	value, ok := c.buckets.Load(latest)
	if ok {
		value.(*counterValue).Reset()
	} else {
		c.createLatest()
	}
}

func (c *counter) Add(delta int) int {
	_, value := c.createLatest()
	return int(value.Add(int32(delta)))
}

func (c *counter) Total() int32 {
	var total int32 = 0
	c.buckets.Range(func(key, value interface{}) bool {
		total += value.(*counterValue).GetValue()
		return true
	})
	return total
}