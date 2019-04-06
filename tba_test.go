package tba

import (
	"testing"
	"time"
)

func TestBase(t *testing.T) {
	b1 := NewQPSLimit(10) //This must be 1 query eery 100ms
	defer b1.Stop()
	if b1.currCnt != 0 {
		t.Errorf("Invalid init count: %+v", b1)
	}
	if b1.tokenAdd != 1 || b1.period != time.Millisecond*100 {
		t.Errorf("Wrong bucket created: %+v", b1)
	}

	b2 := NewQPMLimit(60) //Must be 1 query every 1s
	defer b2.Stop()
	if b2.tokenAdd != 1 || b2.period != time.Second {
		t.Errorf("Wrong bucket created: %+v", b2)
	}
}

func TestNewQueryPerDuration(t *testing.T) {
	b1 := newQueryPerDuration(1, time.Second)
	defer b1.Stop()
	if b1.tokenAdd != 1 || b1.period != time.Second {
		t.Errorf("Wrong bucket created: %+v", b1)
	}

	b2 := newQueryPerDuration(1, time.Microsecond)
	defer b2.Stop()
	if b2.tokenAdd != 1000 || b2.period != time.Millisecond {
		t.Errorf("Wrong bucket created: %+v", b2)
	}
}
