package page

import (
	"context"
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

func TestGetCertificateProviderLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()
	userInfo := onelogin.UserInfo{Sub: "a-sub", Email: "a-email", CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "John Doe", RetrievedAt: now}

	sessionStore := &mockSessionsStore{}
	session := sessions.NewSession(sessionStore, "session")

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"certificate-provider": &CertificateProviderSession{
			Sub:       "a-sub",
			Email:     "a-email",
			LpaID:     "lpa-id",
			SessionID: "session-id",
		},
	}

	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	ctxMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		session := sessionDataFromContext(ctx)

		return assert.Equal(t, &sessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
	})

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", ctxMatcher).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", ctxMatcher, &Lpa{
			CertificateProviderUserData: userData,
		}).
		Return(nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", ctxMatcher, "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", ctxMatcher, "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", ctxMatcher, userInfo).
		Return(userData, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &certificateProviderLoginCallbackData{
			App:         appData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := CertificateProviderLoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient, template)
}

func TestGetCertificateProviderLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	testCases := map[string]struct {
		userData identity.UserData
		url      string
		error    error
	}{
		"not ok": {
			url: "/?code=a-code",
		},
		"errored": {
			url:      "/?code=a-code",
			userData: identity.UserData{OK: true},
			error:    expectedError,
		},
		"provider access denied": {
			url:      "/?error=access_denied",
			userData: identity.UserData{OK: true},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything).
				Return(&Lpa{}, nil)

			sessionStore := &mockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &OneLoginSession{
							State:               "a-state",
							Nonce:               "a-nonce",
							CertificateProvider: true,
							Identity:            true,
							LpaID:               "lpa-id",
							SessionID:           "session-id",
						},
					},
				}, nil)

			oneLoginClient := &mockOneLoginClient{}
			oneLoginClient.
				On("Exchange", mock.Anything, mock.Anything, mock.Anything).
				Return("a-jwt", nil)
			oneLoginClient.
				On("UserInfo", mock.Anything, mock.Anything).
				Return(userInfo, nil)
			oneLoginClient.
				On("ParseIdentityClaim", mock.Anything, mock.Anything).
				Return(tc.userData, tc.error)

			template := &mockTemplate{}
			template.
				On("Func", w, &certificateProviderLoginCallbackData{
					App:             appData,
					CouldNotConfirm: true,
				}).
				Return(nil)

			err := CertificateProviderLoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetCertificateProviderLoginCallbackWhenExchangeError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&Lpa{}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("", expectedError)

	err := CertificateProviderLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetCertificateProviderLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&Lpa{}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(onelogin.UserInfo{}, expectedError)

	err := CertificateProviderLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetCertificateProviderLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&Lpa{}, expectedError)

	err := CertificateProviderLoginCallback(nil, nil, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, sessionStore, lpaStore)
}

func TestGetCertificateProviderLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, mock.Anything).
		Return(expectedError)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := CertificateProviderLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetCertificateProviderLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&Lpa{CertificateProviderUserData: userData}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &certificateProviderLoginCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := CertificateProviderLoginCallback(template.Func, nil, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionStore, lpaStore, template)
}

func TestPostCertificateProviderLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&Lpa{
		CertificateProviderUserData: identity.UserData{OK: true},
	}, nil)

	err := CertificateProviderLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.CertificateProviderYourDetails, resp.Header.Get("Location"))
}

func TestPostCertificateProviderLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&Lpa{}, nil)

	err := CertificateProviderLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.Start, resp.Header.Get("Location"))
}
