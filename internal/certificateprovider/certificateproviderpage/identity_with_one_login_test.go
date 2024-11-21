package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdentityWithOneLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "cy", true).
		Return("http://auth", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetOneLogin(r, w, &sesh.OneLoginSession{State: "i am random", Nonce: "i am random", Locale: "cy", Redirect: certificateprovider.PathIdentityWithOneLoginCallback.Format("lpa-id")}).
		Return(nil)

	err := IdentityWithOneLogin(client, sessionStore, func(int) string { return "i am random" })(appcontext.Data{LpaID: "lpa-id", Lang: localize.Cy}, w, r, nil, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
}

func TestIdentityWithOneLoginWhenAuthCodeURLError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "", true).
		Return("http://auth?locale=en", expectedError)

	err := IdentityWithOneLogin(client, nil, func(int) string { return "i am random" })(testAppData, w, r, nil, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIdentityWithOneLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL("i am random", "i am random", "", true).
		Return("http://auth?locale=en", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetOneLogin(r, w, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLogin(client, sessionStore, func(int) string { return "i am random" })(testAppData, w, r, nil, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
