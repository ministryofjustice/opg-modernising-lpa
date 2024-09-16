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

func TestGetWhatYouCanDoNowExpired(t *testing.T) {
	testcases := map[int]struct {
		BannerContent      string
		NewVoucherLabel    string
		ProveOwnIDLabel    string
		CanHaveVoucher     bool
		VouchedForIdentity bool
	}{
		0: {
			BannerContent:   "yourConfirmedIdentityHasExpired",
			NewVoucherLabel: "iHaveSomeoneWhoCanVouch",
			ProveOwnIDLabel: "iWillReturnToOneLogin",
			CanHaveVoucher:  true,
		},
		1: {
			BannerContent:      "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:    "iHaveSomeoneWhoCanVouch",
			ProveOwnIDLabel:    "iWillGetOrFindID",
			CanHaveVoucher:     true,
			VouchedForIdentity: true,
		},
		2: {
			BannerContent:      "yourVouchedForIdentityHasExpiredSecondAttempt",
			NewVoucherLabel:    "iHaveSomeoneWhoCanVouch",
			ProveOwnIDLabel:    "iWillGetOrFindID",
			VouchedForIdentity: true,
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
					ProveOwnIDLabel:     tc.ProveOwnIDLabel,
				}).
				Return(nil)

			err := WhatYouCanDoNowExpired(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				FailedVouchAttempts: failedVouchAttempts,
				IdentityUserData:    identity.UserData{VouchedFor: tc.VouchedForIdentity},
			})

			assert.Nil(t, err)
		})
	}

}

func TestGetWhatYouCanDoNowExpiredWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatYouCanDoNowExpired(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})

	assert.Error(t, err)
}

func TestPostWhatYouCanDoNowExpired(t *testing.T) {
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

			err := WhatYouCanDoNowExpired(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedPath, resp.Header.Get("Location"))
		})
	}
}

func TestPostWhatYouCanDoNowExpiredWhenDonorStoreError(t *testing.T) {
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

	err := WhatYouCanDoNowExpired(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhatYouCanDoNowExpiredWhenValidationErrors(t *testing.T) {
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

	err := WhatYouCanDoNowExpired(template.Execute, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
