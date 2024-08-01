package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldReplacementAttorneysStepIn(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howShouldReplacementAttorneysStepInData{
			App:     testAppData,
			Form:    &howShouldReplacementAttorneysStepInForm{},
			Options: donordata.ReplacementAttorneysStepInValues,
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howShouldReplacementAttorneysStepInData{
			App: testAppData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: donordata.ReplacementAttorneysStepInAnotherWay,
				OtherDetails: "some details",
			},
			Options: donordata.ReplacementAttorneysStepInValues,
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		HowShouldReplacementAttorneysStepIn:        donordata.ReplacementAttorneysStepInAnotherWay,
		HowShouldReplacementAttorneysStepInDetails: "some details",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepIn(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {donordata.ReplacementAttorneysStepInAnotherWay.String()},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID:                               "lpa-id",
			HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInAnotherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowShouldReplacementAttorneysStepInRedirects(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                            donordata.Attorneys
		ReplacementAttorneys                 donordata.Attorneys
		HowAttorneysMakeDecisions            donordata.AttorneysAct
		HowReplacementAttorneysMakeDecisions donordata.AttorneysAct
		HowShouldReplacementAttorneysStepIn  donordata.ReplacementAttorneysStepIn
		ExpectedRedirectUrl                  page.LpaPath
		TaskState                            actor.TaskState
	}{
		"multiple attorneys acting jointly and severally replacements step in when none left": {
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			HowAttorneysMakeDecisions:           donordata.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			TaskState:                           actor.TaskInProgress,
		},
		"multiple attorneys acting jointly": {
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			HowAttorneysMakeDecisions:            donordata.Jointly,
			HowShouldReplacementAttorneysStepIn:  donordata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			HowReplacementAttorneysMakeDecisions: donordata.Jointly,
			ExpectedRedirectUrl:                  page.Paths.TaskList,
			TaskState:                            actor.TaskInProgress,
		},
		"multiple attorneys acting jointly and severally replacements step in when one loses capacity": {
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			HowAttorneysMakeDecisions:           donordata.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.TaskList,
			TaskState:                           actor.TaskNotStarted,
		},
		"multiple attorneys acting jointly and severally": {
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			HowAttorneysMakeDecisions:           donordata.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.TaskList,
			TaskState:                           actor.TaskInProgress,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.HowShouldReplacementAttorneysStepIn.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID:                               "lpa-id",
					Attorneys:                           tc.Attorneys,
					AttorneyDecisions:                   donordata.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
					ReplacementAttorneys:                tc.ReplacementAttorneys,
					ReplacementAttorneyDecisions:        donordata.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
					HowShouldReplacementAttorneysStepIn: tc.HowShouldReplacementAttorneysStepIn,
					Tasks:                               donordata.DonorTasks{ChooseReplacementAttorneys: tc.TaskState},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:                        "lpa-id",
				Attorneys:                    tc.Attorneys,
				AttorneyDecisions:            donordata.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
				ReplacementAttorneys:         tc.ReplacementAttorneys,
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirectUrl.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingWhenStepIn   donordata.ReplacementAttorneysStepIn
		existingOtherDetails string
		updatedWhenStepIn    donordata.ReplacementAttorneysStepIn
		updatedOtherDetails  string
		formWhenStepIn       donordata.ReplacementAttorneysStepIn
		formOtherDetails     string
	}{
		"existing otherDetails not set": {
			existingWhenStepIn:   donordata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			existingOtherDetails: "",
			updatedWhenStepIn:    donordata.ReplacementAttorneysStepInAnotherWay,
			updatedOtherDetails:  "some details",
		},
		"existing otherDetails set": {
			existingWhenStepIn:   donordata.ReplacementAttorneysStepInAnotherWay,
			existingOtherDetails: "some details",
			updatedWhenStepIn:    donordata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			updatedOtherDetails:  "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.updatedWhenStepIn.String()},
				"other-details":   {tc.updatedOtherDetails},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID:                               "lpa-id",
					HowShouldReplacementAttorneysStepIn: tc.updatedWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.updatedOtherDetails}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:                               "lpa-id",
				HowShouldReplacementAttorneysStepIn: tc.existingWhenStepIn,
				HowShouldReplacementAttorneysStepInDetails: tc.existingOtherDetails,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowShouldReplacementAttorneysStepInFormValidation(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {""},
		"other-details":   {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howShouldReplacementAttorneysStepInData) bool {
			return assert.Equal(t, validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}), data.Errors)
		})).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepInWhenPutStoreError(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {donordata.ReplacementAttorneysStepInAnotherWay.String()},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			HowShouldReplacementAttorneysStepIn:        donordata.ReplacementAttorneysStepInAnotherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHowShouldReplacementAttorneysStepInFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form           *howShouldReplacementAttorneysStepInForm
		expectedErrors validation.List
	}{
		"valid": {
			form: &howShouldReplacementAttorneysStepInForm{},
		},
		"invalid": {
			form: &howShouldReplacementAttorneysStepInForm{
				Error: expectedError,
			},
			expectedErrors: validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}),
		},
		"missing other details": {
			form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: donordata.ReplacementAttorneysStepInAnotherWay,
			},
			expectedErrors: validation.With("other-details", validation.EnterError{Label: "detailsOfWhenToStepIn"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedErrors, tc.form.Validate())
		})
	}
}
