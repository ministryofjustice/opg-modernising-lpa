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

func TestGetAreYouHappyIfOneAttorneyCantActNoneCan(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneAttorneyCantActNoneCanData{
			App: testAppData,
		}).
		Return(nil)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneAttorneyCantActNoneCanWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneAttorneyCantActNoneCanWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneAttorneyCantActNoneCanData{
			App: testAppData,
		}).
		Return(expectedError)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouHappyIfOneAttorneyCantActNoneCan(t *testing.T) {
	testcases := map[string]string{
		"yes": page.Paths.DoYouWantReplacementAttorneys,
		"no":  page.Paths.AreYouHappyIfRemainingAttorneysCanContinueToAct,
	}

	for happy, redirect := range testcases {
		t.Run(happy, func(t *testing.T) {
			form := url.Values{
				"happy": {happy},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Get", r.Context()).
				Return(&page.Lpa{}, nil)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					AttorneyDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: happy},
				}).
				Return(nil)

			err := AreYouHappyIfOneAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouHappyIfOneAttorneyCantActNoneCanWhenStoreErrors(t *testing.T) {
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

	err := AreYouHappyIfOneAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouHappyIfOneAttorneyCantActNoneCanWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneAttorneyCantActNoneCanData{
			App:    testAppData,
			Errors: validation.With("happy", validation.SelectError{Label: "yesIfYouAreHappyIfOneAttorneyCantActNoneCan"}),
		}).
		Return(nil)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
