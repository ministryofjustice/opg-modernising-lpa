package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App:  page.TestAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{},
			Lpa:  &page.Lpa{},
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowAttorneysMakeDecisionsDetails: "some decisions", HowAttorneysMakeDecisions: "jointly"}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: page.TestAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "jointly",
				DecisionsDetails: "some decisions",
			},
			Lpa: &page.Lpa{HowAttorneysMakeDecisionsDetails: "some decisions", HowAttorneysMakeDecisions: "jointly"},
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := HowShouldAttorneysMakeDecisions(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: page.TestAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
			Lpa: &page.Lpa{},
		}).
		Return(page.ExpectedError)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldAttorneysMakeDecisions(t *testing.T) {
	f := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: "jointly"}).
		Return(nil)

	template := &page.MockTemplate{}

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingType    string
		existingDetails string
		updatedType     string
		updatedDetails  string
		formType        string
		formDetails     string
	}{
		"existing details not set": {
			existingType:    "jointly-and-severally",
			existingDetails: "",
			updatedType:     "mixed",
			updatedDetails:  "some details",
			formType:        "mixed",
			formDetails:     "some details",
		},
		"existing details set": {
			existingType:    "mixed",
			existingDetails: "some details",
			updatedType:     "jointly",
			updatedDetails:  "",
			formType:        "jointly",
			formDetails:     "some details",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"decision-type": {tc.formType},
				"mixed-details": {tc.formDetails},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{HowAttorneysMakeDecisionsDetails: tc.existingDetails, HowAttorneysMakeDecisions: tc.existingType}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{HowAttorneysMakeDecisionsDetails: tc.updatedDetails, HowAttorneysMakeDecisions: tc.updatedType}).
				Return(nil)

			template := &page.MockTemplate{}

			err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostHowShouldAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {"some decisions"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := HowShouldAttorneysMakeDecisions(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowShouldAttorneysMakeDecisionsWhenValidationErrors(t *testing.T) {
	f := url.Values{
		"decision-type": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App:    page.TestAppData,
			Errors: validation.With("decision-type", validation.SelectError{Label: "howAttorneysShouldMakeDecisions"}),
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
				errorLabel:       "howAttorneysShouldMakeDecisions",
			},
			Lpa: &page.Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""},
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestHowShouldAttorneysMakeDecisionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		DecisionType   string
		DecisionDetail string
		ExpectedErrors validation.List
	}{
		"valid": {
			DecisionType:   "jointly-and-severally",
			DecisionDetail: "",
		},
		"valid with detail": {
			DecisionType:   "mixed",
			DecisionDetail: "some details",
		},
		"unsupported decision type": {
			DecisionType:   "not-supported",
			DecisionDetail: "",
			ExpectedErrors: validation.With("decision-type", validation.SelectError{Label: "xyz"}),
		},
		"missing decision type": {
			DecisionType:   "",
			DecisionDetail: "",
			ExpectedErrors: validation.With("decision-type", validation.SelectError{Label: "xyz"}),
		},
		"missing decision detail when mixed": {
			DecisionType:   "mixed",
			DecisionDetail: "",
			ExpectedErrors: validation.With("mixed-details", validation.EnterError{Label: "details"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    tc.DecisionType,
				DecisionsDetails: tc.DecisionDetail,
				errorLabel:       "xyz",
			}

			assert.Equal(t, tc.ExpectedErrors, f.Validate())
		})
	}
}

func TestPostHowShouldAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	f := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: "jointly"}).
		Return(page.ExpectedError)

	template := &page.MockTemplate{}

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}
