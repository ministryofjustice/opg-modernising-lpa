package localize

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func LowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func enPossessive(s string) string {
	format := "%s’s"

	if strings.HasSuffix(s, "s") {
		format = "%s’"
	}

	return fmt.Sprintf(format, s)
}

func cyAac(s string) string {
	r, _ := utf8.DecodeRuneInString(s)
	switch r {
	case 'A', 'E', 'I', 'O', 'U', 'Y':
		return "ac " + s
	default:
		return "a " + s
	}
}

func cySoftMutate(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	switch r {
	case 'C':
		return "G" + s[n:]
	case 'c':
		return "g" + s[n:]
	case 'P':
		return "B" + s[n:]
	case 'p':
		return "b" + s[n:]
	case 'T':
		return "D" + s[n:]
	case 't':
		return "d" + s[n:]
	case 'G', 'g':
		return s[n:]
	case 'B':
		return "F" + s[n:]
	case 'b':
		return "f" + s[n:]
	case 'D':
		return "Dd" + s[n:]
	case 'd':
		return "dd" + s[n:]
	case 'M':
		return "F" + s[n:]
	case 'm':
		return "f" + s[n:]
	case 'R':
		if r2, n2 := utf8.DecodeRuneInString(s[n:]); r2 == 'h' {
			return "R" + s[n+n2:]
		}
	case 'r':
		if r2, n2 := utf8.DecodeRuneInString(s[n:]); r2 == 'h' {
			return "r" + s[n+n2:]
		}
	case 'L':
		if r2, n2 := utf8.DecodeRuneInString(s[n:]); r2 == 'l' {
			return "L" + s[n+n2:]
		}
	case 'l':
		if r2, n2 := utf8.DecodeRuneInString(s[n:]); r2 == 'l' {
			return "l" + s[n+n2:]
		}
	}

	return s
}
