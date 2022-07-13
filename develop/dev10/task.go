package main

/*
=== Утилита telnet ===

Реализовать примитивный telnet клиент:
Примеры вызовов:
go-telnet --timeout=10s host port
go-telnet mysite.ru 8080
go-telnet --timeout=3s 1.1.1.1 123

Программа должна подключаться к указанному хосту (ip или доменное имя) и порту по протоколу TCP.
После подключения STDIN программы должен записываться в сокет, а данные полученные и сокета должны выводиться в STDOUT
Опционально в программу можно передать таймаут на подключение к серверу (через аргумент --timeout, по умолчанию 10s).

При нажатии Ctrl+D программа должна закрывать сокет и завершаться. Если сокет закрывается со стороны сервера,
программа должна также завершаться.
При подключении к несуществующему сервер, программа должна завершаться через timeout.
*/

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	opts := new(options)
	if err := do(os.Stdin, os.Stdout, os.Args, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type options struct {
	timeout time.Duration
	args    []string
	host    string
	port    int
	addr    string
}

func (opts *options) parseFlags(args []string) {
	flagset := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagset.DurationVar(&opts.timeout, "timeout", 10*time.Second, "connection timeout")
	flagset.Parse(args[1:])

	opts.args = flagset.Args()
}

var ErrArgsRequired = errors.New("you must specify exactly two arguments: host and port")

func (opts *options) validate() error {
	if len(opts.args) != 2 {
		return ErrArgsRequired
	}

	return nil
}

func (opts *options) complete() error {
	opts.host = opts.args[0]
	port, err := strconv.ParseUint(opts.args[1], 10, 16)
	if err != nil {
		return err
	}

	opts.port = int(port)
	opts.addr = fmt.Sprintf("%s:%d", opts.host, opts.port)

	return nil
}

func do(in io.Reader, out io.Writer, args []string, opts *options) error {
	opts.parseFlags(args)

	if err := opts.validate(); err != nil {
		return err
	}

	if err := opts.complete(); err != nil {
		return err
	}

	return doTelnet(in, out, opts)
}

var ErrConnectionClosed = errors.New("connection closed by foreign host")

func doTelnet(in io.Reader, out io.Writer, opts *options) error {
	conn, err := connect("tcp", opts.addr, opts.timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Fprintln(out, "connected to", conn.RemoteAddr())

	status := make(chan error, 2)

	// Читаем из соединения, пишем в stdout.
	go rw(conn, out, status, ErrConnectionClosed)

	// Читаем из stdin, пишем в соединение.
	go rw(in, conn, status, nil)

	return <-status
}

var ErrConnectionTimedOut = errors.New("connection timed out")

func connect(network string, address string, timeout time.Duration) (net.Conn, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
	for {
		select {
		case <-timeoutExceeded:
			return nil, ErrConnectionTimedOut
		case <-ticker.C:
			conn, err := net.DialTimeout(network, address, timeout)
			if err == nil {
				return conn, nil
			}
		}
	}
}

func rw(in io.Reader, out io.Writer, status chan<- error, err error) {
	reader := bufio.NewReader(in)
	for {
		input, rErr := reader.ReadBytes('\n')
		if rErr != nil {
			if rErr == io.EOF {
				status <- err
				return
			}

			status <- rErr
			return
		}

		if _, wErr := out.Write(input); wErr != nil {
			status <- wErr
			return
		}
	}
}
