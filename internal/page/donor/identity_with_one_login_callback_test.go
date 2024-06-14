package donor

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
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
	updatedDonor := &actor.DonorProvidedDetails{
		Donor:                 actor.Donor{FirstNames: "John", LastName: "Doe"},
		DonorIdentityUserData: userData,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
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

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FirstNames:  "John",
			LastName:    "Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, donorStore, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "John", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithOneLoginCallbackWhenLpaPaidAndSigned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FirstNames: "John", LastName: "Doe", RetrievedAt: now}
	updatedDonor := &actor.DonorProvidedDetails{
		Donor:                 actor.Donor{FirstNames: "John", LastName: "Doe"},
		DonorIdentityUserData: userData,
		Tasks: actor.DonorTasks{
			PayForLpa:                  actor.PaymentTaskCompleted,
			ConfirmYourIdentityAndSign: actor.TaskCompleted,
		},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
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
		SendDonorConfirmIdentity(r.Context(), updatedDonor).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FirstNames:  "John",
			LastName:    "Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, donorStore, lpaStoreClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "John", LastName: "Doe"},
		Tasks: actor.DonorTasks{
			PayForLpa:                  actor.PaymentTaskCompleted,
			ConfirmYourIdentityAndSign: actor.TaskCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	templateCalled := func(t *testing.T, w io.Writer) *mockTemplate {
		template := newMockTemplate(t)
		template.EXPECT().
			Execute(w, &identityWithOneLoginCallbackData{
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
		sessionStore.EXPECT().
			OneLogin(mock.Anything).
			Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)
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
			template:     templateCalled,
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
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
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
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
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

			sessionStore := tc.sessionStore(t)
			oneLoginClient := tc.oneLoginClient(t)
			template := tc.template(t, w)

			err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenPutDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

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
		Return(identity.UserData{OK: true}, nil)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, donorStore, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FirstNames: "John", LastName: "Doe", RetrievedAt: now}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(r.Context(), "a-code", "a-nonce").
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(r.Context(), mock.Anything).
		Return(userData, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendDonorConfirmIdentity(r.Context(), mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, donorStore, lpaStoreClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "John", LastName: "Doe"},
		Tasks: actor.DonorTasks{
			PayForLpa:                  actor.PaymentTaskCompleted,
			ConfirmYourIdentityAndSign: actor.TaskCompleted,
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FirstNames:  "first-name",
			LastName:    "last-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor:                 actor.Donor{FirstNames: "first-name", LastName: "last-name"},
		DonorIdentityUserData: userData,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		DonorIdentityUserData: identity.UserData{OK: true},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ReadYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostIdentityWithOneLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{
			CanSign: form.Yes,
		},
		Type: actor.LpaTypePersonalWelfare,
		Tasks: actor.DonorTasks{
			YourDetails:                actor.TaskCompleted,
			ChooseAttorneys:            actor.TaskCompleted,
			ChooseReplacementAttorneys: actor.TaskCompleted,
			LifeSustainingTreatment:    actor.TaskCompleted,
			Restrictions:               actor.TaskCompleted,
			CertificateProvider:        actor.TaskCompleted,
			PeopleToNotify:             actor.TaskCompleted,
			CheckYourLpa:               actor.TaskCompleted,
			PayForLpa:                  actor.PaymentTaskCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ProveYourIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}
