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

func TestGetWhatYouCanDoNow(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whatYouCanDoNowData{
			App: testAppData,
			Form: &whatYouCanDoNowForm{
				Options: actor.NoVoucherDecisionValues,
			},
		}).
		Return(nil)

	err := WhatYouCanDoNow(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Nil(t, err)
}

func TestGetWhatYouCanDoNowWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatYouCanDoNow(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Error(t, err)
}

func TestPostWhatYouCanDoNow(t *testing.T) {
	testcases := map[actor.NoVoucherDecision]string{
		actor.ProveOwnID:       page.Paths.TaskList.Format("lpa-id"),
		actor.SelectNewVoucher: page.Paths.EnterVoucher.Format("lpa-id"),
		actor.WithdrawLPA:      page.Paths.WithdrawThisLpa.Format("lpa-id"),
		actor.ApplyToCOP:       page.Paths.TaskList.Format("lpa-id"),
	}

	for noVoucherDecision, path := range testcases {
		t.Run(noVoucherDecision.String(), func(t *testing.T) {
			f := url.Values{
				"do-next": {noVoucherDecision.String()},
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", NoVoucherDecision: noVoucherDecision}).
				Return(nil)

			err := WhatYouCanDoNow(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, path, resp.Header.Get("Location"))
		})
	}
}

func TestPostWhatYouCanDoNowWhenDonorStoreError(t *testing.T) {
	f := url.Values{
		"do-next": {actor.ApplyToCOP.String()},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", NoVoucherDecision: actor.ApplyToCOP}).
		Return(expectedError)

	err := WhatYouCanDoNow(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhatYouCanDoNowWhenValidationErrors(t *testing.T) {
	f := url.Values{
		"do-next": {"not a valid value"},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whatYouCanDoNowData) bool {
			return assert.Equal(t, validation.With("do-next", validation.SelectError{Label: "whatYouWouldLikeToDo"}), data.Errors)
		})).
		Return(nil)

	err := WhatYouCanDoNow(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhatYouCanDoNowForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"do-next": {"  withdraw-lpa  "},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhatYouCanDoNowForm(r)

	assert.Equal(actor.WithdrawLPA, result.DoNext)
	assert.Nil(result.Error)
}

func TestWhatYouCanDoNowFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whatYouCanDoNowForm
		errors validation.List
	}{
		"valid": {
			form: &whatYouCanDoNowForm{
				DoNext: actor.WithdrawLPA,
			},
		},
		"invalid": {
			form: &whatYouCanDoNowForm{
				DoNext: actor.NoVoucherDecision(99),
				Error:  expectedError,
			},
			errors: validation.
				With("do-next", validation.SelectError{Label: "whatYouWouldLikeToDo"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
