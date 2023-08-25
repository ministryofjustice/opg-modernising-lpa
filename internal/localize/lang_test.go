package localize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLangAbbreviation(t *testing.T) {
	type test struct {
		language string
		lang     Lang
		want     string
	}

	testCases := []test{
		{language: "English", lang: En, want: "en"},
		{language: "Welsh", lang: Cy, want: "cy"},
		{language: "Defaults to English with unsupported lang", lang: Lang(3), want: "en"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			a := tc.lang.String()
			assert.Equal(t, tc.want, a)
		})
	}
}
