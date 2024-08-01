package attorneypage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeAttorneyData{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeAttorney(template.Execute, lpaStoreResolvingService, nil, nil, "example.com")(testAppData, w, r, &attorneydata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpastore.Lpa{}, expectedError)

	err := ConfirmDontWantToBeAttorney(nil, lpaStoreResolvingService, nil, nil, "example.com")(testAppData, w, r, &attorneydata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmDontWantToBeAttorney(template.Execute, lpaStoreResolvingService, nil, nil, "example.com")(testAppData, w, r, &attorneydata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDontWantToBeAttorney(t *testing.T) {
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"}), http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	uid := actoruid.New()

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			LpaUID:   "lpa-uid",
			SignedAt: time.Now(),
			Donor: lpastore.Donor{
				FirstNames: "a b", LastName: "c", Email: "a@example.com",
			},
			Attorneys: lpastore.Attorneys{
				Attorneys: []lpastore.Attorney{
					{FirstNames: "d e", LastName: "f", UID: uid},
				},
			},
			Type: actor.LpaTypePersonalWelfare,
		}, nil)

	certificateProviderStore := newMockAttorneyStore(t)
	certificateProviderStore.EXPECT().
		Delete(r.Context()).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), "a@example.com", "lpa-uid", notify.AttorneyOptedOutEmail{
			AttorneyFullName:  "d e f",
			DonorFullName:     "a b c",
			LpaType:           "Personal welfare",
			LpaUID:            "lpa-uid",
			DonorStartPageURL: "example.com" + page.Paths.Start.Format(),
		}).
		Return(nil)

	err := ConfirmDontWantToBeAttorney(nil, lpaStoreResolvingService, certificateProviderStore, notifyClient, "example.com")(testAppData, w, r, &attorneydata.Provided{
		UID: uid,
	})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, page.Paths.Attorney.YouHaveDecidedNotToBeAttorney.Format()+"?donorFirstNames=a+b&donorFullName=a+b+c", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestPostConfirmDontWantToBeAttorneyWhenAttorneyNotFound(t *testing.T) {
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"}), http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	uid := actoruid.New()

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			LpaUID:   "lpa-uid",
			SignedAt: time.Now(),
			Donor: lpastore.Donor{
				FirstNames: "a b", LastName: "c", Email: "a@example.com",
			},
			Type: actor.LpaTypePersonalWelfare,
		}, nil)

	err := ConfirmDontWantToBeAttorney(nil, lpaStoreResolvingService, nil, nil, "example.com")(testAppData, w, r, &attorneydata.Provided{
		UID: uid,
	})
	assert.EqualError(t, err, "attorney not found")
}

func TestPostConfirmDontWantToBeAttorneyErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)

	testcases := map[string]struct {
		certificateProviderStore func(*testing.T) *mockAttorneyStore
		notifyClient             func(*testing.T) *mockNotifyClient
	}{
		"when attorneyStore.Delete() error": {
			certificateProviderStore: func(t *testing.T) *mockAttorneyStore {
				certificateProviderStore := newMockAttorneyStore(t)
				certificateProviderStore.EXPECT().
					Delete(mock.Anything).
					Return(expectedError)

				return certificateProviderStore
			},
		},
		"when notifyClient.SendActorEmail() error": {
			certificateProviderStore: func(t *testing.T) *mockAttorneyStore {
				certificateProviderStore := newMockAttorneyStore(t)
				certificateProviderStore.EXPECT().
					Delete(mock.Anything).
					Return(nil)

				return certificateProviderStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
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

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()}, nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(mock.Anything).
				Return("a")

			testAppData.Localizer = localizer

			err := ConfirmDontWantToBeAttorney(nil, lpaStoreResolvingService, evalT(tc.certificateProviderStore, t), evalT(tc.notifyClient, t), "example.com")(testAppData, w, r, &attorneydata.Provided{})

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
