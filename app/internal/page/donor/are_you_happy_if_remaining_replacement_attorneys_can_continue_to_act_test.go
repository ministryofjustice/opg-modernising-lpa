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

func TestGetAreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfRemainingReplacementAttorneysCanContinueToActData{
			App: testAppData,
		}).
		Return(nil)

	err := AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfRemainingReplacementAttorneysCanContinueToActWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfRemainingReplacementAttorneysCanContinueToActWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfRemainingReplacementAttorneysCanContinueToActData{
			App: testAppData,
		}).
		Return(expectedError)

	err := AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(t *testing.T) {
	testcases := map[string]struct {
		happy    string
		lpaType  string
		redirect string
	}{
		"yes pfa": {
			happy:    "yes",
			lpaType:  page.LpaTypePropertyFinance,
			redirect: page.Paths.WhenCanTheLpaBeUsed,
		},
		"yes hw": {
			happy:    "yes",
			lpaType:  page.LpaTypeHealthWelfare,
			redirect: page.Paths.LifeSustainingTreatment,
		},
		"no pfa": {
			happy:    "no",
			lpaType:  page.LpaTypePropertyFinance,
			redirect: page.Paths.WhenCanTheLpaBeUsed,
		},
		"no hw": {
			happy:    "no",
			lpaType:  page.LpaTypeHealthWelfare,
			redirect: page.Paths.LifeSustainingTreatment,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"happy": {tc.happy},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Type:  tc.lpaType,
					Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					Type:                         tc.lpaType,
					ReplacementAttorneyDecisions: actor.AttorneyDecisions{HappyIfRemainingCanContinueToAct: tc.happy},
					Tasks:                        page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
				}).
				Return(nil)

			err := AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouHappyIfRemainingReplacementAttorneysCanContinueToActWhenStoreErrors(t *testing.T) {
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

	err := AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouHappyIfRemainingReplacementAttorneysCanContinueToActWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfRemainingReplacementAttorneysCanContinueToActData{
			App:    testAppData,
			Errors: validation.With("happy", validation.SelectError{Label: "yesIfYouAreHappyIfRemainingReplacementAttorneysCanContinueToAct"}),
		}).
		Return(nil)

	err := AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
