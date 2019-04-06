package tba

import (
	"sync/atomic"
	"time"
)

const (
	// Internal "minimum" time for ticks. Too small times can generate unwanted overheads
	minTickResolution = time.Millisecond
)

type Bucket struct {
	currCnt    int64
	bucketSize int64
	tokenAdd   int64
	period     time.Duration
	done       chan int
}

// Create a bucket that limits QPS (queries per second)
func NewQPSLimit(limitQPS int64) *Bucket {
	return newQueryPerDuration(limitQPS, time.Second)
}

// Create a bucket that limits QPM (queries per minutes)
func NewQPMLimit(limitQPM int64) *Bucket {
	return newQueryPerDuration(limitQPM, time.Minute)
}

func NewBucket(bucketSize, tokenAdd int64, period time.Duration) *Bucket {
	b := &Bucket{
		currCnt:    0,
		bucketSize: bucketSize,
		tokenAdd:   tokenAdd,
		period:     period,
		done:       make(chan int, 1),
	}
	go b.start()
	return b
}

// Stop working, shutdown
func (b *Bucket) Stop() {
	close(b.done)
}

func (b *Bucket) Ask() bool {
	return b.AskN(1)
}

func (b *Bucket) AskN(v int64) bool {
	if v == 0 {
		return true
	}
	if v < 0 {
		panic("Buggy request")
	}
	newVal := atomic.AddInt64(&b.currCnt, -v)
	if newVal < 0 {
		//Return v back and reject
		atomic.AddInt64(&b.currCnt, v)
		return false
	}
	return true
}

/// Internal-use ///

func newQueryPerDuration(limit int64, dur time.Duration) *Bucket {
	// Calc time quantile we have per 1 "query"
	t := time.Duration(dur.Nanoseconds() / limit)
	var add int64 = 1
	for t < minTickResolution {
		t *= 10
		add *= 10
	}
	return NewBucket(add, add, t)
}

func (b *Bucket) start() {
	t := time.NewTicker(b.period)
	defer t.Stop()

	for {
		select {
		case <-b.done:
			return
		case <-t.C:
			newVal := atomic.AddInt64(&b.currCnt, b.tokenAdd)
			if newVal > b.bucketSize {
				// Withdraw excess
				atomic.AddInt64(&b.currCnt, b.bucketSize-newVal)
			}
		} //select
	} //for
}
