package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysSummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAttorneysSummaryData{
			App:  testAppData,
			Lpa:  &page.Lpa{Attorneys: actor.Attorneys{{}}},
			Form: &chooseAttorneysSummaryForm{},
		}).
		Return(nil)

	err := ChooseAttorneysSummary(template.Execute)(testAppData, w, r, &page.Lpa{Attorneys: actor.Attorneys{{}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysSummaryWhenNoAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseAttorneys+"?addAnother=1", resp.Header.Get("Location"))
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
			expectedUrl:      "/lpa/lpa-id" + page.Paths.TaskList,
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

			err := ChooseAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{Attorneys: tc.Attorneys})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
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

	validationError := validation.With("add-attorney", validation.SelectError{Label: "yesToAddAnotherAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *chooseAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseAttorneysSummary(template.Execute)(testAppData, w, r, &page.Lpa{Attorneys: actor.Attorneys{{}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
