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

func TestGetChooseAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	provided := &donordata.Provided{}
	attorneys := []donordata.Attorney{{FirstNames: "John"}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), provided).
		Return(attorneys, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAttorneysData{
			App:       testAppData,
			Form:      &chooseAttorneysForm{},
			Donor:     &donordata.Provided{},
			Attorneys: attorneys,
		}).
		Return(nil)

	err := ChooseAttorneys(template.Execute, nil, reuseStore, nil)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysWhenNoReusableAttorneys(t *testing.T) {
	testcases := map[string]error{
		"none":      nil,
		"not found": dynamo.NotFoundError{},
	}

	for name, reuseError := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			provided := &donordata.Provided{LpaID: "lpa-id"}

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				Attorneys(r.Context(), provided).
				Return(nil, reuseError)

			err := ChooseAttorneys(nil, nil, reuseStore, testUIDFn)(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathEnterAttorney.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}), resp.Header.Get("Location"))
		})
	}
}

func TestGetChooseAttorneysWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := ChooseAttorneys(nil, nil, reuseStore, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), mock.Anything).
		Return([]donordata.Attorney{{FirstNames: "John"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(template.Execute, nil, reuseStore, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneys(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneys := []donordata.Attorney{{FirstNames: "John"}, {FirstNames: "Dave", Address: place.Address{Line1: "123"}}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), mock.Anything).
		Return(attorneys, nil)
	reuseStore.EXPECT().
		PutAttorneys(r.Context(), []donordata.Attorney{{
			UID:        testUID,
			FirstNames: "Dave",
			Address:    place.Address{Line1: "123"},
		}}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Attorneys: donordata.Attorneys{
				Attorneys: []donordata.Attorney{{UID: testUID, FirstNames: "Dave", Address: place.Address{Line1: "123"}}},
			},
			Tasks: donordata.Tasks{ChooseAttorneys: task.StateCompleted},
		}).
		Return(nil)

	err := ChooseAttorneys(nil, donorStore, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseAttorneysWhenNoneSelected(t *testing.T) {
	form := url.Values{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), mock.Anything).
		Return([]donordata.Attorney{{}}, nil)

	err := ChooseAttorneys(nil, nil, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterAttorney.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}), resp.Header.Get("Location"))
}

func TestPostChooseAttorneysWhenReuseStoreError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), mock.Anything).
		Return([]donordata.Attorney{{}}, nil)
	reuseStore.EXPECT().
		PutAttorneys(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(nil, nil, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostChooseAttorneysWhenDonorStoreError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(r.Context(), mock.Anything).
		Return([]donordata.Attorney{{}}, nil)
	reuseStore.EXPECT().
		PutAttorneys(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseAttorneys(nil, donorStore, reuseStore, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestReadChooseAttorneysForm(t *testing.T) {
	form := url.Values{
		"option": {"1", "2"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseAttorneysForm(r)

	assert.Equal(t, []int{1, 2}, result.Indices)
}
