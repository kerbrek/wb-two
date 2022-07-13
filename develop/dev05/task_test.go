package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGrep(t *testing.T) {
	file, err := os.Open("testdata/data.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	args := []string{"test-grep", "pet"}
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

	args = []string{"test-grep", "-n", "-C", "1", "pet", "-", "testdata/data.txt"}
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

	args = []string{"test-grep", "-i", "-n", "-A", "1", "-B", "4", "PeT", "testdata/data.txt"}
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

	args = []string{"test-grep", "-v", "pet", "testdata/data.txt"}
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

	args = []string{"test-grep", "-F", ".*", "testdata/data.txt"}
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

	args = []string{"test-grep", "-c", "pet", "testdata/data.txt", "testdata/empty.txt"}
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

	args = []string{"test-grep", "pet", "testdata/empty.txt"}
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

func TestGrepErrors(t *testing.T) {
	args := []string{"test-grep"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		writer := bufio.NewWriter(os.Stdout)
		err := do(os.Stdin, writer, args, opts)
		if err != ErrPatternRequired {
			t.Fatal("Not equal")
		}
	})
}
