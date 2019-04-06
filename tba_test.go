package tba

import (
	"testing"
	"time"
	"runtime"
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

func TestAsk(t *testing.T) {
	t.Logf("With GOMAXPROCS=1")
	runtime.GOMAXPROCS(1)
	ask_test_helper(t)
	t.Logf("With GOMAXPROCS=2")
	runtime.GOMAXPROCS(2)
	ask_test_helper(t)
	t.Logf("With GOMAXPROCS=5")
	runtime.GOMAXPROCS(5)
	ask_test_helper(t)
}

func ask_test_helper(t *testing.T) {
	b1 := NewBucket(2, 1, time.Millisecond)
	defer b1.Stop()
	if b1.AskN(3) == true {
		t.Errorf("Can get 3 where should not")
	}
	if b1.AskN(333) == true {
		t.Errorf("Can get 333 where should not")
	}
	if b1.AskN(0) == false {
		t.Errorf("Can't get 0")
	}
	if b1.Ask() == true { //Empty bucket at the start
		t.Errorf("Cant get 1 where should not")
	}
	time.Sleep(time.Millisecond * 3)

	if b1.AskN(1) == false {
		t.Errorf("Can't get 1 (#1)")
	}
	if b1.Ask() == false {
		t.Errorf("Can't get 1 (#2)")
	}
	if b1.Ask() == true { //Limited
		t.Errorf("Cant get 1 where should not")
	}
	time.Sleep(time.Millisecond * 3)
	if b1.AskN(2) == false {
		t.Errorf("Can't get 2")
	}
	if b1.Ask() == true { //Limited
		t.Errorf("Cant get 1 where should not")
	}
	b1.Wait(2)
	if b1.Ask() == true { //Limited
		t.Errorf("Cant get 1 where should not")
	}
}
