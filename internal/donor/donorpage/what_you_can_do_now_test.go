package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhatYouCanDoNow(t *testing.T) {
	testcases := map[int]struct {
		BannerContent   string
		NewVoucherLabel string
		CanHaveVoucher  bool
	}{
		0: {
			BannerContent:   "youHaveNotChosenAnyoneToVouchForYou",
			NewVoucherLabel: "iHaveSomeoneWhoCanVouch",
			CanHaveVoucher:  true,
		},
		1: {
			BannerContent:   "thePersonYouAskedToVouchHasBeenUnableToContinue",
			NewVoucherLabel: "iHaveSomeoneElseWhoCanVouch",
			CanHaveVoucher:  true,
		},
		2: {
			BannerContent: "thePersonYouAskedToVouchHasBeenUnableToContinueSecondAttempt",
		},
	}

	for failedVouchAttempts, tc := range testcases {
		t.Run(strconv.Itoa(failedVouchAttempts), func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &whatYouCanDoNowData{
					App: testAppData,
					Form: &whatYouCanDoNowForm{
						Options:        donordata.NoVoucherDecisionValues,
						CanHaveVoucher: tc.CanHaveVoucher,
					},
					FailedVouchAttempts: failedVouchAttempts,
					BannerContent:       tc.BannerContent,
					NewVoucherLabel:     tc.NewVoucherLabel,
				}).
				Return(nil)

			err := WhatYouCanDoNow(template.Execute, nil)(testAppData, w, r, &donordata.Provided{FailedVouchAttempts: failedVouchAttempts})

			assert.Nil(t, err)
		})
	}

}

func TestGetWhatYouCanDoNowWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatYouCanDoNow(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})

	assert.Error(t, err)
}

func TestPostWhatYouCanDoNow(t *testing.T) {
	testcases := map[donordata.NoVoucherDecision]struct {
		expectedPath  string
		expectedDonor *donordata.Provided
	}{
		donordata.ProveOwnID: {
			expectedPath: donor.PathTaskList.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{},
			},
		},
		donordata.SelectNewVoucher: {
			expectedPath: donor.PathEnterVoucher.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.Yes,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		donordata.WithdrawLPA: {
			expectedPath: donor.PathWithdrawThisLpa.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		donordata.ApplyToCOP: {
			expectedPath: donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:                            "lpa-id",
				RegisteringWithCourtOfProtection: true,
				IdentityUserData:                 identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
	}

	for noVoucherDecision, tc := range testcases {
		t.Run(noVoucherDecision.String(), func(t *testing.T) {
			f := url.Values{
				"do-next": {noVoucherDecision.String()},
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.expectedDonor).
				Return(nil)

			err := WhatYouCanDoNow(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedPath, resp.Header.Get("Location"))
		})
	}
}

func TestPostWhatYouCanDoNowWhenDonorStoreError(t *testing.T) {
	f := url.Values{
		"do-next": {donordata.ApplyToCOP.String()},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatYouCanDoNow(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
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

	err := WhatYouCanDoNow(template.Execute, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
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

	result := readWhatYouCanDoNowForm(r, &donordata.Provided{})

	assert.Equal(donordata.WithdrawLPA, result.DoNext)
	assert.Nil(result.Error)
}

func TestWhatYouCanDoNowFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whatYouCanDoNowForm
		errors validation.List
	}{
		"valid": {
			form: &whatYouCanDoNowForm{
				DoNext: donordata.WithdrawLPA,
			},
		},
		"invalid": {
			form: &whatYouCanDoNowForm{
				DoNext: donordata.NoVoucherDecision(99),
				Error:  expectedError,
			},
			errors: validation.
				With("do-next", validation.SelectError{Label: "whatYouWouldLikeToDo"}),
		},
		"not allowed another vouch": {
			form: &whatYouCanDoNowForm{
				DoNext:         donordata.SelectNewVoucher,
				CanHaveVoucher: false,
			},
			errors: validation.
				With("do-next", validation.CustomError{Label: "youCannotAskAnotherPersonToVouchForYou"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
