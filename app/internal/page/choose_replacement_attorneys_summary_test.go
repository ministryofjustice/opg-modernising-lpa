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
	testcases := map[string]struct {
		addMoreFormValue string
		expectedUrl      string
		Attorneys        []Attorney
	}{
		"add attorney": {
			addMoreFormValue: "yes",
			expectedUrl:      "/choose-replacement-attorneys?addAnother=1",
			Attorneys:        []Attorney{},
		},
		"do not add attorney - with single attorney": {
			addMoreFormValue: "no",
			expectedUrl:      "/how-should-replacement-attorneys-step-in",
			Attorneys:        []Attorney{{ID: "123"}},
		},
		"do not add attorney - with multiple attorneys": {
			addMoreFormValue: "no",
			expectedUrl:      "/how-should-replacement-attorneys-step-in",
			Attorneys:        []Attorney{{ID: "123"}, {ID: "456"}},
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{ReplacementAttorneys: tc.Attorneys}, nil)

			form := url.Values{
				"add-attorney": {tc.addMoreFormValue},
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
