package main

/*
=== Утилита sort ===

Отсортировать строки (man sort)
Основное

Поддержать ключи

-k — указание колонки для сортировки
-n — сортировать по числовому значению
-r — сортировать в обратном порядке
-u — не выводить повторяющиеся строки

Дополнительное

Поддержать ключи

-M — сортировать по названию месяца
-b — игнорировать хвостовые пробелы
-c — проверять отсортированы ли данные
-h — сортировать по числовому значению с учётом суффиксов

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
	"sort"
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
	field   int
	numeric bool
	reverse bool
	unique  bool
	args    []string
}

func (opts *options) parseFlags(args []string) {
	flagset := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagset.IntVar(&opts.field, "k", 1, "sort via specified field")
	flagset.BoolVar(&opts.numeric, "n", false, "compare according to string numerical value")
	flagset.BoolVar(&opts.reverse, "r", false, "reverse the result of comparisons")
	flagset.BoolVar(&opts.unique, "u", false, "output only unique lines")
	flagset.Parse(args[1:])

	opts.args = flagset.Args()
}

var ErrInvalidFieldValue = errors.New("invalid field number")

func (opts *options) validate() error {
	if opts.field < 1 {
		return ErrInvalidFieldValue
	}

	return nil
}

func (opts *options) complete() {
	// Уменьшаем номер поля на единицу, чтобы проще было работать с индексами.
	opts.field -= 1
}

func do(in io.Reader, out *bufio.Writer, args []string, opts *options) error {
	opts.parseFlags(args)

	if err := opts.validate(); err != nil {
		return err
	}

	opts.complete()

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

	if err := doSort(readers, out, opts); err != nil {
		return err
	}

	return nil
}

func doSort(files []io.Reader, out *bufio.Writer, opts *options) error {
	lines, err := readLines(files)
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return nil
	}

	sortLines(lines, opts)

	if err := writeLines(lines, out, opts); err != nil {
		return err
	}

	return nil
}

func readLines(files []io.Reader) ([]string, error) {
	lines := make([]string, 0)

	for _, file := range files {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return lines, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sortLines(lines []string, opts *options) {
	sort.Slice(lines, func(i, j int) bool {
		if opts.reverse {
			i, j = j, i
		}

		// Извлекаем из строк поля.
		iFields := strings.Fields(lines[i])
		jFields := strings.Fields(lines[j])

		// Сравниваем указанные поля.
		cmp := compareField(iFields, jFields, opts.field, opts)
		if cmp == -1 {
			return true
		} else if cmp == 1 {
			return false
		}

		// Если сортировка не числовая, сравниваем все поля следующие за указанным.
		if !opts.numeric {
			maxLen := max(len(iFields), len(jFields))
			for f := opts.field + 1; f < maxLen; f++ {
				cmp := compareField(iFields, jFields, f, opts)
				if cmp == -1 {
					return true
				} else if cmp == 1 {
					return false
				}
			}
		}

		// На этот момент указанные поля либо равны либо отсутствуют. Начинаем
		// лексикографическую проверку по порядку с первого поля до меньшего
		// количества полей.
		minLen := min(len(iFields), len(jFields))
		for f := 0; f < minLen; f++ {
			// Пропускаем проверку полей, которые уже сравнивали.
			if !opts.numeric && f >= opts.field {
				break
			}

			if iFields[f] < jFields[f] {
				return true
			}
			if iFields[f] > jFields[f] {
				return false
			}
		}

		// На этот момент поля равны. Сравниваем по количеству полей.
		if len(iFields) < len(jFields) {
			return true
		}
		if len(iFields) > len(jFields) {
			return false
		}

		// Количество полей равно. Сравниваем полностью строки.
		return lines[i] < lines[j]
	})
}

func compareField(i, j []string, field int, opts *options) int {
	iHasField := len(i) >= field+1
	jHasField := len(j) >= field+1

	if opts.numeric {
		// Если поля отсутствуют используем дефолтное значение: ноль.
		iNum, jNum := float64(0), float64(0)
		if iHasField {
			iNum = parseNum(i[field])
		}

		if jHasField {
			jNum = parseNum(j[field])
		}

		if iNum < jNum {
			return -1
		} else if iNum > jNum {
			return 1
		} else {
			return 0
		}
	}

	if !iHasField && !jHasField {
		return 0
	}

	if !iHasField && jHasField {
		return -1
	}

	if iHasField && !jHasField {
		return 1
	}

	if i[field] < j[field] {
		return -1
	} else if i[field] > j[field] {
		return 1
	} else {
		return 0
	}
}

func parseNum(num string) float64 {
	f, err := strconv.ParseFloat(num, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return f
}

func writeLines(lines []string, out *bufio.Writer, opts *options) error {
	switch opts.unique {
	case true:
		prevLine := lines[0]
		if err := writeLine(prevLine, out); err != nil {
			return err
		}

		for _, line := range lines[1:] {
			if line == prevLine {
				continue
			}

			if err := writeLine(line, out); err != nil {
				return err
			}

			prevLine = line
		}
	case false:
		for _, line := range lines {
			if err := writeLine(line, out); err != nil {
				return err
			}
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
