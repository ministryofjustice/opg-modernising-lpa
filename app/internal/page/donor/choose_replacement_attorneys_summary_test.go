package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementAttorneysSummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{ReplacementAttorneys: actor.Attorneys{{}}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysSummaryData{
			App:     testAppData,
			Lpa:     lpa,
			Form:    &chooseAttorneysSummaryForm{},
			Options: actor.YesNoValues,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(template.Execute)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysSummaryWhenNoReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{
		ID:    "lpa-id",
		Tasks: page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.DoYouWantReplacementAttorneys.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysSummaryAddAttorney(t *testing.T) {
	form := url.Values{
		"add-attorney": {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := ChooseReplacementAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", ReplacementAttorneys: actor.Attorneys{{}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneys.Format("lpa-id")+"?addAnother=1", resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysSummaryDoNotAddAttorney(t *testing.T) {
	attorney1 := actor.Attorney{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}
	attorney2 := actor.Attorney{FirstNames: "x", LastName: "y", Address: place.Address{Line1: "z"}, DateOfBirth: date.New("2000", "1", "1")}

	testcases := map[string]struct {
		redirectUrl          page.LpaPath
		attorneys            actor.Attorneys
		replacementAttorneys actor.Attorneys
		howAttorneysAct      actor.AttorneysAct
		decisionDetails      string
	}{
		"with multiple attorneys acting jointly and severally and single replacement attorney": {
			redirectUrl:          page.Paths.HowShouldReplacementAttorneysStepIn,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1},
			howAttorneysAct:      actor.JointlyAndSeverally,
		},
		"with multiple attorneys acting jointly and severally and multiple replacement attorney": {
			redirectUrl:          page.Paths.HowShouldReplacementAttorneysStepIn,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1, attorney2},
			howAttorneysAct:      actor.JointlyAndSeverally,
		},
		"with multiple attorneys acting jointly and multiple replacement attorneys": {
			redirectUrl:          page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1, attorney2},
			howAttorneysAct:      actor.Jointly,
		},
		"with multiple attorneys acting jointly for some decisions and jointly and severally for other decisions and single replacement attorney": {
			redirectUrl:          page.Paths.TaskList,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1},
			howAttorneysAct:      actor.JointlyForSomeSeverallyForOthers,
			decisionDetails:      "some words",
		},
		"with multiple attorneys acting jointly for some decisions, and jointly and severally for other decisions and multiple replacement attorneys": {
			redirectUrl:          page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1, attorney2},
			howAttorneysAct:      actor.JointlyForSomeSeverallyForOthers,
			decisionDetails:      "some words",
		},
		"with multiple attorneys acting jointly and single replacement attorneys": {
			redirectUrl:          page.Paths.TaskList,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1},
			howAttorneysAct:      actor.Jointly,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"add-attorney": {actor.No.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := ChooseReplacementAttorneysSummary(nil)(testAppData, w, r, &page.Lpa{
				ID:                   "lpa-id",
				ReplacementAttorneys: tc.replacementAttorneys,
				AttorneyDecisions: actor.AttorneyDecisions{
					How:     tc.howAttorneysAct,
					Details: tc.decisionDetails,
				},
				Attorneys: tc.attorneys,
				Tasks: page.Tasks{
					YourDetails:     actor.TaskCompleted,
					ChooseAttorneys: actor.TaskCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirectUrl.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseReplacementAttorneySummaryFormValidation(t *testing.T) {
	form := url.Values{
		"add-attorney": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("add-attorney", validation.SelectError{Label: "yesToAddAnotherReplacementAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *chooseReplacementAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(template.Execute)(testAppData, w, r, &page.Lpa{ReplacementAttorneys: actor.Attorneys{{}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
