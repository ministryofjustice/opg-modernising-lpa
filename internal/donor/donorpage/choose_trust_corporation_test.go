package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	trustCorporations := []donordata.TrustCorporation{{Name: "Corp"}}
	provided := &donordata.Provided{LpaID: "lpa-id"}

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), provided).
		Return(trustCorporations, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseTrustCorporationData{
			App:                 testAppData,
			Form:                &chooseTrustCorporationForm{},
			Donor:               &donordata.Provided{LpaID: "lpa-id"},
			TrustCorporations:   trustCorporations,
			ChooseAttorneysPath: donor.PathEnterAttorney.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}),
		}).
		Return(nil)

	err := ChooseTrustCorporation(template.Execute, service, testUIDFn)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseTrustCorporationWhenNoReusableTrustCorporations(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), mock.Anything).
		Return(nil, nil)

	err := ChooseTrustCorporation(nil, service, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterTrustCorporation.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChooseTrustCorporationWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := ChooseTrustCorporation(nil, service, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseTrustCorporationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), mock.Anything).
		Return([]donordata.TrustCorporation{{Name: "Corp"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseTrustCorporation(template.Execute, service, testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseTrustCorporation(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	trustCorporations := []donordata.TrustCorporation{{Name: "Corp"}, {Name: "Trust", Address: place.Address{Line1: "1"}}}

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), mock.Anything).
		Return(trustCorporations, nil)
	service.EXPECT().
		PutTrustCorporation(r.Context(), &donordata.Provided{LpaID: "lpa-id"}, donordata.TrustCorporation{
			Name:    "Trust",
			Address: place.Address{Line1: "1"},
		}).
		Return(nil)

	err := ChooseTrustCorporation(nil, service, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseTrustCorporationWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	trustCorporations := []donordata.TrustCorporation{{Name: "Corp"}, {Name: "Trust"}}

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), mock.Anything).
		Return(trustCorporations, nil)

	err := ChooseTrustCorporation(nil, service, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterTrustCorporation.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseTrustCorporationWhenReuseStoreError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := testAttorneyService(t)
	service.EXPECT().
		ReusableTrustCorporations(r.Context(), mock.Anything).
		Return([]donordata.TrustCorporation{{}}, nil)
	service.EXPECT().
		PutTrustCorporation(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseTrustCorporation(nil, service, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestReadChooseTrustCorporationForm(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseTrustCorporationForm(r)

	assert.False(t, result.New)
	assert.Equal(t, 1, result.Index)
	assert.Nil(t, result.Err)
}

func TestReadChooseTrustCorporationFormWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseTrustCorporationForm(r)

	assert.True(t, result.New)
	assert.NotNil(t, result.Err)
}

func TestChooseTrustCorporationFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *chooseTrustCorporationForm
		errors validation.List
	}{
		"new": {
			form: &chooseTrustCorporationForm{New: true, Err: expectedError},
		},
		"index": {
			form: &chooseTrustCorporationForm{Index: 1},
		},
		"error": {
			form:   &chooseTrustCorporationForm{Err: expectedError},
			errors: validation.With("option", validation.SelectError{Label: "aTrustCorporationOrToAddANewTrustCorporation"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
