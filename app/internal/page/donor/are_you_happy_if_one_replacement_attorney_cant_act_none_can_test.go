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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App: testAppData,
		}).
		Return(nil)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App: testAppData,
		}).
		Return(expectedError)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouHappyIfOneReplacementAttorneyCantActNoneCan(t *testing.T) {
	testcases := map[string]struct {
		happy    string
		lpaType  string
		lpa      *page.Lpa
		redirect string
	}{
		"yes hw": {
			happy:   "yes",
			lpaType: page.LpaTypeHealthWelfare,
			lpa: &page.Lpa{
				Type:                         page.LpaTypeHealthWelfare,
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: "yes"},
				Tasks:                        page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
			},
			redirect: page.Paths.LifeSustainingTreatment,
		},
		"yes pfa": {
			happy:   "yes",
			lpaType: page.LpaTypePropertyFinance,
			lpa: &page.Lpa{
				Type:                         page.LpaTypePropertyFinance,
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: "yes"},
				Tasks:                        page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
			},
			redirect: page.Paths.WhenCanTheLpaBeUsed,
		},
		"no": {
			happy: "no",
			lpa: &page.Lpa{
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: "no"},
				Tasks:                        page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
			},
			redirect: page.Paths.AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct,
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

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Type:  tc.lpaType,
					Tasks: page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
				}, nil)
			donorStore.
				On("Put", r.Context(), tc.lpa).
				Return(nil)

			err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r)
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouHappyIfOneReplacementAttorneyCantActNoneCanWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App:    testAppData,
			Errors: validation.With("happy", validation.SelectError{Label: "yesIfYouAreHappyIfOneReplacementAttorneyCantActNoneCan"}),
		}).
		Return(nil)

	err := AreYouHappyIfOneReplacementAttorneyCantActNoneCan(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
