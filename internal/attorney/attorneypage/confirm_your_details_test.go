package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourDetails(t *testing.T) {
	uid := actoruid.New()
	attorneyProvidedDetails := &attorneydata.Provided{UID: uid}

	testcases := map[string]struct {
		appData page.AppData
		donor   *lpastore.Lpa
		data    *confirmYourDetailsData
	}{
		"attorney": {
			appData: testAppData,
			donor: &lpastore.Lpa{
				Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, FirstNames: "John"}}},
			},
			data: &confirmYourDetailsData{
				App: testAppData,
				Lpa: &lpastore.Lpa{
					Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, FirstNames: "John"}}},
				},
				Attorney:                lpastore.Attorney{UID: uid, FirstNames: "John"},
				AttorneyProvidedDetails: attorneyProvidedDetails,
			},
		},
		"trust corporation": {
			appData: testTrustCorporationAppData,
			donor: &lpastore.Lpa{
				Attorneys: lpastore.Attorneys{TrustCorporation: lpastore.TrustCorporation{Name: "company"}},
			},
			data: &confirmYourDetailsData{
				App: testTrustCorporationAppData,
				Lpa: &lpastore.Lpa{
					Attorneys: lpastore.Attorneys{TrustCorporation: lpastore.TrustCorporation{Name: "company"}},
				},
				TrustCorporation:        lpastore.TrustCorporation{Name: "company"},
				AttorneyProvidedDetails: attorneyProvidedDetails,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.donor, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.data).
				Return(nil)

			err := ConfirmYourDetails(template.Execute, nil, lpaStoreResolvingService)(tc.appData, w, r, attorneyProvidedDetails)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetConfirmYourDetailsWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := ConfirmYourDetails(nil, nil, lpaStoreResolvingService)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, nil, lpaStoreResolvingService)(testAppData, w, r, &attorneydata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), &attorneydata.Provided{
			UID:   uid,
			LpaID: "lpa-id",
			Tasks: attorneydata.Tasks{ConfirmYourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	err := ConfirmYourDetails(nil, attorneyStore, nil)(testAppData, w, r, &attorneydata.Provided{UID: uid, LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, attorneyStore, nil)(testAppData, w, r, &attorneydata.Provided{UID: actoruid.New(), LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
