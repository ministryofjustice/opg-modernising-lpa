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
	"github.com/stretchr/testify/mock"
)

func TestGetAreYouHappyIfOneReplacementAttorneyCantActNoneCan(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App: testAppData,
		}).
		Return(nil)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App: testAppData,
		}).
		Return(expectedError)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouHappyIfOneReplacementAttorneyCantActNoneCan(t *testing.T) {
	testcases := map[string]struct {
		lpa      *page.Lpa
		redirect string
	}{
		"yes": {
			lpa: &page.Lpa{
				HowReplacementAttorneysMakeDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: "yes"},
				Tasks:                                page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
			},
			redirect: page.Paths.WhenCanTheLpaBeUsed,
		},
		"no": {
			lpa: &page.Lpa{
				HowReplacementAttorneysMakeDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: "no"},
				Tasks:                                page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
			},
			redirect: page.Paths.AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct,
		},
	}

	for happy, tc := range testcases {
		t.Run(happy, func(t *testing.T) {
			form := url.Values{
				"happy": {happy},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted}}, nil)
			lpaStore.
				On("Put", r.Context(), tc.lpa).
				Return(nil)

			err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"happy": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App:    testAppData,
			Errors: validation.With("happy", validation.SelectError{Label: "yesIfYouAreHappyIfOneReplacementAttorneyCantActNoneCan"}),
		}).
		Return(nil)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
