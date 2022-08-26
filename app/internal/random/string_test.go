package random

import (
	"strings"
	"testing"
)

func TestStringIsExpectedLength(t *testing.T) {
	type test struct {
		length int
	}

	tests := []test{
		{length: 1},
		{length: 10},
		{length: 100},
		{length: 999},
	}

	for _, tc := range tests {
		got := String(tc.length)

		if len(got) != tc.length {
			t.Fatalf("not expected length: wanted %v, got: %v", tc.length, len(got))
		}
	}
}

func TestStringContainsExpectedChars(t *testing.T) {
	type test struct {
		charSet string
	}

	tests := []test{
		{charSet: "abcdefghijk"},
		{charSet: "!@$%&*()abc"},
		{charSet: "12345abcde!@$"},
	}

	for _, tc := range tests {
		SetCharset(tc.charSet)
		got := String(15)

		for _, c := range got {
			if !strings.Contains(tc.charSet, string(c)) {
				t.Fatalf("String contained unexpected characters. Character set was: %s, string was: %s", tc.charSet, got)
			}
		}
	}
}
