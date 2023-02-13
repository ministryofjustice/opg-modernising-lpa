package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestLangRedirect(t *testing.T) {
	testCases := map[localize.Lang]string{
		localize.En: "/dashboard",
		localize.Cy: "/cy/dashboard",
	}

	for lang, url := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, nil, "/dashboard")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestLangRedirectWhenLpaRoute(t *testing.T) {
	testCases := map[localize.Lang]string{
		localize.En: "/lpa/lpa-id/somewhere",
		localize.Cy: "/cy/lpa/lpa-id/somewhere",
	}

	for lang, url := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, nil, "/somewhere")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestLangRedirectWhenCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected string
	}{
		"nil": {
			lpa:      nil,
			expected: Paths.HowToConfirmYourIdentityAndSign,
		},
		"allowed": {
			lpa:      &Lpa{Tasks: Tasks{PayForLpa: TaskCompleted}},
			expected: Paths.HowToConfirmYourIdentityAndSign,
		},
		"not allowed": {
			lpa:      &Lpa{},
			expected: Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: localize.En, LpaID: "lpa-id"}.Redirect(w, r, tc.lpa, Paths.HowToConfirmYourIdentityAndSign)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.expected, resp.Header.Get("Location"))
		})
	}
}

func TestLangBuildUrl(t *testing.T) {
	type test struct {
		language string
		lang     localize.Lang
		url      string
		want     string
	}

	testCases := []test{
		{language: "English", lang: localize.En, url: "/example.org", want: "/lpa/123/example.org"},
		{language: "Welsh", lang: localize.Cy, url: "/example.org", want: "/cy/lpa/123/example.org"},
		{language: "Other", lang: localize.Lang(3), url: "/example.org", want: "/lpa/123/example.org"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			builtUrl := AppData{Lang: tc.lang, LpaID: "123"}.BuildUrl(tc.url)
			assert.Equal(t, tc.want, builtUrl)
		})
	}
}
