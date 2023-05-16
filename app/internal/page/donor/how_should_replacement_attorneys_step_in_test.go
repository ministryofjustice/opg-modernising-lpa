package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
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

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ReplacementAttorneys: tc.attorneys,
				}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &howShouldReplacementAttorneysStepInData{
					App:               testAppData,
					AllowSomeOtherWay: tc.allowSomeOtherWay,
					Form:              &howShouldReplacementAttorneysStepInForm{},
				}).
				Return(nil)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details",
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysStepInData{
			App: testAppData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: page.SomeOtherWay,
				OtherDetails: "some details",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysStepInWhenStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepIn(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {page.SomeOtherWay},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			HowShouldReplacementAttorneysStepIn:        "",
			HowShouldReplacementAttorneysStepInDetails: "",
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostHowShouldReplacementAttorneysStepInRedirects(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                            actor.Attorneys
		ReplacementAttorneys                 actor.Attorneys
		HowAttorneysMakeDecisions            string
		HowReplacementAttorneysMakeDecisions string
		HowShouldReplacementAttorneysStepIn  string
		ExpectedRedirectUrl                  string
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
			HowShouldReplacementAttorneysStepIn: page.AllCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + page.Paths.HowShouldReplacementAttorneysMakeDecisions,
			TaskState:                           actor.TaskInProgress,
		},
		"multiple attorneys acting jointly": {
			ReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:            actor.Jointly,
			HowShouldReplacementAttorneysStepIn:  page.OneCanNoLongerAct,
			HowReplacementAttorneysMakeDecisions: actor.Jointly,
			ExpectedRedirectUrl:                  "/lpa/lpa-id" + page.Paths.AreYouHappyIfOneReplacementAttorneyCantActNoneCan,
			TaskState:                            actor.TaskInProgress,
		},
		"multiple attorneys acting jointly and severally replacements step in when one loses capacity": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           actor.JointlyAndSeverally,
			HowShouldReplacementAttorneysStepIn: page.OneCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + page.Paths.TaskList,
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
			HowShouldReplacementAttorneysStepIn: page.OneCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + page.Paths.TaskList,
			TaskState:                           actor.TaskInProgress,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.HowShouldReplacementAttorneysStepIn},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Attorneys:                    tc.Attorneys,
					AttorneyDecisions:            actor.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
					ReplacementAttorneys:         tc.ReplacementAttorneys,
					ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					Attorneys:                           tc.Attorneys,
					AttorneyDecisions:                   actor.AttorneyDecisions{How: tc.HowAttorneysMakeDecisions},
					ReplacementAttorneys:                tc.ReplacementAttorneys,
					ReplacementAttorneyDecisions:        actor.AttorneyDecisions{How: tc.HowReplacementAttorneysMakeDecisions},
					HowShouldReplacementAttorneysStepIn: tc.HowShouldReplacementAttorneysStepIn,
					Tasks:                               page.Tasks{ChooseReplacementAttorneys: tc.TaskState},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirectUrl, resp.Header.Get("Location"))

		})
	}
}

func TestPostHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingWhenStepIn   string
		existingOtherDetails string
		updatedWhenStepIn    string
		updatedOtherDetails  string
		formWhenStepIn       string
		formOtherDetails     string
	}{
		"existing otherDetails not set": {
			existingWhenStepIn:   page.AllCanNoLongerAct,
			existingOtherDetails: "",
			updatedWhenStepIn:    page.SomeOtherWay,
			updatedOtherDetails:  "some details",
			formWhenStepIn:       page.SomeOtherWay,
			formOtherDetails:     "some details",
		},
		"existing otherDetails set": {
			existingWhenStepIn:   page.SomeOtherWay,
			existingOtherDetails: "some details",
			updatedWhenStepIn:    page.OneCanNoLongerAct,
			updatedOtherDetails:  "",
			formWhenStepIn:       page.OneCanNoLongerAct,
			formOtherDetails:     "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.formWhenStepIn},
				"other-details":   {tc.formOtherDetails},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					HowShouldReplacementAttorneysStepIn:        tc.existingWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.existingOtherDetails,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					HowShouldReplacementAttorneysStepIn:        tc.updatedWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.updatedOtherDetails}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))

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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysStepInData{
			App:    testAppData,
			Errors: validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}),
			Form:   &howShouldReplacementAttorneysStepInForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysStepInWhenPutStoreError(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {page.SomeOtherWay},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			HowShouldReplacementAttorneysStepIn:        "",
			HowShouldReplacementAttorneysStepInDetails: "",
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysStepIn(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHowShouldReplacementAttorneysStepInFormValidate(t *testing.T) {
	testCases := map[string]struct {
		whenToStepIn   string
		otherDetails   string
		expectedErrors validation.List
	}{
		"missing whenToStepIn": {
			whenToStepIn:   "",
			otherDetails:   "",
			expectedErrors: validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}),
		},
		"other missing otherDetail": {
			whenToStepIn:   page.SomeOtherWay,
			otherDetails:   "",
			expectedErrors: validation.With("other-details", validation.EnterError{Label: "detailsOfWhenToStepIn"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: tc.whenToStepIn,
				OtherDetails: tc.otherDetails,
			}

			assert.Equal(t, tc.expectedErrors, form.Validate())
		})
	}
}
