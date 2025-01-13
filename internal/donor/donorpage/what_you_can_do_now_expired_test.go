package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
	testcases := map[string]struct {
		BannerContent         string
		NewVoucherLabel       string
		ProveOwnIdentityLabel string
		CanHaveVoucher        bool
		WantVoucher           form.YesNo
		FailedVouchCount      int
	}{
		"no failed vouches": {
			BannerContent:         "yourConfirmedIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillReturnToOneLogin",
			CanHaveVoucher:        true,
		},
		"no failed vouches - has selected vouch option": {
			BannerContent:         "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			WantVoucher:           form.Yes,
		},
		"one failed vouch - has not selected voucher option": {
			BannerContent:         "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			WantVoucher:           form.YesNoUnknown,
			FailedVouchCount:      1,
		},
		"one failed vouch - has selected voucher option": {
			BannerContent:         "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			WantVoucher:           form.Yes,
			FailedVouchCount:      1,
		},
		"two failed vouches": {
			BannerContent:         "yourVouchedForIdentityHasExpiredSecondAttempt",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			WantVoucher:           form.Yes,
			FailedVouchCount:      2,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
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
					FailedVouchAttempts:   tc.FailedVouchCount,
					BannerContent:         tc.BannerContent,
					NewVoucherLabel:       tc.NewVoucherLabel,
					ProveOwnIdentityLabel: tc.ProveOwnIdentityLabel,
				}).
				Return(nil)

			err := WhatYouCanDoNowExpired(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				FailedVouchAttempts: tc.FailedVouchCount,
				WantVoucher:         tc.WantVoucher,
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
		donordata.ProveOwnIdentity: {
			expectedPath: donor.PathConfirmYourIdentity.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
				WantVoucher:      form.No,
			},
		},
		donordata.SelectNewVoucher: {
			expectedPath: donor.PathEnterVoucher.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
				WantVoucher:      form.Yes,
			},
		},
		donordata.WithdrawLPA: {
			expectedPath: donor.PathWithdrawThisLpa.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
				WantVoucher:      form.No,
			},
		},
		donordata.ApplyToCOP: {
			expectedPath: donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:                            "lpa-id",
				RegisteringWithCourtOfProtection: true,
				IdentityUserData:                 identity.UserData{Status: identity.StatusInsufficientEvidence},
				WantVoucher:                      form.No,
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
