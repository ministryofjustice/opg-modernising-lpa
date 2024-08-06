package certificateproviderpage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	now := time.Now()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "John", LastName: "Doe", RetrievedAt: now}

	updatedCertificateProvider := &certificateproviderdata.Provided{
		IdentityUserData: userData,
		LpaID:            "lpa-id",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), updatedCertificateProvider).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpastore.CertificateProvider{FirstNames: "John", LastName: "Doe"}}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(r.Context(), "a-code", "a-nonce").
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(r.Context(), userInfo).
		Return(userData, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProviderConfirmIdentity(r.Context(), "lpa-uid", updatedCertificateProvider).
		Return(nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreResolvingService, nil, lpaStoreClient, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathOneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenFailedIDCheck(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusFailed}

	updatedCertificateProvider := &certificateproviderdata.Provided{
		IdentityUserData: userData,
		LpaID:            "lpa-id",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), updatedCertificateProvider).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			LpaUID:              "lpa-uid",
			CertificateProvider: lpastore.CertificateProvider{FirstNames: "a", LastName: "b"},
			Donor:               lpastore.Donor{Email: "a@example.com", FirstNames: "c", LastName: "d"},
			Type:                lpadata.LpaTypePersonalWelfare,
			SignedAt:            time.Now(),
		}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(r.Context(), "a-code", "a-nonce").
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(r.Context(), userInfo).
		Return(userData, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("translated LPA type")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("Dear donor")
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), "a@example.com", "lpa-uid", notify.CertificateProviderFailedIDCheckEmail{
			Greeting:                    "Dear donor",
			DonorFullName:               "c d",
			CertificateProviderFullName: "a b",
			LpaType:                     "translated LPA type",
			DonorStartPageURL:           "www.example.com" + page.Paths.Start.Format(),
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreResolvingService, notifyClient, nil, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathUnableToConfirmIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenSendingEmailError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusFailed}

	updatedCertificateProvider := &certificateproviderdata.Provided{
		IdentityUserData: userData,
		LpaID:            "lpa-id",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(mock.Anything, updatedCertificateProvider).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpastore.Lpa{
			LpaUID:              "lpa-uid",
			CertificateProvider: lpastore.CertificateProvider{FirstNames: "a", LastName: "b"},
			Donor:               lpastore.Donor{Email: "a@example.com", FirstNames: "c", LastName: "d"},
			Type:                lpadata.LpaTypePersonalWelfare,
			SignedAt:            time.Now(),
		}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(mock.Anything, mock.Anything).
		Return(userData, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("translated LPA type")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreResolvingService, notifyClient, nil, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	sessionRetrieved := func(t *testing.T) *mockSessionStore {
		sessionStore := newMockSessionStore(t)
		sessionStore.EXPECT().
			OneLogin(mock.Anything).
			Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)
		return sessionStore
	}

	sessionIgnored := func(t *testing.T) *mockSessionStore {
		return nil
	}
	certificateProviderIgnored := func(t *testing.T) *mockCertificateProviderStore {
		return nil
	}

	testCases := map[string]struct {
		oneLoginClient           func(t *testing.T) *mockOneLoginClient
		sessionStore             func(*testing.T) *mockSessionStore
		certificateProviderStore func(t *testing.T) *mockCertificateProviderStore
		url                      string
		error                    error
		expectedRedirectURL      string
		expectedStatus           int
	}{
		"not ok": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.EXPECT().
					ParseIdentityClaim(mock.Anything, mock.Anything).
					Return(identity.UserData{}, nil)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			certificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.EXPECT().
					Put(context.Background(), &certificateproviderdata.Provided{LpaID: "lpa-id"}).
					Return(nil)

				return certificateProviderStore
			},
			expectedRedirectURL: certificateprovider.PathUnableToConfirmIdentity.Format("lpa-id"),
			expectedStatus:      http.StatusFound,
		},
		"errored on parse": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.EXPECT().
					ParseIdentityClaim(mock.Anything, mock.Anything).
					Return(identity.UserData{Status: identity.StatusConfirmed}, expectedError)
				return oneLoginClient
			},
			sessionStore:             sessionRetrieved,
			error:                    expectedError,
			certificateProviderStore: certificateProviderIgnored,
			expectedStatus:           http.StatusOK,
		},
		"errored on userinfo": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
					Return(onelogin.UserInfo{}, expectedError)
				return oneLoginClient
			},
			sessionStore:             sessionRetrieved,
			error:                    expectedError,
			certificateProviderStore: certificateProviderIgnored,
			expectedStatus:           http.StatusOK,
		},
		"errored on exchange": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("", "", expectedError)
				return oneLoginClient
			},
			sessionStore:             sessionRetrieved,
			error:                    expectedError,
			certificateProviderStore: certificateProviderIgnored,
			expectedStatus:           http.StatusOK,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return newMockOneLoginClient(t)
			},
			sessionStore:             sessionIgnored,
			error:                    errors.New("access denied"),
			certificateProviderStore: certificateProviderIgnored,
			expectedStatus:           http.StatusOK,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpastore.Lpa{CertificateProvider: lpastore.CertificateProvider{}}, nil)

			sessionStore := tc.sessionStore(t)
			oneLoginClient := tc.oneLoginClient(t)

			err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, tc.certificateProviderStore(t), lpaStoreResolvingService, nil, nil, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirectURL, resp.Header.Get("Location"))
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenGetLpaStoreResolvingServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{CertificateProvider: lpastore.CertificateProvider{}}, expectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStoreResolvingService, nil, nil, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenPutCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{CertificateProvider: lpastore.CertificateProvider{}}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(mock.Anything, mock.Anything).
		Return(identity.UserData{Status: identity.StatusConfirmed}, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreResolvingService, nil, nil, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenLpaStoreClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{CertificateProvider: lpastore.CertificateProvider{}}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(mock.Anything, mock.Anything).
		Return(identity.UserData{Status: identity.StatusConfirmed}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProviderConfirmIdentity(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreResolvingService, nil, lpaStoreClient, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "first-names", LastName: "last-name", RetrievedAt: now}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{CertificateProvider: lpastore.CertificateProvider{FirstNames: "first-names", LastName: "last-name"}}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStoreResolvingService, nil, nil, "www.example.com")(testAppData, w, r, &certificateproviderdata.Provided{
		IdentityUserData: userData,
		LpaID:            "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathOneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}
