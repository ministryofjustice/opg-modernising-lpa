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

func TestGetChooseAttorneysSummary(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysSummaryData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysSummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseAttorneysSummaryWhenStoreErrors(t *testing.T) {
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

	err := ChooseAttorneysSummary(logger, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChooseAttorneysSummaryAddAttorney(t *testing.T) {
	testcases := map[string]struct {
		addMoreFormValue string
		expectedUrl      string
		Attorneys        []Attorney
	}{
		"add attorney": {
			addMoreFormValue: "yes",
			expectedUrl:      "/choose-attorneys?addAnother=1",
			Attorneys:        []Attorney{},
		},
		"do not add attorney - with single attorney": {
			addMoreFormValue: "no",
			expectedUrl:      "/do-you-want-replacement-attorneys",
			Attorneys:        []Attorney{{ID: "123"}},
		},
		"do not add attorney - with multiple attorneys": {
			addMoreFormValue: "no",
			expectedUrl:      "/how-should-attorneys-make-decisions",
			Attorneys:        []Attorney{{ID: "123"}, {ID: "456"}},
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{Attorneys: tc.Attorneys}, nil)

			form := url.Values{
				"add-attorney": {tc.addMoreFormValue},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChooseAttorneysSummary(nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseAttorneysSummaryFormValidation(t *testing.T) {
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
		On("Func", w, mock.MatchedBy(func(data *chooseAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"add-attorney": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseAttorneysSummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestChooseAttorneysSummaryFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *chooseAttorneysSummaryForm
		errors map[string]string
	}{
		"yes": {
			form: &chooseAttorneysSummaryForm{
				AddAttorney: "yes",
			},
			errors: map[string]string{},
		},
		"no": {
			form: &chooseAttorneysSummaryForm{
				AddAttorney: "no",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &chooseAttorneysSummaryForm{},
			errors: map[string]string{
				"add-attorney": "selectAddMoreAttorneys",
			},
		},
		"invalid": {
			form: &chooseAttorneysSummaryForm{
				AddAttorney: "what",
			},
			errors: map[string]string{
				"add-attorney": "selectAddMoreAttorneys",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
