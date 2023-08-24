package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWouldLikeSecondSignatory(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &wouldLikeSecondSignatoryData{
			App:     testAppData,
			Options: form.YesNoValues,
		}).
		Return(nil)

	err := WouldLikeSecondSignatory(template.Execute, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWouldLikeSecondSignatoryWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &wouldLikeSecondSignatoryData{
			App:     testAppData,
			Options: form.YesNoValues,
		}).
		Return(expectedError)

	err := WouldLikeSecondSignatory(template.Execute, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWouldLikeSecondSignatory(t *testing.T) {
	testcases := map[form.YesNo]string{
		form.Yes: page.Paths.Attorney.Sign.Format("lpa-id") + "?second",
		form.No:  page.Paths.Attorney.WhatHappensNext.Format("lpa-id"),
	}

	for wouldLike, redirect := range testcases {
		t.Run(wouldLike.String(), func(t *testing.T) {
			f := url.Values{
				"yes-no": {wouldLike.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Put", r.Context(), &actor.AttorneyProvidedDetails{
					LpaID:                    "lpa-id",
					WouldLikeSecondSignatory: wouldLike,
				}).
				Return(nil)

			err := WouldLikeSecondSignatory(nil, attorneyStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostWouldLikeSecondSignatoryWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		"yes-no": {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := WouldLikeSecondSignatory(nil, attorneyStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostWouldLikeSecondSignatoryWhenValidationError(t *testing.T) {
	form := url.Values{
		"yes-no": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("yes-no", validation.SelectError{Label: "yesIfWouldLikeSecondSignatory"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *wouldLikeSecondSignatoryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := WouldLikeSecondSignatory(template.Execute, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
