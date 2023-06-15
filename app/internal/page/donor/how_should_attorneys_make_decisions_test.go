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

func TestGetHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldAttorneysMakeDecisionsData{
			App:  testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{},
			Lpa:  &page.Lpa{},
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldAttorneysMakeDecisionsData{
			App: testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "jointly",
				DecisionsDetails: "some decisions",
			},
			Lpa: &page.Lpa{AttorneyDecisions: actor.AttorneyDecisions{Details: "some decisions", How: "jointly"}},
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{AttorneyDecisions: actor.AttorneyDecisions{Details: "some decisions", How: "jointly"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldAttorneysMakeDecisionsData{
			App: testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
			Lpa: &page.Lpa{},
		}).
		Return(expectedError)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldAttorneysMakeDecisions(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly-and-severally"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Attorneys:         actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
			AttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			Tasks:             page.Tasks{ChooseAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{Attorneys: actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingType    string
		existingDetails string
		updatedType     string
		updatedDetails  string
		formType        string
		formDetails     string
	}{
		"existing details not set": {
			existingType:    "jointly-and-severally",
			existingDetails: "",
			updatedType:     "mixed",
			updatedDetails:  "some details",
			formType:        "mixed",
			formDetails:     "some details",
		},
		"existing details set": {
			existingType:    "mixed",
			existingDetails: "some details",
			updatedType:     "jointly",
			updatedDetails:  "",
			formType:        "jointly",
			formDetails:     "some details",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"decision-type": {tc.formType},
				"mixed-details": {tc.formDetails},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					Attorneys:         actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
					AttorneyDecisions: actor.AttorneyDecisions{Details: tc.updatedDetails, How: tc.updatedType},
					Tasks:             page.Tasks{ChooseAttorneys: actor.TaskInProgress},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{
				Attorneys:         actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
				AttorneyDecisions: actor.AttorneyDecisions{Details: tc.existingDetails, How: tc.existingType},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.AreYouHappyIfOneAttorneyCantActNoneCan, resp.Header.Get("Location"))
		})
	}
}

func TestPostHowShouldAttorneysMakeDecisionsWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"decision-type": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldAttorneysMakeDecisionsData{
			App:    testAppData,
			Errors: validation.With("decision-type", validation.SelectError{Label: "howAttorneysShouldMakeDecisions"}),
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
				errorLabel:       "howAttorneysShouldMakeDecisions",
			},
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHowShouldAttorneysMakeDecisionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		DecisionType   string
		DecisionDetail string
		ExpectedErrors validation.List
	}{
		"valid": {
			DecisionType:   "jointly-and-severally",
			DecisionDetail: "",
		},
		"valid with detail": {
			DecisionType:   "mixed",
			DecisionDetail: "some details",
		},
		"unsupported decision type": {
			DecisionType:   "not-supported",
			DecisionDetail: "",
			ExpectedErrors: validation.With("decision-type", validation.SelectError{Label: "xyz"}),
		},
		"missing decision type": {
			DecisionType:   "",
			DecisionDetail: "",
			ExpectedErrors: validation.With("decision-type", validation.SelectError{Label: "xyz"}),
		},
		"missing decision detail when mixed": {
			DecisionType:   "mixed",
			DecisionDetail: "",
			ExpectedErrors: validation.With("mixed-details", validation.EnterError{Label: "details"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    tc.DecisionType,
				DecisionsDetails: tc.DecisionDetail,
				errorLabel:       "xyz",
			}

			assert.Equal(t, tc.ExpectedErrors, form.Validate())
		})
	}
}

func TestPostHowShouldAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly-and-severally"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
