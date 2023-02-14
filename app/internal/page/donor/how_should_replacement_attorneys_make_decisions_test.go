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

func TestGetHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App:  appData,
			Form: &howShouldAttorneysMakeDecisionsForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: "some decisions", HowReplacementAttorneysMakeDecisions: "jointly"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "jointly",
				DecisionsDetails: "some decisions",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
		}).
		Return(expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: "jointly"}).
		Return(nil)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
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
			form := url.Values{
				"decision-type": {tc.formType},
				"mixed-details": {tc.formDetails},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: tc.existingDetails, HowReplacementAttorneysMakeDecisions: tc.existingType}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: tc.updatedDetails, HowReplacementAttorneysMakeDecisions: tc.updatedType}).
				Return(nil)

			template := &mockTemplate{}

			err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {"some decisions"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"decision-type": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: ""}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App:    appData,
			Errors: validation.With("decision-type", validation.SelectError{Label: "howReplacementAttorneysShouldMakeDecisions"}),
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
				errorLabel:       "howReplacementAttorneysShouldMakeDecisions",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: "jointly"}).
		Return(expectedError)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}
