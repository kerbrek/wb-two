package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSort(t *testing.T) {
	file, err := os.Open("testdata/data.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	args := []string{"test-sort"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		file.Seek(0, 0)
		opts := new(options)
		buf := &bytes.Buffer{}
		writer := bufio.NewWriter(buf)
		if err := do(file, writer, args, opts); err != nil {
			t.Fatal(err)
		}

		filename := "testdata/" + strings.ReplaceAll(strings.Join(args, "_"), "/", ".")
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-sort", "-r", "-", "testdata/data.txt"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		file.Seek(0, 0)
		opts := new(options)
		buf := &bytes.Buffer{}
		writer := bufio.NewWriter(buf)
		if err := do(file, writer, args, opts); err != nil {
			t.Fatal(err)
		}

		filename := "testdata/" + strings.ReplaceAll(strings.Join(args, "_"), "/", ".")
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-sort", "-u", "testdata/data.txt", "testdata/data.txt"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		buf := &bytes.Buffer{}
		writer := bufio.NewWriter(buf)
		if err := do(os.Stdin, writer, args, opts); err != nil {
			t.Fatal(err)
		}

		filename := "testdata/" + strings.ReplaceAll(strings.Join(args, "_"), "/", ".")
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-sort", "-k", "2", "testdata/data.txt"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		buf := &bytes.Buffer{}
		writer := bufio.NewWriter(buf)
		if err := do(os.Stdin, writer, args, opts); err != nil {
			t.Fatal(err)
		}

		filename := "testdata/" + strings.ReplaceAll(strings.Join(args, "_"), "/", ".")
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-sort", "-k", "3", "-n", "testdata/data.txt"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		buf := &bytes.Buffer{}
		writer := bufio.NewWriter(buf)
		if err := do(os.Stdin, writer, args, opts); err != nil {
			t.Fatal(err)
		}

		filename := "testdata/" + strings.ReplaceAll(strings.Join(args, "_"), "/", ".")
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-sort", "testdata/empty.txt"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		buf := &bytes.Buffer{}
		writer := bufio.NewWriter(buf)
		if err := do(os.Stdin, writer, args, opts); err != nil {
			t.Fatal(err)
		}

		expected := []byte{}
		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Fatal("Not equal")
		}
	})
}

func TestSortErrors(t *testing.T) {
	args := []string{"test-sort", "-k", "0"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrInvalidFieldValue {
			t.Fatal("Not equal")
		}
	})
}
