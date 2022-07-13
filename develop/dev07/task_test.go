package main

import (
	"testing"
	"time"
)

func sigTest(after time.Duration, num int) <-chan interface{} {
	c := make(chan interface{})

	go func() {
		defer close(c)
		time.Sleep(after)
		c <- num
	}()

	return c
}

func TestOr(t *testing.T) {
	n := 5
	cs := make([]<-chan interface{}, 0, n)
	for i := 0; i < n; i++ {
		cs = append(cs, sigTest(time.Duration(i)*50*time.Millisecond, i))
	}

	sum := 0
	out := or(cs...)
	for i := 0; i < n; i++ {
		num := <-out
		if i == 0 {
			continue
		}
		sum += num.(int)
	}

	if sum != 10 {
		t.Error("sum must be equal to 10")
	}

	zeroVal := <-out
	if zeroVal != nil {
		t.Error("zeroVal must be equal to nil")
	}
}
