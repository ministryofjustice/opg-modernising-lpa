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

func TestGetHowShouldReplacementAttorneysStepIn(t *testing.T) {
	testcases := map[string]struct {
		attorneys         actor.Attorneys
		allowSomeOtherWay bool
	}{
		"single": {
			attorneys:         actor.Attorneys{{}},
			allowSomeOtherWay: true,
		},
		"multiple": {
			attorneys:         actor.Attorneys{{}, {}},
			allowSomeOtherWay: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &howShouldReplacementAttorneysStepInData{
					App:               testAppData,
					AllowSomeOtherWay: tc.allowSomeOtherWay,
					Form:              &howShouldReplacementAttorneysStepInForm{},
					Options:           page.ReplacementAttorneysStepInValues,
				}).
				Return(nil)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &page.Lpa{
				ReplacementAttorneys: tc.attorneys,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysStepInData{
			App: testAppData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: page.ReplacementAttorneysStepInAnotherWay,
				OtherDetails: "some details",
			},
			Options: page.ReplacementAttorneysStepInValues,
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &page.Lpa{
		HowShouldReplacementAttorneysStepIn:        page.ReplacementAttorneysStepInAnotherWay,
		HowShouldReplacementAttorneysStepInDetails: "some details",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepIn(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {page.ReplacementAttorneysStepInAnotherWay.String()},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                                  "lpa-id",
			HowShouldReplacementAttorneysStepIn: page.ReplacementAttorneysStepInAnotherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowShouldReplacementAttorneysStepInRedirects(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                            actor.Attorneys
		ReplacementAttorneys                 actor.Attorneys
		HowAttorneysMakeDecisions            actor.AttorneysAct
		HowReplacementAttorneysMakeDecisions actor.AttorneysAct
		HowShouldReplacementAttorneysStepIn  page.ReplacementAttorneysStepIn
		ExpectedRedirectUrl                  page.LpaPath
		TaskState                            actor.TaskState
	}{
		"multiple attorneys acting jointly and severally replacements step in when none left": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			ReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           actor.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: page.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			TaskState:                           actor.TaskInProgress,
		},
		"multiple attorneys acting jointly": {
			ReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:            actor.Jointly,
			HowShouldReplacementAttorneysStepIn:  page.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			HowReplacementAttorneysMakeDecisions: actor.Jointly,
			ExpectedRedirectUrl:                  page.Paths.TaskList,
			TaskState:                            actor.TaskInProgress,
		},
		"multiple attorneys acting jointly and severally replacements step in when one loses capacity": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           actor.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: page.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			ExpectedRedirectUrl:                 page.Paths.TaskList,
			TaskState:                           actor.TaskNotStarted,
		},
		"multiple attorneys acting jointly and severally": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			ReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           actor.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: page.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
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
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                                  "lpa-id",
					Attorneys:                           tc.Attorneys,
					AttorneyDecisions:                   actor.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
					ReplacementAttorneys:                tc.ReplacementAttorneys,
					ReplacementAttorneyDecisions:        actor.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
					HowShouldReplacementAttorneysStepIn: tc.HowShouldReplacementAttorneysStepIn,
					Tasks:                               page.Tasks{ChooseReplacementAttorneys: tc.TaskState},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{
				ID:                           "lpa-id",
				Attorneys:                    tc.Attorneys,
				AttorneyDecisions:            actor.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
				ReplacementAttorneys:         tc.ReplacementAttorneys,
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
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
		existingWhenStepIn   page.ReplacementAttorneysStepIn
		existingOtherDetails string
		updatedWhenStepIn    page.ReplacementAttorneysStepIn
		updatedOtherDetails  string
		formWhenStepIn       page.ReplacementAttorneysStepIn
		formOtherDetails     string
	}{
		"existing otherDetails not set": {
			existingWhenStepIn:   page.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			existingOtherDetails: "",
			updatedWhenStepIn:    page.ReplacementAttorneysStepInAnotherWay,
			updatedOtherDetails:  "some details",
		},
		"existing otherDetails set": {
			existingWhenStepIn:   page.ReplacementAttorneysStepInAnotherWay,
			existingOtherDetails: "some details",
			updatedWhenStepIn:    page.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
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
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                                  "lpa-id",
					HowShouldReplacementAttorneysStepIn: tc.updatedWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.updatedOtherDetails}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{
				ID:                                  "lpa-id",
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
	template.
		On("Execute", w, mock.MatchedBy(func(data *howShouldReplacementAttorneysStepInData) bool {
			return assert.Equal(t, validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}), data.Errors)
		})).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepInWhenPutStoreError(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {page.ReplacementAttorneysStepInAnotherWay.String()},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.ReplacementAttorneysStepInAnotherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, donorStore)(testAppData, w, r, &page.Lpa{})
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
				WhenToStepIn: page.ReplacementAttorneysStepInAnotherWay,
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
