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

func TestGetChooseReplacementAttorneysSummary(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysSummaryData{
			App:                            appData,
			Lpa:                            &Lpa{},
			ReplacementAttorneyAddressPath: chooseReplacementAttorneysAddressPath,
			ReplacementAttorneyDetailsPath: chooseReplacementAttorneysPath,
			RemoveReplacementAttorneyPath:  removeReplacementAttorneyPath,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneysSummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseReplacementAttorneySummaryWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	logger := &mockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneysSummary(logger, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChooseReplacementAttorneysSummaryAddAttorney(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{}}, nil)

	form := url.Values{
		"add-attorney": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseReplacementAttorneysSummary(nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys?addAnother=1", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseReplacementAttorneysSummaryDoNotAddAttorney(t *testing.T) {
	testcases := map[string]struct {
		expectedUrl          string
		Attorneys            []Attorney
		ReplacementAttorneys []Attorney
		HowAttorneysAct      string
	}{
		"with multiple attorneys acting jointly and severally and single replacement attorney": {
			expectedUrl:          "/how-should-replacement-attorneys-step-in",
			Attorneys:            []Attorney{{ID: "123"}, {ID: "456"}},
			ReplacementAttorneys: []Attorney{{ID: "123"}},
			HowAttorneysAct:      "jointly-and-severally",
		},
		"with multiple attorneys acting jointly and severally and multiple replacement attorney": {
			expectedUrl:          "/how-should-replacement-attorneys-step-in",
			Attorneys:            []Attorney{{ID: "123"}, {ID: "456"}},
			ReplacementAttorneys: []Attorney{{ID: "123"}, {ID: "456"}},
			HowAttorneysAct:      "jointly-and-severally",
		},
		"with multiple attorneys acting jointly for some decisions and jointly and severally for other decisions and single replacement attorney": {
			expectedUrl:          "/task-list",
			Attorneys:            []Attorney{{ID: "123"}, {ID: "456"}},
			ReplacementAttorneys: []Attorney{{ID: "123"}},
			HowAttorneysAct:      "mixed",
		},
		"with multiple attorneys acting jointly for some decisions, and jointly and severally for other decisions and multiple replacement attorneys": {
			expectedUrl:          "/task-list",
			Attorneys:            []Attorney{{ID: "123"}, {ID: "456"}},
			ReplacementAttorneys: []Attorney{{ID: "123"}, {ID: "123"}},
			HowAttorneysAct:      "mixed",
		},
		"with multiple attorneys acting jointly and single replacement attorneys": {
			expectedUrl:          "/task-list",
			Attorneys:            []Attorney{{ID: "123"}, {ID: "456"}},
			ReplacementAttorneys: []Attorney{{ID: "123"}},
			HowAttorneysAct:      "jointly",
		},
		"with multiple attorneys acting jointly and multiple replacement attorneys": {
			expectedUrl:          "/how-should-replacement-attorneys-make-decisions",
			Attorneys:            []Attorney{{ID: "123"}, {ID: "456"}},
			ReplacementAttorneys: []Attorney{{ID: "123"}, {ID: "123"}},
			HowAttorneysAct:      "jointly",
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					ReplacementAttorneys:      tc.ReplacementAttorneys,
					HowAttorneysMakeDecisions: tc.HowAttorneysAct,
					Attorneys:                 tc.Attorneys,
				}, nil)

			form := url.Values{
				"add-attorney": {"no"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseReplacementAttorneysSummary(nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseReplacementAttorneySummaryFormValidation(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	validationError := map[string]string{
		"add-attorney": "selectAddMoreAttorneys",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *chooseReplacementAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"add-attorney": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseReplacementAttorneysSummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
