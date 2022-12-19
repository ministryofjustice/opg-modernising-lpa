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

func TestGetHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App:  appData,
			Form: &howShouldAttorneysMakeDecisionsForm{},
			Lpa:  &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowAttorneysMakeDecisionsDetails: "some decisions", HowAttorneysMakeDecisions: "jointly"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "jointly",
				DecisionsDetails: "some decisions",
			},
			Lpa: &Lpa{HowAttorneysMakeDecisionsDetails: "some decisions", HowAttorneysMakeDecisions: "jointly"},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: "jointly"}).
		Return(nil)

	template := &mockTemplate{}

	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
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
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{HowAttorneysMakeDecisionsDetails: tc.existingDetails, HowAttorneysMakeDecisions: tc.existingType}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{HowAttorneysMakeDecisionsDetails: tc.updatedDetails, HowAttorneysMakeDecisions: tc.updatedType}).
				Return(nil)

			template := &mockTemplate{}

			form := url.Values{
				"decision-type": {tc.formType},
				"mixed-details": {tc.formDetails},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, appData.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostHowShouldAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
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

	err := HowShouldAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowShouldAttorneysMakeDecisionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Errors: map[string]string{
				"decision-type": "chooseADecisionType",
			},
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
			Lpa: &Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""},
		}).
		Return(nil)

	form := url.Values{
		"decision-type": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestValidateForm(t *testing.T) {
	testCases := map[string]struct {
		DecisionType   string
		DecisionDetail string
		ExpectedErrors map[string]string
	}{
		"valid": {
			DecisionType:   "jointly-and-severally",
			DecisionDetail: "",
			ExpectedErrors: map[string]string{},
		},
		"valid with detail": {
			DecisionType:   "mixed",
			DecisionDetail: "some details",
			ExpectedErrors: map[string]string{},
		},
		"unsupported decision type": {
			DecisionType:   "not-supported",
			DecisionDetail: "",
			ExpectedErrors: map[string]string{"decision-type": "chooseADecisionType"},
		},
		"missing decision type": {
			DecisionType:   "",
			DecisionDetail: "",
			ExpectedErrors: map[string]string{"decision-type": "chooseADecisionType"},
		},
		"missing decision detail when mixed": {
			DecisionType:   "mixed",
			DecisionDetail: "",
			ExpectedErrors: map[string]string{"mixed-details": "provideDecisionDetails"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    tc.DecisionType,
				DecisionsDetails: tc.DecisionDetail,
			}

			assert.Equal(t, tc.ExpectedErrors, form.Validate())
		})
	}
}

func TestPostHowShouldAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: ""}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{HowAttorneysMakeDecisionsDetails: "", HowAttorneysMakeDecisions: "jointly"}).
		Return(expectedError)

	template := &mockTemplate{}

	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}
