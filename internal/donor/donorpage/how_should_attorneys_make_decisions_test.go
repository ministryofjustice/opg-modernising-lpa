package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howShouldAttorneysMakeDecisionsData{
			App:     testAppData,
			Form:    &howShouldAttorneysMakeDecisionsForm{},
			Donor:   &donordata.Provided{},
			Options: donordata.AttorneysActValues,
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howShouldAttorneysMakeDecisionsData{
			App: testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    donordata.Jointly,
				DecisionsDetails: "some decisions",
			},
			Donor:   &donordata.Provided{AttorneyDecisions: donordata.AttorneyDecisions{Details: "some decisions", How: donordata.Jointly}},
			Options: donordata.AttorneysActValues,
		}).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{AttorneyDecisions: donordata.AttorneyDecisions{Details: "some decisions", How: donordata.Jointly}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldAttorneysMakeDecisions(t *testing.T) {
	form := url.Values{
		"decision-type": {donordata.JointlyAndSeverally.String()},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneys := donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", Address: testAddress}, {FirstNames: "b", Address: testAddress}}}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:             "lpa-id",
			Attorneys:         attorneys,
			AttorneyDecisions: donordata.AttorneyDecisions{How: donordata.JointlyAndSeverally},
			Tasks:             donordata.Tasks{ChooseAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Attorneys: attorneys})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingType    donordata.AttorneysAct
		existingDetails string
		updatedType     donordata.AttorneysAct
		updatedDetails  string
		formType        string
		formDetails     string
		redirect        page.LpaPath
	}{
		"existing details not set": {
			existingType:    donordata.JointlyAndSeverally,
			existingDetails: "",
			updatedType:     donordata.JointlyForSomeSeverallyForOthers,
			updatedDetails:  "some details",
			formType:        donordata.JointlyForSomeSeverallyForOthers.String(),
			formDetails:     "some details",
			redirect:        page.Paths.BecauseYouHaveChosenJointlyForSomeSeverallyForOthers,
		},
		"existing details set": {
			existingType:    donordata.JointlyForSomeSeverallyForOthers,
			existingDetails: "some details",
			updatedType:     donordata.Jointly,
			updatedDetails:  "",
			formType:        donordata.Jointly.String(),
			formDetails:     "some details",
			redirect:        page.Paths.BecauseYouHaveChosenJointly,
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
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:             "lpa-id",
					Attorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", Address: testAddress}, {FirstNames: "b", Address: testAddress}}},
					AttorneyDecisions: donordata.AttorneyDecisions{Details: tc.updatedDetails, How: tc.updatedType},
					Tasks:             donordata.Tasks{ChooseAttorneys: actor.TaskCompleted},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:             "lpa-id",
				Attorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", Address: testAddress}, {FirstNames: "b", Address: testAddress}}},
				AttorneyDecisions: donordata.AttorneyDecisions{Details: tc.existingDetails, How: tc.existingType},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
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
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howShouldAttorneysMakeDecisionsData) bool {
			return assert.Equal(t, validation.With("decision-type", validation.SelectError{Label: "howAttorneysShouldMakeDecisions"}), data.Errors)
		})).
		Return(nil)

	err := HowShouldAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"decision-type": {donordata.JointlyAndSeverally.String()},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{})
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
				DecisionsType:    donordata.JointlyForSomeSeverallyForOthers,
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
		"missing decision detail when jointly for some severally for others": {
			form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:     donordata.JointlyForSomeSeverallyForOthers,
				detailsErrorLabel: "xyz",
			},
			errors: validation.With("mixed-details", validation.EnterError{Label: "xyz"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
