package main

/*
=== Поиск анаграмм по словарю ===

Напишите функцию поиска всех множеств анаграмм по словарю.
Например:
'пятак', 'пятка' и 'тяпка' - принадлежат одному множеству,
'листок', 'слиток' и 'столик' - другому.

Входные данные для функции: ссылка на массив - каждый элемент которого - слово на русском языке
в кодировке utf8.
Выходные данные: Ссылка на мапу множеств анаграмм.
Ключ - первое встретившееся в словаре слово из множества
Значение - ссылка на массив, каждый элемент которого, слово из множества. Массив должен быть
отсортирован по возрастанию.
Множества из одного элемента не должны попасть в результат.
Все слова должны быть приведены к нижнему регистру.
В результате каждое слово должно встречаться только один раз.

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

import (
	"fmt"
	"sort"
	"strings"
)

func sortChars(s string) string {
	rs := []rune(s)
	sort.Slice(rs, func(i int, j int) bool { return rs[i] < rs[j] })
	return string(rs)
}

func getAnagrams(words []string) map[string][]string {
	if len(words) < 2 {
		return make(map[string][]string)
	}

	// Коллекция анаграмм вида: ключ анаграммы -> множество слов.
	// Пример: акптя -> {пятак, тяпка}
	anagrams := make(map[string]map[string]struct{})
	// Маппинг ключа анаграммы на первое подходящее слово.
	mapping := make(map[string]string)

	for _, word := range words {
		word = strings.ToLower(word)
		key := sortChars(word)
		if _, ok := mapping[key]; !ok {
			mapping[key] = word
		}

		if _, ok := anagrams[key]; !ok {
			anagrams[key] = make(map[string]struct{})
		}

		anagrams[key][word] = struct{}{}
	}

	res := make(map[string][]string)
	for key, firstEncounteredWord := range mapping {
		if len(anagrams[key]) == 1 {
			continue
		}

		sequence := make([]string, 0, len(anagrams[key]))
		for word := range anagrams[key] {
			sequence = append(sequence, word)
		}

		sort.Strings(sequence)
		res[firstEncounteredWord] = sequence
	}

	return res
}

func main() {
	words := []string{"тяпка", "Пятак", "слиток", "листок", "пятка", "столик", "Столик", "анаграмма"}
	res := getAnagrams(words)
	fmt.Println(res)

	words = []string{}
	res = getAnagrams(words)
	fmt.Println(res)

	words = []string{"тяпка"}
	res = getAnagrams(words)
	fmt.Println(res)

	words = []string{"столик", "Столик"}
	res = getAnagrams(words)
	fmt.Println(res)
}
