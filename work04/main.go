package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {

	// 时间窗口大小为 1s
	// 将时间窗口平均分为 10 个大小相等的时间片
	// 每个时间片的令牌数量为 10 个, 即限制每个时间片时最多可以处理 10 个令牌请求
	rWindow := NewRollingWindow(1000, 10)
	lm := NewLimiter(100)
	err := rWindow.AddMetric(lm)
	if err != nil {
		log.Fatalf("add rate limiter metric: %v", err)
	}

	_ = rWindow.Start()

	fmt.Println("================================= rate limiter ================================= ")
	ctx, cancelFunc := context.WithCancel(context.Background())

	// 汇总每秒处理的请求总量
	go summaryLimiter(ctx, lm)

	// 不断的随机请求令牌
	go acquire(ctx, lm)
	go acquire(ctx, lm)
	go acquire(ctx, lm)
	go acquire(ctx, lm)
	go acquire(ctx, lm)

	time.Sleep(time.Second * 10)
	cancelFunc()
	_ = rWindow.Stop()

	fmt.Println("================================= counter ================================= ")
	ctx, cancelFunc = context.WithCancel(context.Background())

	ct := NewCounter()
	// 汇总每秒处理的请求总量
	go summaryCounter(ctx, ct)

	// 不断的随机增加计数
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)
	go add(ctx, ct)

	_ = rWindow.AddMetric(ct)
	_ = rWindow.Start()

	time.Sleep(time.Second * 10)

	cancelFunc()
}

// acquire 随机请求令牌
func acquire(ctx context.Context, lm RateLimiter) {
	period := 10 + rand.Int()%20
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Millisecond * time.Duration(period))
			if ok, _ := lm.Acquire(); ok {
				fmt.Println("请求令牌: 成功")
			} else {
				fmt.Println("请求令牌: 失败")
			}
		}
	}
}

// summaryLimiter 统计输出每秒的请求总量
func summaryLimiter(ctx context.Context, lm RateLimiter) {
	i := 1
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second)
			fmt.Printf("============================== 第%d秒, 请求总量: %d\n", i, lm.Total())
			i += 1
		}
	}
}

// add 计数
func add(ctx context.Context, counter Counter) {
	period := 10 + rand.Int()%20
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Millisecond * time.Duration(period))
			counter.Add(1)
		}
	}
}

// summaryCounter 统计输出计数
func summaryCounter(ctx context.Context, counter Counter) {
	i := 1
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second)
			fmt.Printf("============================== 第%d秒, 请求总量: %d\n", i, counter.Total())
			i += 1
		}
	}
}
