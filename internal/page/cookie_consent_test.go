package page

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCookieConsent(t *testing.T) {
	for _, consent := range []string{"accept", "reject"} {
		t.Run(consent, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("cookies="+consent))
			r.Header.Add("Content-Type", FormUrlEncoded)

			CookieConsent()(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, PathStart.Format(), resp.Header.Get("Location"))

			cookies := resp.Cookies()
			if assert.Len(t, cookies, 1) {
				cookie := cookies[0]

				assert.Equal(t, "cookies-consent", cookie.Name)
				assert.Equal(t, consent, cookie.Value)
				assert.Equal(t, 365*24*60*60, cookie.MaxAge)
				assert.Equal(t, "/", cookie.Path)
			}
		})
	}
}

func TestCookieConsentRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("cookies-redirect=/here&cookies=accept"))
	r.Header.Add("Content-Type", FormUrlEncoded)

	CookieConsent()(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/here", resp.Header.Get("Location"))

	cookies := resp.Cookies()
	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]

		assert.Equal(t, "cookies-consent", cookie.Name)
		assert.Equal(t, "accept", cookie.Value)
		assert.Equal(t, 365*24*60*60, cookie.MaxAge)
		assert.Equal(t, "/", cookie.Path)
	}
}

func TestCookieConsentBadRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("cookies-redirect=http://google&cookies=accept"))
	r.Header.Add("Content-Type", FormUrlEncoded)

	CookieConsent()(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, PathStart.Format(), resp.Header.Get("Location"))

	cookies := resp.Cookies()
	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]

		assert.Equal(t, "cookies-consent", cookie.Name)
		assert.Equal(t, "accept", cookie.Value)
		assert.Equal(t, 365*24*60*60, cookie.MaxAge)
		assert.Equal(t, "/", cookie.Path)
	}
}
