// Package names provides functions for checking equality of names.
package names

import "strings"

// Equal returns true if the names provided should be considered equal.
//
// Prefer [identity.UserData.MatchName] when dealing with the results of an
// identity check.
func Equal(a, b string) bool {
	return strings.EqualFold(normalise(a), normalise(b))
}

// EqualDoubleBarrel returns true if the names provided should be considered
// equal. Unlike [Equal] it will consider the names "X-Y" and "Y-Z" equal, as
// they share the "Y" component.
func EqualDoubleBarrel(a, b string) bool {
	for ap := range strings.SplitSeq(normalise(a), "-") {
		for bp := range strings.SplitSeq(normalise(b), "-") {
			if strings.EqualFold(ap, bp) {
				return true
			}
		}
	}

	return false
}

// EqualFull returns true if the full names should be considered equal.
func EqualFull(a, b interface{ FullName() string }) bool {
	return Equal(a.FullName(), b.FullName())
}

var replacer = strings.NewReplacer(
	"’", "'",
	"‘", "'",
	"–", "-",
	"—", "-",
)

func normalise(s string) string {
	return replacer.Replace(s)
}
