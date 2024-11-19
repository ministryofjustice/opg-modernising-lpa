package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCompletingYourIdentityConfirmation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Donor:    lpadata.Donor{FirstNames: "A", LastName: "B"},
			SignedAt: testNow,
		}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &completingYourIdentityConfirmationData{
			App:      testAppData,
			Form:     &howWillYouConfirmYourIdentityForm{},
			Options:  howYouWillConfirmYourIdentityValues,
			Donor:    lpadata.Donor{FirstNames: "A", LastName: "B"},
			Deadline: testNow.AddDate(0, 6, 0),
		}).
		Return(nil)

	err := CompletingYourIdentityConfirmation(template.Execute, lpaStoreResolvingService)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCompletingYourIdentityConfirmationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := CompletingYourIdentityConfirmation(template.Execute, lpaStoreResolvingService)(testAppData, w, r, &voucherdata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostCompletingYourIdentityConfirmation(t *testing.T) {
	testCases := map[string]struct {
		how      howYouWillConfirmYourIdentity
		provided *voucherdata.Provided
		redirect voucher.Path
	}{
		"post office successful": {
			how:      howYouWillConfirmYourIdentityPostOfficeSuccessfully,
			provided: &voucherdata.Provided{LpaID: "lpa-id"},
			redirect: voucher.PathIdentityWithOneLogin,
		},
		"one login": {
			how:      howYouWillConfirmYourIdentityOneLogin,
			provided: &voucherdata.Provided{LpaID: "lpa-id"},
			redirect: voucher.PathIdentityWithOneLogin,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"how": {tc.how.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := CompletingYourIdentityConfirmation(nil, nil)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCompletingYourIdentityConfirmationWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *completingYourIdentityConfirmationData) bool {
			return assert.Equal(t, validation.With("how", validation.SelectError{Label: "howYouWouldLikeToContinue"}), data.Errors)
		})).
		Return(nil)

	err := CompletingYourIdentityConfirmation(template.Execute, lpaStoreResolvingService)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
