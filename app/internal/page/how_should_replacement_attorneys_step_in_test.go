package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldReplacementAttorneysStepIn(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App:  appData,
			Form: &howShouldReplacementAttorneysStepInForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			HowShouldReplacementAttorneysStepIn:        SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details",
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: SomeOtherWay,
				OtherDetails: "some details",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysStepInWhenStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysStepIn(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {SomeOtherWay},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			HowShouldReplacementAttorneysStepIn:        "",
			HowShouldReplacementAttorneysStepInDetails: "",
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			HowShouldReplacementAttorneysStepIn:        SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(nil)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.TaskList, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysStepInRedirects(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                           []Attorney
		ReplacementAttorneys                []Attorney
		HowAttorneysMakeDecisions           string
		HowShouldReplacementAttorneysStepIn string
		ExpectedRedirectUrl                 string
	}{
		"single attorney": {
			Attorneys: []Attorney{
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "doesnt matter",
			HowShouldReplacementAttorneysStepIn: "doesnt matter",
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + Paths.TaskList,
		},
		"multiple attorneys acting jointly and severally replacements step in when none left": {
			Attorneys: []Attorney{
				{ID: "123"},
				{ID: "123"},
			},
			ReplacementAttorneys: []Attorney{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "jointly-and-severally",
			HowShouldReplacementAttorneysStepIn: AllCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + Paths.HowShouldReplacementAttorneysMakeDecisions,
		},
		"multiple attorneys acting jointly and severally replacements step in when one loses capacity": {
			Attorneys: []Attorney{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "jointly-and-severally",
			HowShouldReplacementAttorneysStepIn: OneCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + Paths.TaskList,
		},
		"multiple attorneys acting jointly": {
			Attorneys: []Attorney{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "jointly-and-severally",
			HowShouldReplacementAttorneysStepIn: "doesnt matter",
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.HowShouldReplacementAttorneysStepIn},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					HowAttorneysMakeDecisions: tc.HowAttorneysMakeDecisions,
					Attorneys:                 tc.Attorneys,
					ReplacementAttorneys:      tc.ReplacementAttorneys,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
					Attorneys:                           tc.Attorneys,
					ReplacementAttorneys:                tc.ReplacementAttorneys,
					HowAttorneysMakeDecisions:           tc.HowAttorneysMakeDecisions,
					HowShouldReplacementAttorneysStepIn: tc.HowShouldReplacementAttorneysStepIn}).
				Return(nil)

			template := &mockTemplate{}

			err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirectUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, template, lpaStore)
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
			existingWhenStepIn:   AllCanNoLongerAct,
			existingOtherDetails: "",
			updatedWhenStepIn:    SomeOtherWay,
			updatedOtherDetails:  "some details",
			formWhenStepIn:       SomeOtherWay,
			formOtherDetails:     "some details",
		},
		"existing otherDetails set": {
			existingWhenStepIn:   SomeOtherWay,
			existingOtherDetails: "some details",
			updatedWhenStepIn:    OneCanNoLongerAct,
			updatedOtherDetails:  "",
			formWhenStepIn:       OneCanNoLongerAct,
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
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					HowShouldReplacementAttorneysStepIn:        tc.existingWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.existingOtherDetails,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
					HowShouldReplacementAttorneysStepIn:        tc.updatedWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.updatedOtherDetails}).
				Return(nil)

			template := &mockTemplate{}

			err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+Paths.TaskList, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, template, lpaStore)
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
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App: appData,
			Errors: map[string]string{
				"when-to-step-in": "selectWhenToStepIn",
			},
			Form: &howShouldReplacementAttorneysStepInForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysStepInWhenPutStoreError(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {SomeOtherWay},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			HowShouldReplacementAttorneysStepIn:        "",
			HowShouldReplacementAttorneysStepInDetails: "",
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			HowShouldReplacementAttorneysStepIn:        SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(expectedError)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestHowShouldReplacementAttorneysStepInFormValidate(t *testing.T) {
	testCases := map[string]struct {
		whenToStepIn   string
		otherDetails   string
		expectedErrors map[string]string
	}{
		"missing whenToStepIn": {
			whenToStepIn:   "",
			otherDetails:   "",
			expectedErrors: map[string]string{"when-to-step-in": "selectWhenToStepIn"},
		},
		"other missing otherDetail": {
			whenToStepIn:   SomeOtherWay,
			otherDetails:   "",
			expectedErrors: map[string]string{"other-details": "provideDetailsOfWhenToStepIn"},
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
