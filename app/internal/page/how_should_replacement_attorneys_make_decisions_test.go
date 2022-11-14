package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App:  appData,
			Form: &howShouldAttorneysMakeDecisionsForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowReplacementAttorneysMakeDecisionsDetails: "some decisions", HowReplacementAttorneysMakeDecisions: "jointly"}, nil)

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

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

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

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: "jointly"}).
		Return(nil)

	template := &mockTemplate{}

	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
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
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{HowReplacementAttorneysMakeDecisionsDetails: tc.existingDetails, HowReplacementAttorneysMakeDecisions: tc.existingType}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{HowReplacementAttorneysMakeDecisionsDetails: tc.updatedDetails, HowReplacementAttorneysMakeDecisions: tc.updatedType}).
				Return(nil)

			template := &mockTemplate{}

			form := url.Values{
				"decision-type": {tc.formType},
				"mixed-details": {tc.formDetails},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, taskListPath, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {"some decisions"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: ""}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Errors: map[string]string{
				"decision-type": "chooseADecisionType",
			},
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
		}).
		Return(nil)

	form := url.Values{
		"decision-type": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{HowReplacementAttorneysMakeDecisionsDetails: "", HowReplacementAttorneysMakeDecisions: "jointly"}).
		Return(expectedError)

	template := &mockTemplate{}

	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}
