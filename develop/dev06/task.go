package main

/*
=== Утилита cut ===

Принимает STDIN, разбивает по разделителю (TAB) на колонки, выводит запрошенные

Поддержать флаги:
-f - "fields" - выбрать поля (колонки)
-d - "delimiter" - использовать другой разделитель
-s - "separated" - только строки с разделителем

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func main() {
	opts := new(options)
	writer := bufio.NewWriter(os.Stdout)
	if err := do(os.Stdin, writer, os.Args, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type fieldRange struct {
	start uint64
	end   uint64
}

type options struct {
	fields      string
	delimiter   string
	separated   bool
	args        []string
	fieldSet    map[uint64]bool
	fieldRanges []fieldRange
}

func (opts *options) parseFlags(args []string) {
	flagset := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagset.StringVar(&opts.fields, "f", "", "select  only  these fields")
	flagset.StringVar(&opts.delimiter, "d", "\t", "use specified string instead of TAB for field delimiter")
	flagset.BoolVar(&opts.separated, "s", false, "do not print lines not containing delimiters")
	flagset.Parse(args[1:])

	opts.args = flagset.Args()
}

var ErrInvalidDelimiter = errors.New("the delimiter must be a single character")
var ErrFieldsListRequired = errors.New("you must specify a list of fields")

func (opts *options) validate() error {
	if utf8.RuneCountInString(opts.delimiter) > 1 {
		return ErrInvalidDelimiter
	}

	if opts.fields == "" {
		return ErrFieldsListRequired
	}

	return nil
}

func (opts *options) complete() error {
	opts.fieldSet = make(map[uint64]bool)

	if err := parseFields(opts); err != nil {
		return err
	}

	return nil
}

var ErrFieldZero = errors.New("fields are numbered from 1")
var ErrFieldTooLarge = errors.New("field number is too large")
var ErrInvalidFieldValue = errors.New("invalid field value")
var ErrInvalidRange = errors.New("invalid field range")
var ErrDecreasingRange = errors.New("invalid decreasing range")

func parseFields(opts *options) error {
	fields := strings.Split(opts.fields, ",")
	for _, field := range fields {
		if strings.ContainsRune(field, '-') {
			start, end, err := parseRange(field, opts)
			if err != nil {
				return err
			}

			// Если диапазон небольшой - сохраняем его значения в сет, в противном
			// случае добавляем диапазон в слайс.
			if (end - start) < 1000 {
				for i := start; i <= end; i++ {
					opts.fieldSet[i] = true
				}
			} else {
				opts.fieldRanges = append(opts.fieldRanges, fieldRange{start, end})
			}
		} else {
			num, err := parseNum(field)
			if err != nil {
				return err
			}

			opts.fieldSet[num] = true
		}
	}

	return nil
}

func parseRange(field string, opts *options) (uint64, uint64, error) {
	before, after, _ := strings.Cut(field, "-")
	if before == "" && after == "" {
		return 0, 0, ErrInvalidRange
	}

	start, end := uint64(1), uint64(1)

	if before == "" {
		start = 1
	} else {
		num, err := parseNum(before)
		if err != nil {
			return 0, 0, err
		}

		start = num
	}

	if after == "" {
		end = math.MaxUint64
	} else {
		num, err := parseNum(after)
		if err != nil {
			return 0, 0, err
		}

		end = num
	}

	if start > end {
		return 0, 0, ErrDecreasingRange
	}

	return start, end, nil
}

func parseNum(num string) (uint64, error) {
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		if err.Error() == "value out of range" {
			return 0, ErrFieldTooLarge
		}

		return 0, ErrInvalidFieldValue
	}

	if n == 0 {
		return 0, ErrFieldZero
	}

	if n == math.MaxUint64 {
		return 0, ErrFieldTooLarge
	}

	return n, nil
}

func do(in io.Reader, out *bufio.Writer, args []string, opts *options) error {
	opts.parseFlags(args)

	if err := opts.validate(); err != nil {
		return err
	}

	if err := opts.complete(); err != nil {
		return err
	}

	if opts.delimiter == "" && opts.separated {
		return nil
	}

	readers := make([]io.Reader, 0)
	if len(opts.args) == 0 {
		readers = append(readers, in)
	} else {
		for _, name := range opts.args {
			if name == "-" {
				readers = append(readers, in)
				continue
			}

			file, err := os.Open(name)
			if err != nil {
				return err
			}
			defer file.Close()

			readers = append(readers, file)
		}
	}

	if err := doCut(readers, out, opts); err != nil {
		return err
	}

	return nil
}

func doCut(files []io.Reader, out *bufio.Writer, opts *options) error {
	for _, file := range files {
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			if opts.delimiter == "" {
				if err := writeLine(line, out); err != nil {
					return err
				}
				continue
			}

			parts := strings.Split(line, opts.delimiter)

			if opts.separated && len(parts) == 1 {
				continue
			}

			if len(parts) == 1 {
				if err := writeLine(line, out); err != nil {
					return err
				}
				continue
			}

			if err := writeParts(parts, out, opts); err != nil {
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	if err := out.Flush(); err != nil {
		return err
	}

	return nil
}

func writeLine(line string, out *bufio.Writer) error {
	if _, err := out.WriteString(line); err != nil {
		return err
	}

	if _, err := out.WriteRune('\n'); err != nil {
		return err
	}

	return nil
}

func writeParts(parts []string, out *bufio.Writer, opts *options) error {
	matchedParts := make([]string, 0)

	for i := uint64(0); i < uint64(len(parts)); i++ {
		field := i + 1
		matched := false

		if opts.fieldSet[field] {
			matched = true
		}

		for _, r := range opts.fieldRanges {
			if field >= r.start && field <= r.end {
				matched = true
				break
			}
		}

		if matched {
			matchedParts = append(matchedParts, parts[i])
		}
	}

	line := strings.Join(matchedParts, opts.delimiter)
	if err := writeLine(line, out); err != nil {
		return err
	}

	return nil
}
