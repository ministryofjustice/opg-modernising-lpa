package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestYotiRedirect(t *testing.T) {
	testcases := map[string]struct {
		session  *sesh.YotiSession
		redirect string
	}{
		"donor identity": {
			session: &sesh.YotiSession{
				LpaID: "123",
			},
			redirect: "/lpa/123" + Paths.IdentityWithYotiCallback,
		},
		"donor identity welsh": {
			session: &sesh.YotiSession{
				Locale: "cy",
				LpaID:  "123",
			},
			redirect: "/cy/lpa/123" + Paths.IdentityWithYotiCallback,
		},
		"certificate provider identity": {
			session: &sesh.YotiSession{
				Locale:              "en",
				LpaID:               "123",
				CertificateProvider: true,
			},
			redirect: Paths.CertificateProviderIdentityWithYotiCallback,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?token=my-token", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.
				On("Get", r, "yoti").
				Return(&sessions.Session{
					Values: map[any]any{
						"yoti": tc.session,
					},
				}, nil)

			YotiRedirect(nil, sessionStore)(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect+"?token=my-token", resp.Header.Get("Location"))
		})
	}
}

func TestYotiRedirectSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=my-token", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", ExpectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "yoti").
		Return(nil, ExpectedError)

	YotiRedirect(logger, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
