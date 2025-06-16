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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseCorrespondent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	correspondents := []donordata.Correspondent{{FirstNames: "John"}}

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(correspondents, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseCorrespondentData{
			App:            testAppData,
			Form:           &chooseCorrespondentForm{},
			Donor:          &donordata.Provided{},
			Correspondents: correspondents,
		}).
		Return(nil)

	err := ChooseCorrespondent(template.Execute, service)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseCorrespondentWhenNoReusableCorrespondents(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(nil, nil)

	err := ChooseCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterCorrespondentDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChooseCorrespondentWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(nil, expectedError)

	err := ChooseCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseCorrespondentWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return([]donordata.Correspondent{{FirstNames: "John"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChooseCorrespondent(template.Execute, service)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseCorrespondent(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	correspondents := []donordata.Correspondent{{FirstNames: "John"}, {FirstNames: "Dave"}}

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(correspondents, nil)
	service.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:         "lpa-id",
			Correspondent: donordata.Correspondent{FirstNames: "Dave"},
		}).
		Return(nil)

	err := ChooseCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCorrespondentSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseCorrespondentWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	correspondents := []donordata.Correspondent{{FirstNames: "John"}, {FirstNames: "Dave"}}

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return(correspondents, nil)

	err := ChooseCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterCorrespondentDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseCorrespondentWhenServiceError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Reusable(r.Context()).
		Return([]donordata.Correspondent{{}}, nil)
	service.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestReadChooseCorrespondentForm(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseCorrespondentForm(r)

	assert.False(t, result.New)
	assert.Equal(t, 1, result.Index)
	assert.Nil(t, result.Err)
}

func TestReadChooseCorrespondentFormWhenNew(t *testing.T) {
	form := url.Values{
		"option": {"new"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChooseCorrespondentForm(r)

	assert.True(t, result.New)
	assert.NotNil(t, result.Err)
}

func TestChooseCorrespondentFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *chooseCorrespondentForm
		errors validation.List
	}{
		"new": {
			form: &chooseCorrespondentForm{New: true, Err: expectedError},
		},
		"index": {
			form: &chooseCorrespondentForm{Index: 1},
		},
		"error": {
			form:   &chooseCorrespondentForm{Err: expectedError},
			errors: validation.With("option", validation.SelectError{Label: "aCorrespondentOrToAddANewCorrespondent"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
