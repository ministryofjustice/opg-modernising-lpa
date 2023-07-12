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
			App:     testAppData,
			Form:    &howShouldAttorneysMakeDecisionsForm{},
			Lpa:     &page.Lpa{},
			Options: actor.AttorneysActValues,
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
				DecisionsType:    actor.Jointly,
				DecisionsDetails: "some decisions",
			},
			Lpa:     &page.Lpa{AttorneyDecisions: actor.AttorneyDecisions{Details: "some decisions", How: actor.Jointly}},
			Options: actor.AttorneysActValues,
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{AttorneyDecisions: actor.AttorneyDecisions{Details: "some decisions", How: actor.Jointly}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldAttorneysMakeDecisions(t *testing.T) {
	form := url.Values{
		"decision-type": {actor.JointlyAndSeverally.String()},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                "lpa-id",
			Attorneys:         actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
			AttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			Tasks:             page.Tasks{ChooseAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Attorneys: actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingType    actor.AttorneysAct
		existingDetails string
		updatedType     actor.AttorneysAct
		updatedDetails  string
		formType        string
		formDetails     string
	}{
		"existing details not set": {
			existingType:    actor.JointlyAndSeverally,
			existingDetails: "",
			updatedType:     actor.JointlyForSomeSeverallyForOthers,
			updatedDetails:  "some details",
			formType:        actor.JointlyForSomeSeverallyForOthers.String(),
			formDetails:     "some details",
		},
		"existing details set": {
			existingType:    actor.JointlyForSomeSeverallyForOthers,
			existingDetails: "some details",
			updatedType:     actor.Jointly,
			updatedDetails:  "",
			formType:        actor.Jointly.String(),
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
					ID:                "lpa-id",
					Attorneys:         actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
					AttorneyDecisions: actor.AttorneyDecisions{Details: tc.updatedDetails, How: tc.updatedType},
					Tasks:             page.Tasks{ChooseAttorneys: actor.TaskCompleted},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{
				ID:                "lpa-id",
				Attorneys:         actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
				AttorneyDecisions: actor.AttorneyDecisions{Details: tc.existingDetails, How: tc.existingType},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
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
		On("Execute", w, mock.MatchedBy(func(data *howShouldAttorneysMakeDecisionsData) bool {
			return assert.Equal(t, validation.With("decision-type", validation.SelectError{Label: "howAttorneysShouldMakeDecisions"}), data.Errors)
		})).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"decision-type": {actor.JointlyAndSeverally.String()},
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

func TestHowShouldAttorneysMakeDecisionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howShouldAttorneysMakeDecisionsForm
		errors validation.List
	}{
		"valid": {
			form: &howShouldAttorneysMakeDecisionsForm{
				errorLabel: "xyz",
			},
		},
		"valid with detail": {
			form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    actor.JointlyForSomeSeverallyForOthers,
				DecisionsDetails: "some details",
				errorLabel:       "xyz",
			},
		},
		"invalid": {
			form: &howShouldAttorneysMakeDecisionsForm{
				Error:      expectedError,
				errorLabel: "xyz",
			},
			errors: validation.With("decision-type", validation.SelectError{Label: "xyz"}),
		},
		"missing decision detail when mixed": {
			form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType: actor.JointlyForSomeSeverallyForOthers,
				errorLabel:    "xyz",
			},
			errors: validation.With("mixed-details", validation.EnterError{Label: "details"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
