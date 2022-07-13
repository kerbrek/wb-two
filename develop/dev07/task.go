package main

/*
=== Or channel ===

Реализовать функцию, которая будет объединять один или более done каналов в single канал если
один из его составляющих каналов закроется.
Одним из вариантов было бы очевидно написать выражение при помощи select, которое бы реализовывало
эту связь, однако иногда неизестно общее число done каналов, с которыми вы работаете в рантайме.
В этом случае удобнее использовать вызов единственной функции, которая, приняв на вход один
или более or каналов, реализовывала весь функционал.

Определение функции:
var or func(channels ...<- chan interface{}) <- chan interface{}

Пример использования функции:
sig := func(after time.Duration) <- chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		time.Sleep(after)
}()
return c
}

start := time.Now()
<-or (
	sig(2*time.Hour),
	sig(5*time.Minute),
	sig(1*time.Second),
	sig(1*time.Hour),
	sig(1*time.Minute),
)

fmt.Printf(“fone after %v”, time.Since(start))
*/

import (
	"fmt"
	"sync"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	out := make(chan interface{}, len(channels))

	if len(channels) == 0 {
		close(out)
		return out
	}

	wg := &sync.WaitGroup{}
	merge := make(chan bool)
	first := make(chan bool, len(channels))

	go func() {
		<-first
		close(merge)
		wg.Wait()
		close(out)
	}()

	for _, c := range channels {
		wg.Add(1)

		go func(ch <-chan interface{}, closed chan<- bool) {
			defer wg.Done()

		OuterLoop:
			for {
				select {
				case _, ok := <-ch:
					if !ok {
						closed <- true
						break OuterLoop
					}
				case <-merge:
					break OuterLoop
				}
			}

			out <- <-ch
		}(c, first)
	}

	return out
}

func sig(after time.Duration) <-chan interface{} {
	c := make(chan interface{})

	go func() {
		defer close(c)
		time.Sleep(after)
	}()

	return c
}

func main() {
	start := time.Now()

	<-or()
	fmt.Printf("zero: fone after %v\n", time.Since(start))

	<-or(sig(1 * time.Second))
	fmt.Printf("one: fone after %v\n", time.Since(start))

	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("several: fone after %v\n", time.Since(start))

	n := 5
	cs := make([]<-chan interface{}, 0, n)
	for i := 1; i <= n; i++ {
		cs = append(cs, sig(time.Duration(i)*time.Second))
	}

	out := or(cs...)
	for i := 1; i <= n; i++ {
		<-out
		fmt.Printf("%d: fone after %v\n", i, time.Since(start))
	}

	<-out
}
