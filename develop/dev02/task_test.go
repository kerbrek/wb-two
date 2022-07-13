package main

import (
	"testing"
)

type result struct {
	val string
	err error
}

func TestUnpack(t *testing.T) {
	cases := []struct {
		in   string
		want result
	}{
		{"42", result{"", ErrInvalidString}},
		{`abcd\`, result{"", ErrInvalidString}},
		{"", result{"", nil}},
		{"a0", result{"", nil}},
		{"abcd", result{`abcd`, nil}},
		{"a4bc2d5e", result{"aaaabccddddde", nil}},
		{"a11", result{"aaaaaaaaaaa", nil}},
		{`qwe\4\5`, result{"qwe45", nil}},
		{`qwe\45`, result{"qwe44444", nil}},
		{`qwe\\5`, result{`qwe\\\\\`, nil}},
		{`qwe5\\`, result{`qweeeee\`, nil}},
	}

	for _, c := range cases {
		got, err := unpack(c.in)
		if got != c.want.val || err != c.want.err {
			t.Errorf(`unpack(%q) == %q, %v; want %q, %v`, c.in, got, err, c.want.val, c.want.err)
		}
	}
}
