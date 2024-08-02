package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?sessionId=session-id&lpaId=lpa-id", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "cy", false).
		Return("http://auth", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetOneLogin(r, w, &sesh.OneLoginSession{
			State:    "i am random",
			Nonce:    "i am random",
			Locale:   "cy",
			Redirect: "/redirect",
		}).
		Return(nil)

	Login(client, sessionStore, func(int) string { return "i am random" }, "/redirect")(appcontext.Data{Lang: localize.Cy}, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
}

func TestLoginDefaultLocale(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?sessionId=session-id&lpaId=lpa-id&identity=1", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "en", false).
		Return("http://auth", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetOneLogin(r, w, &sesh.OneLoginSession{
			State:    "i am random",
			Nonce:    "i am random",
			Locale:   "en",
			Redirect: "/redirect",
		}).
		Return(nil)

	Login(client, sessionStore, func(int) string { return "i am random" }, "/redirect")(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
}

func TestLoginWhenAuthCodeURLError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "en", false).
		Return("http://auth?locale=en", expectedError)

	err := Login(client, nil, func(int) string { return "i am random" }, "/redirect")(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "en", false).
		Return("http://auth?locale=en", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetOneLogin(r, w, mock.Anything).
		Return(expectedError)

	err := Login(client, sessionStore, func(int) string { return "i am random" }, "/redirect")(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
