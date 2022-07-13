package main

/*
=== Утилита grep ===

Реализовать утилиту фильтрации (man grep)

Поддержать флаги:
-A - "after" печатать +N строк после совпадения
-B - "before" печатать +N строк до совпадения
-C - "context" (A+B) печатать ±N строк вокруг совпадения
-c - "count" (количество строк)
-i - "ignore-case" (игнорировать регистр)
-v - "invert" (вместо совпадения, исключать)
-F - "fixed", точное совпадение со строкой, не паттерн
-n - "line num", печатать номер строки

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

import (
	"bufio"
	"container/list"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	opts := new(options)
	writer := bufio.NewWriter(os.Stdout)
	if err := do(os.Stdin, writer, os.Args, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type options struct {
	after      int
	before     int
	context    int
	count      bool
	ignoreCase bool
	invert     bool
	fixed      bool
	lineNum    bool
	args       []string
	pattern    string
	filenames  []string
	multifile  bool
	re         *regexp.Regexp
}

func (opts *options) parseFlags(args []string) {
	var after uint
	var before uint
	var context uint
	flagset := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagset.UintVar(&after, "A", 0, "Print NUM lines of trailing context after matching lines.")
	flagset.UintVar(&before, "B", 0, "Print NUM lines of leading context before matching lines.")
	flagset.UintVar(&context, "C", 0, "Print NUM lines of output context.")
	flagset.BoolVar(&opts.count, "c", false, "Print  a count of matching lines for each input file.")
	flagset.BoolVar(&opts.ignoreCase, "i", false, "Ignore case distinctions in patterns and input data.")
	flagset.BoolVar(&opts.invert, "v", false, "Invert the sense of matching, to select non-matching lines.")
	flagset.BoolVar(&opts.fixed, "F", false, "Interpret PATTERNS as fixed strings, not regular expressions.")
	flagset.BoolVar(&opts.lineNum, "n", false, "Prefix each line of output with the 1-based line number within its input file.")
	flagset.Parse(args[1:])

	opts.after = int(after)
	opts.before = int(before)
	opts.context = int(context)
	opts.args = flagset.Args()
}

var ErrPatternRequired = errors.New("you must specify a pattern")

func (opts *options) validate() error {
	if len(opts.args) == 0 {
		return ErrPatternRequired
	}

	return nil
}

func (opts *options) complete() error {
	opts.pattern = opts.args[0]
	opts.filenames = opts.args[1:]

	if len(opts.filenames) > 1 {
		opts.multifile = true
	}

	if opts.after == 0 {
		opts.after = opts.context
	}

	if opts.before == 0 {
		opts.before = opts.context
	}

	re, err := createRegexp(opts.pattern, opts)
	if err != nil {
		return err
	}
	opts.re = re

	return nil
}

func createRegexp(pattern string, opts *options) (*regexp.Regexp, error) {
	if opts.fixed {
		pattern = regexp.QuoteMeta(pattern)
	}

	if opts.ignoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func do(in io.Reader, out *bufio.Writer, args []string, opts *options) error {
	opts.parseFlags(args)

	if err := opts.validate(); err != nil {
		return err
	}

	if err := opts.complete(); err != nil {
		return err
	}

	readers := make([]io.Reader, 0)
	if len(opts.filenames) == 0 {
		readers = append(readers, in)
	} else {
		for _, name := range opts.filenames {
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

	if err := doGrep(readers, out, opts); err != nil {
		return err
	}

	return nil
}

func doGrep(files []io.Reader, out *bufio.Writer, opts *options) error {
	if opts.count {
		return countLines(files, out, opts)
	}

	return findLines(files, out, opts)
}

func countLines(files []io.Reader, out *bufio.Writer, opts *options) error {
	for fileIdx, file := range files {
		counter := 0
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Bytes()
			matched := matchLine(line, opts)
			if matched {
				counter += 1
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		if err := writeCounter(counter, out, opts, fileIdx); err != nil {
			return err
		}
	}

	if err := out.Flush(); err != nil {
		return err
	}

	return nil
}

func writeCounter(counter int, out *bufio.Writer, opts *options, fileIdx int) error {
	sep := ':'
	b := strings.Builder{}

	if opts.multifile {
		filename := opts.filenames[fileIdx]
		if filename == "-" {
			filename = "(standard input)"
		}
		b.WriteString(filename)
		b.WriteRune(sep)
	}

	b.WriteString(strconv.Itoa(counter))
	b.WriteRune('\n')

	if _, err := out.WriteString(b.String()); err != nil {
		return err
	}

	return nil
}

type Line struct {
	num int
	val []byte
}

func findLines(files []io.Reader, out *bufio.Writer, opts *options) error {
	// Требуются ли межфайловые и межконтекстные разделители.
	doesNeedSep := opts.before > 0 || opts.after > 0
	prevFileHasMatches := false

	for fileIdx, file := range files {
		currFileHasMatches := false
		isFileSepPrinted := false
		// Используем список для хранения строк контекста, предшествующих
		// сматченной строке.
		beforeCtx := list.New()
		// Используем счетчик для хранения количества строк контекста, следующих
		// за сматченной строкой. Также с помощью значения -1 сигнализируем о том,
		// что надо вывести межконтекстный разделитель. Значения меньше -1 игнорируем.
		afterCtx := -2

		lineNum := 0
		lastWrittenLineNum := 0
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			lineNum += 1
			line := scanner.Bytes()
			matched := matchLine(line, opts)

			if matched {
				currFileHasMatches = true
				if doesNeedSep && prevFileHasMatches && !isFileSepPrinted {
					// Выводим межфайловый разделитель.
					if _, err := out.WriteString("--\n"); err != nil {
						return err
					}
					isFileSepPrinted = true
				}

				if doesNeedSep && afterCtx == -1 {
					// Выводим межконтекстный разделитель, если необходимо.
					if opts.before > 0 {
						firstCtxLine := beforeCtx.Front().Value.(Line)
						if firstCtxLine.num-lastWrittenLineNum > 1 {
							// Выводим разделитель только если номер первой строки контекста
							// не следует непосредственно за номером последней выведенной строки.
							// Например:
							// 14-last written line
							// --
							// 16-first ctx line
							if _, err := out.WriteString("--\n"); err != nil {
								return err
							}
						}
					} else {
						if _, err := out.WriteString("--\n"); err != nil {
							return err
						}
					}
				}

				if opts.before > 0 {
					// Выводим строки предшествующего контекста.
					for e := beforeCtx.Front(); e != nil; e = e.Next() {
						prevLine := e.Value.(Line)
						err := writeLine(prevLine, out, opts, false, fileIdx)
						if err != nil {
							return err
						}
					}
					// Опустошаем список строк предшествующего контекста.
					beforeCtx = list.New()
				}

				// Выводим сматченную строку.
				err := writeLine(Line{lineNum, line}, out, opts, true, fileIdx)
				if err != nil {
					return err
				}

				lastWrittenLineNum = lineNum
				// Устанавливаем счетчик строк последующего контекста.
				afterCtx = opts.after
			} else {
				if afterCtx > 0 {
					// Если счетчик строк последующего контекста больше нуля,
					// выводим текущую строку.
					err := writeLine(Line{lineNum, line}, out, opts, false, fileIdx)
					if err != nil {
						return err
					}

					lastWrittenLineNum = lineNum
					afterCtx -= 1
				} else {
					// Если счетчик строк последующего контекста меньше или равен нулю,
					// сохраняем текущую строку в список строк предшествующего контекста.
					if opts.before > 0 {
						beforeCtx.PushBack(Line{lineNum, line})

						if beforeCtx.Len() > opts.before {
							beforeCtx.Remove(beforeCtx.Front())
						}
					}

					if afterCtx == 0 {
						// Сигнализируем о том, что надо вывести межконтекстный разделитель.
						afterCtx -= 1
					}
				}
			}
		}

		if currFileHasMatches {
			prevFileHasMatches = true
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

func matchLine(line []byte, opts *options) bool {
	matched := opts.re.Match(line)

	if opts.invert {
		return !matched
	}

	return matched
}

func writeLine(line Line, out *bufio.Writer, opts *options, matched bool, fileIdx int) error {
	var sep rune
	if matched {
		sep = ':'
	} else {
		sep = '-'
	}

	b := strings.Builder{}

	if opts.multifile {
		filename := opts.filenames[fileIdx]
		if filename == "-" {
			filename = "(standard input)"
		}
		b.WriteString(filename)
		b.WriteRune(sep)
	}

	if opts.lineNum {
		b.WriteString(strconv.Itoa(line.num))
		b.WriteRune(sep)
	}

	b.Write(line.val)
	b.WriteRune('\n')

	if _, err := out.WriteString(b.String()); err != nil {
		return err
	}

	return nil
}
