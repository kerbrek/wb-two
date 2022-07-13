package main

/*
=== Задача на распаковку ===

Создать Go функцию, осуществляющую примитивную распаковку строки, содержащую повторяющиеся
символы / руны, например:
	- "a4bc2d5e" => "aaaabccddddde"
	- "abcd" => "abcd"
	- "45" => "" (некорректная строка)
	- "" => ""
Дополнительное задание: поддержка escape - последовательностей
	- qwe\4\5 => qwe45 (*)
	- qwe\45 => qwe44444 (*)
	- qwe\\5 => qwe\\\\\ (*)

В случае если была передана некорректная строка функция должна возвращать ошибку. Написать unit-тесты.

Функция должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidString = errors.New("invalid string")

func unpack(s string) (string, error) {
	if s == "" {
		return s, nil
	}

	// Проверяем, что в начале не цифра.
	first, _ := utf8.DecodeRuneInString(s)
	if unicode.IsDigit(first) {
		return "", ErrInvalidString
	}

	// Проверяем, что в конце 0 или четное количество escape-символов.
	re := regexp.MustCompile(`\\+$`)
	match := re.FindString(s)
	if (utf8.RuneCountInString(match) % 2) != 0 {
		return "", ErrInvalidString
	}

	const (
		characters = iota
		digits
		escape
	)
	b := strings.Builder{}
	state := characters
	prevChar := ""
	num := ""

	// Добавляем пробел в конец строки, чтобы свитч отработал с последним символом
	// искомой строки.
	for _, r := range s + " " {
		switch state {
		case characters:
			if unicode.IsDigit(r) {
				state = digits
				num = string(r)
			} else if r == '\\' {
				state = escape
				b.WriteString(prevChar)
			} else {
				b.WriteString(prevChar)
				prevChar = string(r)
			}
		case digits:
			if unicode.IsDigit(r) {
				num += string(r)
			} else {
				count, _ := strconv.Atoi(num)
				b.WriteString(strings.Repeat(prevChar, count))
				num = ""

				if r == '\\' {
					state = escape
				} else {
					state = characters
					prevChar = string(r)
				}
			}
		case escape:
			state = characters
			prevChar = string(r)
		}
	}

	return b.String(), nil
}

func main() {
	str := `a4bc0\23d\\3e12`
	fmt.Printf("Original: %q\n", str)
	unpacked, err := unpack(str)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Printf("Unpacked: %q\n", unpacked)
	fmt.Printf("Expected: %q\n", `aaaab222d\\\eeeeeeeeeeee`)
}
