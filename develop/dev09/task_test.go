package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/v2/recorder"
)

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}

func TestWget(t *testing.T) {
	domain := "www.restapitutorial.com"
	website := "https://" + domain + "/"

	args := []string{"test-wget", "-r", website}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		saveDirPath := "./" + domain
		testDirPath := "./testdata/" + domain
		os.RemoveAll(saveDirPath)
		defer os.RemoveAll(saveDirPath)

		r, recErr := recorder.New("fixtures/" + domain)
		if recErr != nil {
			t.Fatal(recErr)
		}
		defer r.Stop()

		opts := new(options)
		opts.client = &http.Client{
			Transport: r,
		}
		if err := do(args, opts); err != nil {
			t.Fatal(err)
		}

		expectedSize, err := dirSize(testDirPath)
		if err != nil {
			t.Fatal(err)
		}

		actualSize, err := dirSize(saveDirPath)
		if err != nil {
			t.Fatal(err)
		}

		if expectedSize != actualSize {
			t.Fatal("Dirs sizes are not equal")
		}

		expectedEntries, err := os.ReadDir(testDirPath)
		if err != nil {
			t.Fatal(err)
		}

		actualEntries, err := os.ReadDir(saveDirPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(expectedEntries) != len(actualEntries) {
			t.Fatal("Number of entries are not equal")
		}

		for i := 0; i < len(expectedEntries); i++ {
			expectedName := expectedEntries[i].Name()
			actualName := actualEntries[i].Name()
			if expectedName != actualName {
				t.Fatal("Entries names are not equal")
			}
		}
	})

	args = []string{"test-wget", website}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		filename := "index.html"
		saveFilePath := "./" + filename
		testFilePath := "./testdata/" + filename
		os.Remove(saveFilePath)
		defer os.Remove(saveFilePath)

		r, recErr := recorder.New("fixtures/" + domain)
		if recErr != nil {
			t.Fatal(recErr)
		}
		defer r.Stop()

		opts := new(options)
		opts.client = &http.Client{
			Transport: r,
		}
		if err := do(args, opts); err != nil {
			t.Fatal(err)
		}

		expectedFI, err := os.Stat(testFilePath)
		if err != nil {
			t.Fatal(err)
		}

		actualFI, err := os.Stat(saveFilePath)
		if err != nil {
			t.Fatal(err)
		}

		if expectedFI.Size() != actualFI.Size() {
			t.Fatal("Files sizes are not equal")
		}
	})
}

func TestWgetErrors(t *testing.T) {
	args := []string{"test-wget"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		opts := new(options)
		opts.client = http.DefaultClient
		err := do(args, opts)
		if err != ErrMissingURL {
			t.Fatal("Not equal")
		}
	})
}
