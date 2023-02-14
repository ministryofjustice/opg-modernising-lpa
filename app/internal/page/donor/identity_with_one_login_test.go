package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdentityWithOneLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := &page.MockOneLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random", "cy", true).
		Return("http://auth")

	sessionsStore := &page.MockSessionsStore{}

	session := sessions.NewSession(sessionsStore, "params")

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"one-login": &sesh.OneLoginSession{State: "i am random", Nonce: "i am random", Locale: "cy", Identity: true, LpaID: "123"},
	}

	sessionsStore.
		On("Save", r, w, session).
		Return(nil)

	err := IdentityWithOneLogin(nil, client, sessionsStore, func(int) string { return "i am random" })(page.AppData{Lang: localize.Cy, LpaID: "123"}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, client, sessionsStore)
}

func TestIdentityWithOneLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := &page.MockLogger{}
	logger.
		On("Print", page.ExpectedError)

	client := &page.MockOneLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random", "", true).
		Return("http://auth?locale=en")

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Save", r, w, mock.Anything).
		Return(page.ExpectedError)

	err := IdentityWithOneLogin(logger, client, sessionsStore, func(int) string { return "i am random" })(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, logger, client, sessionsStore)
}
