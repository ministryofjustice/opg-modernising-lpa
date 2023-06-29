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

func TestGetAreYouHappyIfOneAttorneyCantActNoneCan(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfOneAttorneyCantActNoneCanData{
			App:     testAppData,
			Options: actor.YesNoValues,
		}).
		Return(nil)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfOneAttorneyCantActNoneCanWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouHappyIfOneAttorneyCantActNoneCan(t *testing.T) {
	testcases := map[actor.YesNo]page.LpaPath{
		actor.Yes: page.Paths.TaskList,
		actor.No:  page.Paths.AreYouHappyIfRemainingAttorneysCanContinueToAct,
	}

	for happy, redirect := range testcases {
		t.Run(happy.String(), func(t *testing.T) {
			form := url.Values{
				"happy": {happy.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                "lpa-id",
					AttorneyDecisions: actor.AttorneyDecisions{HappyIfOneCannotActNoneCan: happy},
				}).
				Return(nil)

			err := AreYouHappyIfOneAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouHappyIfOneAttorneyCantActNoneCanWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"happy": {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouHappyIfOneAttorneyCantActNoneCanWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *areYouHappyIfOneAttorneyCantActNoneCanData) bool {
			return assert.Equal(t, validation.With("happy", validation.SelectError{Label: "yesIfYouAreHappyIfOneAttorneyCantActNoneCan"}), data.Errors)
		})).
		Return(nil)

	err := AreYouHappyIfOneAttorneyCantActNoneCan(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
