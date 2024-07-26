package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysSummary(t *testing.T) {
	testcases := map[string]*actor.DonorProvidedDetails{
		"attorney": {
			Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
		},
		"trust corporation": {
			Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
		},
	}

	for name, donor := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &chooseAttorneysSummaryData{
					App:   testAppData,
					Donor: donor,
					Form:  form.NewYesNoForm(form.YesNoUnknown),
				}).
				Return(nil)

			err := ChooseAttorneysSummary(template.Execute, nil)(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseAttorneysSummaryWhenNoAttorneysOrTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysSummary(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
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
		"add attorney - no attorneys": {
			addMoreFormValue: form.Yes,
			expectedUrl:      page.Paths.ChooseAttorneys.Format("lpa-id") + "?addAnother=1",
			Attorneys:        actor.Attorneys{Attorneys: []actor.Attorney{}},
		},
		"add attorney - with attorney": {
			addMoreFormValue: form.Yes,
			expectedUrl:      page.Paths.ChooseAttorneys.Format("lpa-id") + "?addAnother=1&id=" + testUID.String(),
			Attorneys:        actor.Attorneys{Attorneys: []actor.Attorney{{UID: actoruid.New()}}},
		},
		"do not add attorney - with single attorney": {
			addMoreFormValue: form.No,
			expectedUrl:      page.Paths.TaskList.Format("lpa-id"),
			Attorneys:        actor.Attorneys{Attorneys: []actor.Attorney{{UID: actoruid.New()}}},
		},
		"do not add attorney - with multiple attorneys": {
			addMoreFormValue: form.No,
			expectedUrl:      page.Paths.HowShouldAttorneysMakeDecisions.Format("lpa-id"),
			Attorneys:        actor.Attorneys{Attorneys: []actor.Attorney{{UID: actoruid.New()}, {UID: actoruid.New()}}},
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {tc.addMoreFormValue.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := ChooseAttorneysSummary(nil, testUIDFn)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", Attorneys: tc.Attorneys})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseAttorneysSummaryFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToAddAnotherAttorney"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *chooseAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseAttorneysSummary(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
