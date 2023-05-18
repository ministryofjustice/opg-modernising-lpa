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

	lpa := &page.Lpa{ReplacementAttorneys: actor.Attorneys{{}}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysSummaryData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &chooseAttorneysSummaryForm{},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(nil, template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysSummaryWhenNoReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	err := ChooseReplacementAttorneysSummary(nil, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.DoYouWantReplacementAttorneys, resp.Header.Get("Location"))
}

func TestGetChooseReplacementAttorneySummaryWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{{}}}, expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	err := ChooseReplacementAttorneysSummary(logger, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysSummaryAddAttorney(t *testing.T) {
	form := url.Values{
		"add-attorney": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{{}}}, nil)

	err := ChooseReplacementAttorneysSummary(nil, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneys+"?addAnother=1", resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysSummaryDoNotAddAttorney(t *testing.T) {
	attorney1 := actor.Attorney{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}
	attorney2 := actor.Attorney{FirstNames: "x", LastName: "y", Address: place.Address{Line1: "z"}, DateOfBirth: date.New("2000", "1", "1")}

	testcases := map[string]struct {
		redirectUrl          string
		attorneys            actor.Attorneys
		replacementAttorneys actor.Attorneys
		howAttorneysAct      string
		decisionDetails      string
		lpaType              string
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
			redirectUrl:          page.Paths.WhenCanTheLpaBeUsed,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1},
			howAttorneysAct:      actor.JointlyForSomeSeverallyForOthers,
			decisionDetails:      "some words",
			lpaType:              page.LpaTypePropertyFinance,
		},
		"with multiple attorneys acting jointly for some decisions, and jointly and severally for other decisions and multiple replacement attorneys": {
			redirectUrl:          page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1, attorney2},
			howAttorneysAct:      actor.JointlyForSomeSeverallyForOthers,
			decisionDetails:      "some words",
			lpaType:              page.LpaTypePropertyFinance,
		},
		"pfa with multiple attorneys acting jointly and single replacement attorneys": {
			redirectUrl:          page.Paths.WhenCanTheLpaBeUsed,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1},
			howAttorneysAct:      actor.Jointly,
			lpaType:              page.LpaTypePropertyFinance,
		},
		"hw with multiple attorneys acting jointly and single replacement attorneys": {
			redirectUrl:          page.Paths.LifeSustainingTreatment,
			attorneys:            actor.Attorneys{attorney1, attorney2},
			replacementAttorneys: actor.Attorneys{attorney1},
			howAttorneysAct:      actor.Jointly,
			lpaType:              page.LpaTypeHealthWelfare,
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

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Type:                 tc.lpaType,
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
				}, nil)

			err := ChooseReplacementAttorneysSummary(nil, nil, donorStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.redirectUrl, resp.Header.Get("Location"))
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{{}}}, nil)

	validationError := validation.With("add-attorney", validation.SelectError{Label: "yesToAddAnotherReplacementAttorney"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *chooseReplacementAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(nil, template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
