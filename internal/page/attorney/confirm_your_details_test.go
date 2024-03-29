package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourDetails(t *testing.T) {
	uid := actoruid.New()
	attorneyProvidedDetails := &actor.AttorneyProvidedDetails{UID: uid}

	testcases := map[string]struct {
		appData page.AppData
		donor   *actor.DonorProvidedDetails
		data    *confirmYourDetailsData
	}{
		"attorney": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid, FirstNames: "John"}}},
			},
			data: &confirmYourDetailsData{
				App: testAppData,
				Donor: &actor.DonorProvidedDetails{
					Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid, FirstNames: "John"}}},
				},
				Attorney:                actor.Attorney{UID: uid, FirstNames: "John"},
				AttorneyProvidedDetails: attorneyProvidedDetails,
			},
		},
		"trust corporation": {
			appData: testTrustCorporationAppData,
			donor: &actor.DonorProvidedDetails{
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "company"}},
			},
			data: &confirmYourDetailsData{
				App: testTrustCorporationAppData,
				Donor: &actor.DonorProvidedDetails{
					Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "company"}},
				},
				TrustCorporation:        actor.TrustCorporation{Name: "company"},
				AttorneyProvidedDetails: attorneyProvidedDetails,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				GetAny(r.Context()).
				Return(tc.donor, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.data).
				Return(nil)

			err := ConfirmYourDetails(template.Execute, nil, donorStore)(tc.appData, w, r, attorneyProvidedDetails)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetConfirmYourDetailsWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, expectedError)

	err := ConfirmYourDetails(nil, nil, donorStore)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, nil, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), &actor.AttorneyProvidedDetails{
			UID:   uid,
			LpaID: "lpa-id",
			Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	err := ConfirmYourDetails(nil, attorneyStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{UID: uid, LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.ReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, attorneyStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{UID: actoruid.New(), LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
