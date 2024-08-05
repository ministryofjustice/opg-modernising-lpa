package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	sharecode "github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeAttorneyLoggedOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})).
		Return(&lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeAttorneyDataLoggedOut{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeAttorneyLoggedOut(template.Execute, nil, lpaStoreResolvingService, sessionStore, nil, "example.com")(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyLoggedOutWhenSessionStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{}, expectedError)

	err := ConfirmDontWantToBeAttorneyLoggedOut(nil, nil, nil, sessionStore, nil, "example.com")(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyLoggedOutWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpastore.Lpa{}, expectedError)

	err := ConfirmDontWantToBeAttorneyLoggedOut(nil, nil, lpaStoreResolvingService, sessionStore, nil, "example.com")(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyLoggedOutWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmDontWantToBeAttorneyLoggedOut(template.Execute, nil, lpaStoreResolvingService, sessionStore, nil, "example.com")(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDontWantToBeAttorneyLoggedOut(t *testing.T) {
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	testcases := map[string]struct {
		uid              actoruid.UID
		attorneyFullName string
	}{
		"attorney": {
			uid:              attorneyUID,
			attorneyFullName: "d e f",
		},
		"replacement attorney": {
			uid:              replacementAttorneyUID,
			attorneyFullName: "x y z",
		},
		"trust corporation": {
			uid:              trustCorporationUID,
			attorneyFullName: "trusty",
		},
		"replacement trust corporation": {
			uid:              replacementTrustCorporationUID,
			attorneyFullName: "untrusty",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
			w := httptest.NewRecorder()
			ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

			lpa := &lpastore.Lpa{
				LpaUID:   "lpa-uid",
				SignedAt: time.Now(),
				Donor: lpastore.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
				},
				Attorneys: lpastore.Attorneys{
					TrustCorporation: lpastore.TrustCorporation{UID: trustCorporationUID, Name: "trusty"},
					Attorneys: []lpastore.Attorney{
						{UID: attorneyUID, FirstNames: "d e", LastName: "f"},
					},
				},
				ReplacementAttorneys: lpastore.Attorneys{
					TrustCorporation: lpastore.TrustCorporation{UID: replacementTrustCorporationUID, Name: "untrusty"},
					Attorneys: []lpastore.Attorney{
						{UID: replacementAttorneyUID, FirstNames: "x y", LastName: "z"},
					},
				},
				Type: lpadata.LpaTypePersonalWelfare,
			}

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				LpaData(r).
				Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

			shareCodeData := sharecode.Data{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				ActorUID:    tc.uid,
			}

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, "123").
				Return(shareCodeData, nil)
			shareCodeStore.EXPECT().
				Delete(r.Context(), shareCodeData).
				Return(nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(ctx).
				Return(lpa, nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				EmailGreeting(lpa).
				Return("Dear donor")
			notifyClient.EXPECT().
				SendActorEmail(ctx, "a@example.com", "lpa-uid", notify.AttorneyOptedOutEmail{
					Greeting:          "Dear donor",
					AttorneyFullName:  tc.attorneyFullName,
					DonorFullName:     "a b c",
					LpaType:           "Personal welfare",
					LpaUID:            "lpa-uid",
					DonorStartPageURL: "example.com" + page.Paths.Start.Format(),
				}).
				Return(nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("personal-welfare").
				Return("Personal welfare")

			testAppData.Localizer = localizer

			err := ConfirmDontWantToBeAttorneyLoggedOut(nil, shareCodeStore, lpaStoreResolvingService, sessionStore, notifyClient, "example.com")(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.Paths.Attorney.YouHaveDecidedNotToBeAttorney.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeAttorneyLoggedOutWhenAttorneyNotFound(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	shareCodeData := sharecode.Data{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		ActorUID:    actoruid.New(),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "123").
		Return(shareCodeData, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(ctx).
		Return(&lpastore.Lpa{
			LpaUID:   "lpa-uid",
			SignedAt: time.Now(),
			Donor: lpastore.Donor{
				FirstNames: "a b", LastName: "c", Email: "a@example.com",
			},
			Type: lpadata.LpaTypePersonalWelfare,
		}, nil)

	err := ConfirmDontWantToBeAttorneyLoggedOut(nil, shareCodeStore, lpaStoreResolvingService, sessionStore, nil, "example.com")(testAppData, w, r)
	assert.EqualError(t, err, "attorney not found")
}

func TestPostConfirmDontWantToBeAttorneyLoggedOutErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

	shareCodeData := sharecode.Data{
		LpaKey: dynamo.LpaKey("lpa-id"),
	}

	signedLPA := lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()}

	localizer := func(t *testing.T) *mockLocalizer {
		l := newMockLocalizer(t)
		l.EXPECT().
			T(mock.Anything).
			Return("a")

		return l
	}

	testcases := map[string]struct {
		sessionStore             func(*testing.T) *mockSessionStore
		lpaStoreResolvingService func(*testing.T) *mockLpaStoreResolvingService
		localizer                func(*testing.T) *mockLocalizer
		shareCodeStore           func(*testing.T) *mockShareCodeStore
		notifyClient             func(*testing.T) *mockNotifyClient
	}{
		"when shareCodeStore.Get() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, expectedError)

				return shareCodeStore
			},
		},
		"when shareCodeStore.Delete() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			localizer: localizer,
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)
				shareCodeStore.EXPECT().
					Delete(mock.Anything, mock.Anything).
					Return(expectedError)

				return shareCodeStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
		},
		"when notifyClient.SendActorEmail() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			localizer: localizer,
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			testAppData.Localizer = evalT(tc.localizer, t)

			err := ConfirmDontWantToBeAttorneyLoggedOut(nil, evalT(tc.shareCodeStore, t), evalT(tc.lpaStoreResolvingService, t), evalT(tc.sessionStore, t), evalT(tc.notifyClient, t), "example.com")(testAppData, w, r)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
