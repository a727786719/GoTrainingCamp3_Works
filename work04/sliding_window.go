package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

const (
	stopped = 0
	started = 1
)

type Metric interface {
	moveNext(latest *bucket)
	Init(windowMs int, numBuckets int) error
}

// RollingWindow 滚动窗口
type RollingWindow interface {
	Start() error
	Stop() error
	// AddMetric 添加测量
	AddMetric(Metric) error
}

type counterValue struct {
	value int32
}

func newCounterValue() *counterValue {
	return &counterValue{
		value: 0,
	}
}

func (c *counterValue) Reset() {
	atomic.StoreInt32(&c.value, 0)
}

func (c *counterValue) Add(delta int32) (newValue int32) {
	return atomic.AddInt32(&c.value, delta)
}

func (c *counterValue) GetValue() int32 {
	return atomic.LoadInt32(&c.value)
}

type bucket struct {
	prev  *bucket
	next  *bucket
	start time.Time
}

func newBucket() *bucket {
	return &bucket{}
}

func (b *bucket) Reset() {
	b.start = time.Now()
}

func (b bucket) Since() int64 {
	return time.Since(b.start).Milliseconds()
}



type rollingWindow struct {
	cancelFunc context.CancelFunc
	// 时间窗口大小(ms)
	windowMs int
	// 时间窗口内的令牌桶数量
	numBuckets int
	state      int32
	start      time.Time
	latest     *bucket
	metrics    []Metric
}

func NewRollingWindow(windowMs int, numBuckets int) *rollingWindow {
	return &rollingWindow{
		windowMs:   windowMs,
		numBuckets: numBuckets,
		state:      stopped,
		metrics:    make([]Metric, 0, 2),
	}
}

func (r *rollingWindow) Start() error {
	if !atomic.CompareAndSwapInt32(&r.state, stopped, started) {
		return fmt.Errorf("rolling window has already started")
	}

	r.latest = newBucket()
	cur := r.latest
	for i := 1; i < r.numBuckets; i++ {
		b := newBucket()
		cur.next = b
		b.prev = cur
		cur = b
	}

	cur.next = r.latest
	r.latest.prev = cur

	var ctx context.Context
	ctx, r.cancelFunc = context.WithCancel(context.Background())
	go r.worker(ctx)

	return nil
}

func (r *rollingWindow) Stop() error {
	if !atomic.CompareAndSwapInt32(&r.state, started, stopped) {
		return fmt.Errorf("rolling window has already stopped")
	}

	// stop
	r.cancelFunc()

	return nil
}

func (r *rollingWindow) worker(ctx context.Context) {
	r.moveNext()

	period := r.windowMs / r.numBuckets
	ticker := time.NewTicker(time.Millisecond * time.Duration(period))
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
		case <-ticker.C:
			r.moveNext()
		}
	}
}

func (r *rollingWindow) moveNext() {
	for _, metric := range r.metrics {
		r.latest = r.latest.next
		r.latest.Reset()
		metric.moveNext(r.latest)
	}
}

func (r *rollingWindow) AddMetric(metric Metric) error {
	if r.state == started {
		return fmt.Errorf("cannot create limiter at started state")
	}

	err := metric.Init(r.windowMs, r.numBuckets)
	if err != nil {
		return err
	}

	r.metrics = append(r.metrics, metric)
	return nil
}
