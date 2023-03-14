package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithYoti(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	sessionStore := newMockSessionStore(t)
	session := sessions.NewSession(sessionStore, "yoti")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"yoti": &sesh.YotiSession{Locale: "en", LpaID: "lpa-id"},
	}
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")
	yotiClient.On("ScenarioID").Return("a-scenario-id")

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithYotiData{
			App:         testAppData,
			ClientSdkID: "an-sdk-id",
			ScenarioID:  "a-scenario-id",
		}).
		Return(nil)

	err := IdentityWithYoti(template.Execute, lpaStore, sessionStore, yotiClient)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithYotiWhenAlreadyProvided(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.EasyID}}, nil)

	err := IdentityWithYoti(nil, lpaStore, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
}

func TestGetIdentityWithYotiWhenTest(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("IsTest").Return(true)

	err := IdentityWithYoti(nil, lpaStore, nil, yotiClient)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
}

func TestGetIdentityWithYotiWhenDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, expectedError)

	err := IdentityWithYoti(nil, lpaStore, nil, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithYotiWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")
	yotiClient.On("ScenarioID").Return("a-scenario-id")

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := IdentityWithYoti(template.Execute, lpaStore, sessionStore, yotiClient)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}
