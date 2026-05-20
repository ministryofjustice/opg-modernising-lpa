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
