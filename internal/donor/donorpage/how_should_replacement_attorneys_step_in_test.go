package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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
			Options: lpadata.ReplacementAttorneysStepInValues,
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
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
				WhenToStepIn: lpadata.ReplacementAttorneysStepInAnotherWay,
				OtherDetails: "some details",
			},
			Options: lpadata.ReplacementAttorneysStepInValues,
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		HowShouldReplacementAttorneysStepIn:        lpadata.ReplacementAttorneysStepInAnotherWay,
		HowShouldReplacementAttorneysStepInDetails: "some details",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepIn(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {lpadata.ReplacementAttorneysStepInAnotherWay.String()},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:                               "lpa-id",
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInAnotherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowShouldReplacementAttorneysStepInRedirects(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                            donordata.Attorneys
		ReplacementAttorneys                 donordata.Attorneys
		HowAttorneysMakeDecisions            lpadata.AttorneysAct
		HowReplacementAttorneysMakeDecisions lpadata.AttorneysAct
		HowShouldReplacementAttorneysStepIn  lpadata.ReplacementAttorneysStepIn
		ExpectedRedirectUrl                  donor.Path
		TaskState                            task.State
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
			HowAttorneysMakeDecisions:           lpadata.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			TaskState:                           task.StateInProgress,
		},
		"multiple attorneys acting jointly": {
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			HowAttorneysMakeDecisions:            lpadata.Jointly,
			HowShouldReplacementAttorneysStepIn:  lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			HowReplacementAttorneysMakeDecisions: lpadata.Jointly,
			ExpectedRedirectUrl:                  page.Paths.TaskList,
			TaskState:                            task.StateInProgress,
		},
		"multiple attorneys acting jointly and severally replacements step in when one loses capacity": {
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
				{UID: actoruid.New()},
				{UID: actoruid.New()},
			}},
			HowAttorneysMakeDecisions:           lpadata.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.TaskList,
			TaskState:                           task.StateNotStarted,
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
			HowAttorneysMakeDecisions:           lpadata.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.TaskList,
			TaskState:                           task.StateInProgress,
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
				Put(r.Context(), &donordata.Provided{
					LpaID:                               "lpa-id",
					Attorneys:                           tc.Attorneys,
					AttorneyDecisions:                   donordata.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
					ReplacementAttorneys:                tc.ReplacementAttorneys,
					ReplacementAttorneyDecisions:        donordata.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
					HowShouldReplacementAttorneysStepIn: tc.HowShouldReplacementAttorneysStepIn,
					Tasks:                               donordata.Tasks{ChooseReplacementAttorneys: tc.TaskState},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{
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
		existingWhenStepIn   lpadata.ReplacementAttorneysStepIn
		existingOtherDetails string
		updatedWhenStepIn    lpadata.ReplacementAttorneysStepIn
		updatedOtherDetails  string
		formWhenStepIn       lpadata.ReplacementAttorneysStepIn
		formOtherDetails     string
	}{
		"existing otherDetails not set": {
			existingWhenStepIn:   lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			existingOtherDetails: "",
			updatedWhenStepIn:    lpadata.ReplacementAttorneysStepInAnotherWay,
			updatedOtherDetails:  "some details",
		},
		"existing otherDetails set": {
			existingWhenStepIn:   lpadata.ReplacementAttorneysStepInAnotherWay,
			existingOtherDetails: "some details",
			updatedWhenStepIn:    lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
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
				Put(r.Context(), &donordata.Provided{
					LpaID:                               "lpa-id",
					HowShouldReplacementAttorneysStepIn: tc.updatedWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.updatedOtherDetails}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{
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

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepInWhenPutStoreError(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {lpadata.ReplacementAttorneysStepInAnotherWay.String()},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			HowShouldReplacementAttorneysStepIn:        lpadata.ReplacementAttorneysStepInAnotherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &donordata.Provided{})
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
				WhenToStepIn: lpadata.ReplacementAttorneysStepInAnotherWay,
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
