package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
)

func server(wg *sync.WaitGroup, close bool) {
	defer wg.Done()
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	if close {
		return
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write(line)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func TestTelnet(t *testing.T) {
	args := []string{"test-telnet", "localhost", "8080"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go server(wg, false)

		opts := new(options)
		outbuf := &bytes.Buffer{}

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}

		msg := "Hello World\n"
		w.WriteString(msg)

		err = do(r, outbuf, args, opts)
		if err != ErrConnectionClosed {
			t.Fatal("Not equal")
		}

		wg.Wait()

		expected := "connected to 127.0.0.1:8080\n" + msg
		actual := outbuf.String()
		if expected != actual {
			fmt.Println(expected)
			fmt.Println("--")
			fmt.Println(actual)
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-telnet", "localhost", "8080"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go server(wg, true)

		opts := new(options)
		outbuf := &bytes.Buffer{}

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}

		w.Close()

		err = do(r, outbuf, args, opts)
		if err != nil {
			t.Fatal("Not equal")
		}

		wg.Wait()
	})
}

func TestTelnetErrors(t *testing.T) {
	args := []string{"test-telnet", "--timeout=1s", "nonexistenthost", "8080"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		inbuf := &bytes.Buffer{}
		outbuf := &bytes.Buffer{}

		err := do(inbuf, outbuf, args, opts)
		if err != ErrConnectionTimedOut {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-telnet"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		inbuf := &bytes.Buffer{}
		outbuf := &bytes.Buffer{}

		err := do(inbuf, outbuf, args, opts)
		if err != ErrArgsRequired {
			t.Fatal("Not equal")
		}
	})
}
