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
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			a := tc.lang.String()
			assert.Equal(t, tc.want, a)
		})
	}
}

func TestLangURL(t *testing.T) {
	testCases := map[string]struct {
		lang Lang
		url  string
		want string
	}{
		"english":        {lang: En, url: "/example.org", want: "/example.org"},
		"welsh":          {lang: Cy, url: "/example.org", want: "/cy/example.org"},
		"other language": {lang: Lang(3), url: "/example.org", want: "/example.org"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			builtUrl := tc.lang.URL(tc.url)
			assert.Equal(t, tc.want, builtUrl)
		})
	}
}
