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

func TestGetHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App:  testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneyDecisions: actor.AttorneyDecisions{Details: "some decisions", How: "jointly"}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App: testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "jointly",
				DecisionsDetails: "some decisions",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowShouldReplacementAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App: testAppData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
			},
		}).
		Return(expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysMakeDecisions(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneyDecisions: actor.AttorneyDecisions{Details: "", How: ""}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{ReplacementAttorneyDecisions: actor.AttorneyDecisions{Details: "", How: "jointly"}}).
		Return(nil)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsFromStore(t *testing.T) {
	testCases := map[string]struct {
		form      url.Values
		existing  actor.AttorneyDecisions
		attorneys actor.Attorneys
		updated   actor.AttorneyDecisions
		taskState actor.TaskState
		redirect  string
	}{
		"existing details not set": {
			form: url.Values{
				"decision-type": {"mixed"},
				"mixed-details": {"some details"},
			},
			existing:  actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			attorneys: actor.Attorneys{{FirstNames: "a", Email: "a"}},
			updated:   actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, Details: "some details"},
			taskState: actor.TaskCompleted,
			redirect:  page.Paths.TaskList,
		},
		"existing details set": {
			form: url.Values{
				"decision-type": {"jointly"},
				"mixed-details": {"some details"},
			},
			existing:  actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, Details: "some details"},
			attorneys: actor.Attorneys{{FirstNames: "a", Email: "a"}},
			updated:   actor.AttorneyDecisions{How: actor.Jointly},
			taskState: actor.TaskCompleted,
			redirect:  page.Paths.TaskList,
		},
		"requires happiness": {
			form: url.Values{
				"decision-type": {"jointly"},
				"mixed-details": {"some details"},
			},
			existing:  actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, Details: "some details"},
			attorneys: actor.Attorneys{{FirstNames: "a", Email: "a"}, {FirstNames: "b", Email: "b"}},
			updated:   actor.AttorneyDecisions{How: actor.Jointly},
			taskState: actor.TaskInProgress,
			redirect:  page.Paths.AreYouHappyIfOneReplacementAttorneyCantActNoneCan,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ReplacementAttorneys:         tc.attorneys,
					ReplacementAttorneyDecisions: tc.existing,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					ReplacementAttorneys:         tc.attorneys,
					ReplacementAttorneyDecisions: tc.updated,
					Tasks:                        page.Tasks{ChooseReplacementAttorneys: tc.taskState},
				}).
				Return(nil)

			template := newMockTemplate(t)

			err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {"some decisions"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowShouldReplacementAttorneysMakeDecisions(nil, lpaStore)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneyDecisions: actor.AttorneyDecisions{Details: "", How: ""}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howShouldReplacementAttorneysMakeDecisionsData{
			App:    testAppData,
			Errors: validation.With("decision-type", validation.SelectError{Label: "howReplacementAttorneysShouldMakeDecisions"}),
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    "",
				DecisionsDetails: "",
				errorLabel:       "howReplacementAttorneysShouldMakeDecisions",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowShouldReplacementAttorneysMakeDecisionsErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{ReplacementAttorneyDecisions: actor.AttorneyDecisions{Details: "", How: "jointly"}}).
		Return(expectedError)

	template := newMockTemplate(t)

	err := HowShouldReplacementAttorneysMakeDecisions(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
