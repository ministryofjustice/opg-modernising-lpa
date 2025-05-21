package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	trustCorporations := []donordata.TrustCorporation{{Name: "Corp"}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return(trustCorporations, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseTrustCorporationData{
			App:                 testAppData,
			Form:                &chooseTrustCorporationForm{},
			Donor:               &donordata.Provided{LpaID: "lpa-id"},
			TrustCorporations:   trustCorporations,
			ChooseAttorneysPath: donor.PathEnterReplacementAttorney.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}),
		}).
		Return(nil)

	err := ChooseReplacementTrustCorporation(template.Execute, nil, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementTrustCorporationWhenNoReusableTrustCorporations(t *testing.T) {
	testcases := map[string]error{
		"none":      nil,
		"not found": dynamo.NotFoundError{},
	}

	for name, reuseError := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				TrustCorporations(r.Context()).
				Return(nil, reuseError)

			err := ChooseReplacementTrustCorporation(nil, nil, reuseStore, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathEnterReplacementTrustCorporation.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestGetChooseReplacementTrustCorporationWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return(nil, expectedError)

	err := ChooseReplacementTrustCorporation(nil, nil, reuseStore, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseReplacementTrustCorporationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return([]donordata.TrustCorporation{{Name: "Corp"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseReplacementTrustCorporation(template.Execute, nil, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementTrustCorporation(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	trustCorporations := []donordata.TrustCorporation{{Name: "Corp"}, {Name: "Trust", Address: place.Address{Line1: "1"}}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return(trustCorporations, nil)
	reuseStore.EXPECT().
		PutTrustCorporation(r.Context(), donordata.TrustCorporation{
			UID:     testUID,
			Name:    "Trust",
			Address: place.Address{Line1: "1"},
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			ReplacementAttorneys: donordata.Attorneys{
				TrustCorporation: donordata.TrustCorporation{UID: testUID, Name: "Trust", Address: place.Address{Line1: "1"}},
			},
			Tasks: donordata.Tasks{ChooseReplacementAttorneys: task.StateCompleted},
		}).
		Return(nil)

	err := ChooseReplacementTrustCorporation(nil, donorStore, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementTrustCorporationWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	trustCorporations := []donordata.TrustCorporation{{Name: "Corp"}, {Name: "Trust"}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return(trustCorporations, nil)

	err := ChooseReplacementTrustCorporation(nil, nil, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterReplacementTrustCorporation.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementTrustCorporationWhenReuseStoreError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return([]donordata.TrustCorporation{{}}, nil)
	reuseStore.EXPECT().
		PutTrustCorporation(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseReplacementTrustCorporation(nil, nil, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostChooseReplacementTrustCorporationWhenDonorStoreError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(r.Context()).
		Return([]donordata.TrustCorporation{{}}, nil)
	reuseStore.EXPECT().
		PutTrustCorporation(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseReplacementTrustCorporation(nil, donorStore, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
