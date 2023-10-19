package certificateprovider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FirstNames: "John", LastName: "Doe", RetrievedAt: now}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{
			IdentityUserData: userData,
			Tasks: actor.CertificateProviderTasks{
				ConfirmYourIdentity: actor.TaskCompleted,
			},
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{FirstNames: "John", LastName: "Doe"}}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"},
			},
		}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", r.Context(), "a-code", "a-nonce").
		Return("id-token", "a-jwt", nil)
	oneLoginClient.
		On("UserInfo", r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", r.Context(), userInfo).
		Return(userData, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FirstNames:  "John",
			LastName:    "Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, certificateProviderStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	templateCalled := func(t *testing.T, w io.Writer) *mockTemplate {
		template := newMockTemplate(t)
		template.
			On("Execute", w, &identityWithOneLoginCallbackData{
				App:             testAppData,
				CouldNotConfirm: true,
			}).
			Return(nil)
		return template
	}

	templateIgnored := func(t *testing.T, w io.Writer) *mockTemplate {
		return nil
	}

	sessionRetrieved := func(t *testing.T) *mockSessionStore {
		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Get", mock.Anything, "params").
			Return(&sessions.Session{
				Values: map[any]any{
					"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"},
				},
			}, nil)
		return sessionStore
	}

	sessionIgnored := func(t *testing.T) *mockSessionStore {
		return nil
	}

	testCases := map[string]struct {
		oneLoginClient func(t *testing.T) *mockOneLoginClient
		sessionStore   func(*testing.T) *mockSessionStore
		template       func(*testing.T, io.Writer) *mockTemplate
		url            string
		error          error
	}{
		"not a match": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{OK: true, FirstNames: "x", LastName: "y"}, nil)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			template:     templateCalled,
		},
		"not ok": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{}, nil)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			template:     templateCalled,
		},
		"errored on parse": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{OK: true}, expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			template:     templateIgnored,
			error:        expectedError,
		},
		"errored on userinfo": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(onelogin.UserInfo{}, expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			template:     templateIgnored,
			error:        expectedError,
		},
		"errored on exchange": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("", "", expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			template:     templateIgnored,
			error:        expectedError,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return newMockOneLoginClient(t)
			},
			sessionStore: sessionIgnored,
			template:     templateCalled,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("Get", r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{}, nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

			sessionStore := tc.sessionStore(t)
			oneLoginClient := tc.oneLoginClient(t)
			template := tc.template(t, w)

			err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, certificateProviderStore, donorStore)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenGetCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, certificateProviderStore, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenGetDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, expectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, certificateProviderStore, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenPutCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"},
			},
		}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, certificateProviderStore, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FirstNames: "first-names", LastName: "last-name", RetrievedAt: now}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{
			IdentityUserData: userData,
		}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{FirstNames: "first-names", LastName: "last-name"}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FirstNames:  "first-names",
			LastName:    "last-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, nil, nil, certificateProviderStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{
			LpaID:            "lpa-id",
			IdentityUserData: identity.UserData{OK: true},
		}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, certificateProviderStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.ReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostIdentityWithOneLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, certificateProviderStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.IdentityWithOneLogin.Format("lpa-id"), resp.Header.Get("Location"))
}
