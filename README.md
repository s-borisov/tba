Token Bucket Algorithm
========
The read about this https://en.wikipedia.org/wiki/Token_bucket

## Features

- Simple and clean API
- Load optimized
- Smooth limiting

## Usage example

```
import "github.com/PSIAlt/tba"

##########
Example 1:
b1 := tba.NewQPSLimit(10) //Allow to get 1 token(query) every 100ms
defer b1.Stop()
if ! b1.Ask() {
	// Not allowed
	return ErrLimit
}

##########
Example 2:
b1 := tba.NewQPMLimit(60) //Allow to get 1 token(query) every 1s
defer b1.Stop()
if ! b1.Ask() {
	// Not allowed
	return ErrLimit
}

##########
Example 3:
b1 := tba.NewQPSLimit(100000) //Will atomatically allow 100 requests every 1ms
defer b1.Stop()
if ! b1.Ask() {
	// Not allowed
	return ErrLimit
}

##########
Example 4(advanced):
// Allow to get 10 tokens every 1ms, with bursts up to 20000
b1 := tba.NewBucket(20000, 10, time.Millisecond)
defer b1.Stop()
if ! b1.Ask() {
	// Not allowed
	return ErrLimit
}

##########
Example 5(advanced):
// Account user connection bandwidth, limit it at 10kbit/s rate
// Allow to get 10 tokens(bits) every 1ms, with bursts up to 20kbit/s
b1 := tba.NewBucket(20000, 10, time.Millisecond)
defer b1.Stop()
b1.Wait( pkt.bytes*8 ) //Wait availability
forwardPacket(pkt)


```
