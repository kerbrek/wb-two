package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCut(t *testing.T) {
	file, err := os.Open("testdata/tabbeddata.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	args := []string{"test-cut", "-f", "2"}
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

	args = []string{"test-cut", "-f", "2,4", "-s", "-", "testdata/tabbeddata.txt"}
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

	args = []string{"test-cut", "-f", "2-4", "-d", " ", "testdata/spaceddata.txt"}
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

	args = []string{"test-cut", "-f", "3-", "-s", "-d", " ", "testdata/spaceddata.txt"}
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

	args = []string{"test-cut", "-f", "-3", "testdata/tabbeddata.txt"}
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

	args = []string{"test-cut", "-f", "2", "-d", "", "testdata/tabbeddata.txt"}
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

	args = []string{"test-cut", "-f", "2,3", "testdata/empty.txt"}
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

func TestCutErrors(t *testing.T) {
	args := []string{"test-cut"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrFieldsListRequired {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-cut", "-f", "1", "-d", "  "}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrInvalidDelimiter {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-cut", "-f", "0"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrFieldZero {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-cut", "-f", "-"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrInvalidRange {
			t.Fatal("Not equal")
		}
	})

	args = []string{"test-cut", "-f", "3-2"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrDecreasingRange {
			t.Fatal("Not equal")
		}
	})
}
