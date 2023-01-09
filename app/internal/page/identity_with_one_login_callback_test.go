package page

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)
	lpaStore.On("Put", mock.Anything, "session-id", mock.MatchedBy(func(lpa *Lpa) bool {
		now := time.Now()
		lpa.OneLoginUserData.RetrievedAt = now

		return assert.Equal(t, &Lpa{OneLoginUserData: identity.UserData{OK: true, FullName: "an-identity-jwt", RetrievedAt: now}}, lpa)
	})).Return(nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.On("Get", mock.Anything, "params").Return(&sessions.Session{Values: map[interface{}]interface{}{"nonce": "a-nonce"}}, nil)

	authRedirectClient := &mockAuthRedirectClient{}
	authRedirectClient.On("Exchange", r.Context(), "a-code", "a-nonce").Return("a-jwt", nil)
	authRedirectClient.On("UserInfo", "a-jwt").Return(onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *identityWithOneLoginCallbackData) bool {
			now := time.Now()
			data.ConfirmedAt = now

			return assert.Equal(t, &identityWithOneLoginCallbackData{
				App:         appData,
				FullName:    "an-identity-jwt",
				ConfirmedAt: now,
			}, data)
		})).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Func, authRedirectClient, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, authRedirectClient, template)
}

func TestGetIdentityWithOneLoginCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.On("Get", mock.Anything, "params").Return(&sessions.Session{Values: map[interface{}]interface{}{"nonce": "a-nonce"}}, nil)

	authRedirectClient := &mockAuthRedirectClient{}
	authRedirectClient.On("Exchange", mock.Anything, mock.Anything, mock.Anything).Return("a-jwt", nil)
	authRedirectClient.On("UserInfo", mock.Anything).Return(onelogin.UserInfo{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	err := IdentityWithOneLoginCallback(nil, authRedirectClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, authRedirectClient)
}

func TestGetIdentityWithOneLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithOneLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)
	lpaStore.On("Put", mock.Anything, "session-id", mock.Anything).Return(expectedError)

	sessionStore := &mockSessionsStore{}
	sessionStore.On("Get", mock.Anything, "params").Return(&sessions.Session{Values: map[interface{}]interface{}{"nonce": "a-nonce"}}, nil)

	authRedirectClient := &mockAuthRedirectClient{}
	authRedirectClient.On("Exchange", mock.Anything, mock.Anything, mock.Anything).Return("a-jwt", nil)
	authRedirectClient.On("UserInfo", mock.Anything).Return(userInfo, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	err := IdentityWithOneLoginCallback(nil, authRedirectClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, authRedirectClient)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{OneLoginUserData: userData}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithOneLoginCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	err := IdentityWithOneLoginCallback(template.Func, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{
		OneLoginUserData: identity.UserData{OK: true},
	}, nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ReadYourLpa, resp.Header.Get("Location"))
}

func TestPostIdentityWithOneLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.SelectYourIdentityOptions1, resp.Header.Get("Location"))
}
