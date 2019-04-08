package tba

import (
	"sync/atomic"
	"sync"
	"runtime"
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
	mu         sync.RWMutex
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

// Stop working, shutdown. Concurrent calls to this object can hang/crash
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

// Wait&acquire 'v' tokens
func (b *Bucket) Wait(v int64) {
	for {
		if b.AskN(v) {
			return
		}
		// Lock-unlock to ensure token add event was happen
		b.mu.Lock()
		b.mu.Unlock()
	}
}

func (b *Bucket)MaxBurst(sz int64) {
	atomic.StoreInt64(&b.bucketSize, sz);
	b.Fill()
}

func (b *Bucket)Fill() {
	atomic.StoreInt64(&b.currCnt, b.bucketSize)
}

func (b *Bucket)Drain() {
	atomic.StoreInt64(&b.currCnt, 0)
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
		b.mu.RLock() //Get the read lock, so Wait() cant get write lock until we add the tokens
		select {
		case <-b.done:
			b.mu.RUnlock()
			return
		case <-t.C:
			newVal := atomic.AddInt64(&b.currCnt, b.tokenAdd)
			if newVal > b.bucketSize {
				// Withdraw excess
				atomic.AddInt64(&b.currCnt, b.bucketSize-newVal)
			}
			b.mu.RUnlock()
			runtime.Gosched()
		} //select
	} //for
}
