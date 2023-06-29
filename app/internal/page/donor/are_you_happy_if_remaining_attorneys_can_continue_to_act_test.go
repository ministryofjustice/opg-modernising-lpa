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

func TestGetAreYouHappyIfRemainingAttorneysCanContinueToAct(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouHappyIfRemainingAttorneysCanContinueToActData{
			App:     testAppData,
			Options: actor.YesNoValues,
		}).
		Return(nil)

	err := AreYouHappyIfRemainingAttorneysCanContinueToAct(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouHappyIfRemainingAttorneysCanContinueToActWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := AreYouHappyIfRemainingAttorneysCanContinueToAct(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouHappyIfRemainingAttorneysCanContinueToAct(t *testing.T) {
	for _, happy := range []actor.YesNo{actor.Yes, actor.No} {
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
					AttorneyDecisions: actor.AttorneyDecisions{HappyIfRemainingCanContinueToAct: happy},
				}).
				Return(nil)

			err := AreYouHappyIfRemainingAttorneysCanContinueToAct(nil, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouHappyIfRemainingAttorneysCanContinueToActWhenStoreErrors(t *testing.T) {
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

	err := AreYouHappyIfRemainingAttorneysCanContinueToAct(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouHappyIfRemainingAttorneysCanContinueToActWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *areYouHappyIfRemainingAttorneysCanContinueToActData) bool {
			return assert.Equal(t, validation.With("happy", validation.SelectError{Label: "yesIfYouAreHappyIfRemainingAttorneysCanContinueToAct"}), data.Errors)
		})).
		Return(nil)

	err := AreYouHappyIfRemainingAttorneysCanContinueToAct(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
