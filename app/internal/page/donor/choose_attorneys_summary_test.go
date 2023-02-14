package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysSummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &chooseAttorneysSummaryData{
			App:  page.TestAppData,
			Lpa:  &page.Lpa{},
			Form: &chooseAttorneysSummaryForm{},
		}).
		Return(nil)

	err := ChooseAttorneysSummary(nil, template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseAttorneysSummaryWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	logger := &page.MockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	err := ChooseAttorneysSummary(logger, nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChooseAttorneysSummaryAddAttorney(t *testing.T) {
	testcases := map[string]struct {
		addMoreFormValue string
		expectedUrl      string
		Attorneys        actor.Attorneys
	}{
		"add attorney": {
			addMoreFormValue: "yes",
			expectedUrl:      "/lpa/lpa-id" + page.Paths.ChooseAttorneys + "?addAnother=1",
			Attorneys:        actor.Attorneys{},
		},
		"do not add attorney - with single attorney": {
			addMoreFormValue: "no",
			expectedUrl:      "/lpa/lpa-id" + page.Paths.DoYouWantReplacementAttorneys,
			Attorneys:        actor.Attorneys{{ID: "123"}},
		},
		"do not add attorney - with multiple attorneys": {
			addMoreFormValue: "no",
			expectedUrl:      "/lpa/lpa-id" + page.Paths.HowShouldAttorneysMakeDecisions,
			Attorneys:        actor.Attorneys{{ID: "123"}, {ID: "456"}},
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"add-attorney": {tc.addMoreFormValue},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{Attorneys: tc.Attorneys}, nil)

			err := ChooseAttorneysSummary(nil, nil, lpaStore)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChooseAttorneysSummaryFormValidation(t *testing.T) {
	form := url.Values{
		"add-attorney": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	validationError := validation.With("add-attorney", validation.SelectError{Label: "yesToAddAnotherAttorney"})

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *chooseAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseAttorneysSummary(nil, template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestChooseAttorneysSummaryFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *chooseAttorneysSummaryForm
		errors validation.List
	}{
		"yes": {
			form: &chooseAttorneysSummaryForm{
				AddAttorney: "yes",
			},
		},
		"no": {
			form: &chooseAttorneysSummaryForm{
				AddAttorney: "no",
			},
		},
		"missing": {
			form:   &chooseAttorneysSummaryForm{errorLabel: "xyz"},
			errors: validation.With("add-attorney", validation.SelectError{Label: "xyz"}),
		},
		"invalid": {
			form: &chooseAttorneysSummaryForm{
				AddAttorney: "what",
				errorLabel:  "xyz",
			},
			errors: validation.With("add-attorney", validation.SelectError{Label: "xyz"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
