package main

import (
	"reflect"
	"testing"
)

func TestGetAnagrams(t *testing.T) {
	cases := []struct {
		in   []string
		want map[string][]string
	}{
		{
			in: []string{"тяпка", "Пятак", "слиток", "листок", "пятка", "столик", "Столик", "анаграмма"},
			want: map[string][]string{
				"слиток": {"листок", "слиток", "столик"},
				"тяпка":  {"пятак", "пятка", "тяпка"},
			},
		},
		{
			in:   []string{},
			want: map[string][]string{},
		},
		{
			in:   []string{"тяпка"},
			want: map[string][]string{},
		},
		{
			in:   []string{"столик", "Столик"},
			want: map[string][]string{},
		},
	}

	for _, c := range cases {
		got := getAnagrams(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("getAnagrams(%v) \n\t  == %v; \n\twant %v", c.in, got, c.want)
		}
	}
}
