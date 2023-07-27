package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysSummary(t *testing.T) {
	testcases := map[string]*page.Lpa{
		"attorney": {
			Attorneys: actor.NewAttorneys(nil, []actor.Attorney{{}}),
		},
		"trust corporation": {
			Attorneys: actor.NewAttorneys(&actor.TrustCorporation{Name: "a"}, nil),
		},
	}

	for name, lpa := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &chooseAttorneysSummaryData{
					App:     testAppData,
					Lpa:     lpa,
					Form:    &form.YesNoForm{},
					Options: form.YesNoValues,
				}).
				Return(nil)

			err := ChooseAttorneysSummary(template.Execute)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseAttorneysSummaryWhenNoAttorneysOrTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneys.Format("lpa-id")+"?addAnother=1", resp.Header.Get("Location"))
}

func TestPostChooseAttorneysSummaryAddAttorney(t *testing.T) {
	testcases := map[string]struct {
		addMoreFormValue form.YesNo
		expectedUrl      string
		Attorneys        actor.Attorneys
	}{
		"add attorney": {
			addMoreFormValue: form.Yes,
			expectedUrl:      page.Paths.ChooseAttorneys.Format("lpa-id") + "?addAnother=1",
			Attorneys:        actor.NewAttorneys(nil, []actor.Attorney{}),
		},
		"do not add attorney - with single attorney": {
			addMoreFormValue: form.No,
			expectedUrl:      page.Paths.TaskList.Format("lpa-id"),
			Attorneys:        actor.NewAttorneys(nil, []actor.Attorney{{ID: "123"}}),
		},
		"do not add attorney - with multiple attorneys": {
			addMoreFormValue: form.No,
			expectedUrl:      page.Paths.HowShouldAttorneysMakeDecisions.Format("lpa-id"),
			Attorneys:        actor.NewAttorneys(nil, []actor.Attorney{{ID: "123"}, {ID: "456"}}),
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"yes-no": {tc.addMoreFormValue.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := ChooseAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Attorneys: tc.Attorneys})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseAttorneysSummaryFormValidation(t *testing.T) {
	form := url.Values{
		"yes-no": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("yes-no", validation.SelectError{Label: "yesToAddAnotherAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *chooseAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseAttorneysSummary(template.Execute)(testAppData, w, r, &page.Lpa{Attorneys: actor.NewAttorneys(nil, []actor.Attorney{{}})})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
