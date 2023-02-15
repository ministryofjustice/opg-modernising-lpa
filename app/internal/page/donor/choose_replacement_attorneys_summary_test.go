package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementAttorneysSummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysSummaryData{
			App:  page.TestAppData,
			Lpa:  &page.Lpa{},
			Form: &chooseAttorneysSummaryForm{},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(nil, template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseReplacementAttorneySummaryWhenStoreErrors(t *testing.T) {
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

	err := ChooseReplacementAttorneysSummary(logger, nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChooseReplacementAttorneysSummaryAddAttorney(t *testing.T) {
	form := url.Values{
		"add-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{}}, nil)

	err := ChooseReplacementAttorneysSummary(nil, nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneys+"?addAnother=1", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseReplacementAttorneysSummaryDoNotAddAttorney(t *testing.T) {
	attorney1 := actor.Attorney{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}
	attorney2 := actor.Attorney{FirstNames: "x", LastName: "y", Address: place.Address{Line1: "z"}, DateOfBirth: date.New("2000", "1", "1")}

	testcases := map[string]struct {
		expectedUrl          string
		Attorneys            actor.Attorneys
		ReplacementAttorneys actor.Attorneys
		HowAttorneysAct      string
		DecisionDetails      string
	}{
		"with multiple attorneys acting jointly and severally and single replacement attorney": {
			expectedUrl:          page.Paths.HowShouldReplacementAttorneysStepIn,
			Attorneys:            actor.Attorneys{attorney1, attorney2},
			ReplacementAttorneys: actor.Attorneys{attorney1},
			HowAttorneysAct:      page.JointlyAndSeverally,
		},
		"with multiple attorneys acting jointly and severally and multiple replacement attorney": {
			expectedUrl:          page.Paths.HowShouldReplacementAttorneysStepIn,
			Attorneys:            actor.Attorneys{attorney1, attorney2},
			ReplacementAttorneys: actor.Attorneys{attorney1, attorney2},
			HowAttorneysAct:      page.JointlyAndSeverally,
		},
		"with multiple attorneys acting jointly for some decisions and jointly and severally for other decisions and single replacement attorney": {
			expectedUrl:          page.Paths.WhenCanTheLpaBeUsed,
			Attorneys:            actor.Attorneys{attorney1, attorney2},
			ReplacementAttorneys: actor.Attorneys{attorney1},
			HowAttorneysAct:      page.JointlyForSomeSeverallyForOthers,
			DecisionDetails:      "some words",
		},
		"with multiple attorneys acting jointly for some decisions, and jointly and severally for other decisions and multiple replacement attorneys": {
			expectedUrl:          page.Paths.WhenCanTheLpaBeUsed,
			Attorneys:            actor.Attorneys{attorney1, attorney2},
			ReplacementAttorneys: actor.Attorneys{attorney1, attorney2},
			HowAttorneysAct:      page.JointlyForSomeSeverallyForOthers,
			DecisionDetails:      "some words",
		},
		"with multiple attorneys acting jointly and single replacement attorneys": {
			expectedUrl:          page.Paths.WhenCanTheLpaBeUsed,
			Attorneys:            actor.Attorneys{attorney1, attorney2},
			ReplacementAttorneys: actor.Attorneys{attorney1},
			HowAttorneysAct:      page.Jointly,
		},
		"with multiple attorneys acting jointly and multiple replacement attorneys": {
			expectedUrl:          page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			Attorneys:            actor.Attorneys{attorney1, attorney2},
			ReplacementAttorneys: actor.Attorneys{attorney1, attorney2},
			HowAttorneysAct:      page.Jointly,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"add-attorney": {"no"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ReplacementAttorneys:             tc.ReplacementAttorneys,
					HowAttorneysMakeDecisions:        tc.HowAttorneysAct,
					HowAttorneysMakeDecisionsDetails: tc.DecisionDetails,
					Attorneys:                        tc.Attorneys,
					Tasks: page.Tasks{
						YourDetails:     page.TaskCompleted,
						ChooseAttorneys: page.TaskCompleted,
					},
				}, nil)

			err := ChooseReplacementAttorneysSummary(nil, nil, lpaStore)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.expectedUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
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

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	validationError := validation.With("add-attorney", validation.SelectError{Label: "yesToAddAnotherReplacementAttorney"})

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *chooseReplacementAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(nil, template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
