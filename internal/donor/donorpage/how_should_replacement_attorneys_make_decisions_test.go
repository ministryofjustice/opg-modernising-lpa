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

func TestGetHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howShouldReplacementAttorneysMakeDecisionsData{
			App:     testAppData,
			Form:    &howShouldAttorneysMakeDecisionsForm{},
			Options: donordata.AttorneysActValues,
			Donor:   &donordata.DonorProvidedDetails{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howShouldReplacementAttorneysMakeDecisionsData{
			App: testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    donordata.Jointly,
				DecisionsDetails: "some decisions",
			},
			Options: donordata.AttorneysActValues,
			Donor:   &donordata.DonorProvidedDetails{ReplacementAttorneyDecisions: donordata.AttorneyDecisions{Details: "some decisions", How: donordata.Jointly}},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneyDecisions: donordata.AttorneyDecisions{Details: "some decisions", How: donordata.Jointly}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	form := url.Values{
		"decision-type": {donordata.Jointly.String()},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{LpaID: "lpa-id", ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: donordata.Jointly}}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
	testCases := map[string]struct {
		form      url.Values
		existing  donordata.AttorneyDecisions
		attorneys donordata.Attorneys
		updated   donordata.AttorneyDecisions
		taskState actor.TaskState
		redirect  page.LpaPath
	}{
		"existing details not set": {
			form: url.Values{
				"decision-type": {donordata.JointlyForSomeSeverallyForOthers.String()},
				"mixed-details": {"some details"},
			},
			existing:  donordata.AttorneyDecisions{How: donordata.JointlyAndSeverally},
			attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", Address: testAddress, Email: "a"}}},
			updated:   donordata.AttorneyDecisions{How: donordata.JointlyForSomeSeverallyForOthers, Details: "some details"},
			taskState: actor.TaskCompleted,
			redirect:  page.Paths.TaskList,
		},
		"existing details set": {
			form: url.Values{
				"decision-type": {donordata.Jointly.String()},
				"mixed-details": {"some details"},
			},
			existing:  donordata.AttorneyDecisions{How: donordata.JointlyForSomeSeverallyForOthers, Details: "some details"},
			attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", Address: testAddress, Email: "a"}}},
			updated:   donordata.AttorneyDecisions{How: donordata.Jointly},
			taskState: actor.TaskCompleted,
			redirect:  page.Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID:                        "lpa-id",
					ReplacementAttorneys:         tc.attorneys,
					ReplacementAttorneyDecisions: tc.updated,
					Tasks:                        donordata.DonorTasks{ChooseReplacementAttorneys: tc.taskState},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         tc.attorneys,
				ReplacementAttorneyDecisions: tc.existing,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"decision-type": {donordata.Jointly.String()},
		"mixed-details": {"some decisions"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"decision-type": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howShouldReplacementAttorneysMakeDecisionsData) bool {
			return assert.Equal(t, validation.With("decision-type", validation.SelectError{Label: "howReplacementAttorneysShouldMakeDecisions"}), data.Errors)
		})).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"decision-type": {donordata.Jointly.String()},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{ReplacementAttorneyDecisions: donordata.AttorneyDecisions{Details: "", How: donordata.Jointly}}).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
