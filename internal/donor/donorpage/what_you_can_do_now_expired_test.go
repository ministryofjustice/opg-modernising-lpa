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
		Donor                 *donordata.Provided
	}{
		"no failed vouches": {
			BannerContent:         "yourConfirmedIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillReturnToOneLogin",
			CanHaveVoucher:        true,
			Donor:                 &donordata.Provided{},
		},
		"no failed vouches - has selected vouch option": {
			BannerContent:         "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			Donor:                 &donordata.Provided{WantVoucher: form.Yes},
		},
		"one failed vouch - has not selected voucher option": {
			BannerContent:         "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			Donor:                 &donordata.Provided{VouchAttempts: 1, WantVoucher: form.YesNoUnknown},
		},
		"one failed vouch - has selected voucher option": {
			BannerContent:         "yourVouchedForIdentityHasExpired",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			Donor:                 &donordata.Provided{VouchAttempts: 1, WantVoucher: form.Yes},
		},
		"two failed vouches": {
			BannerContent:         "yourVouchedForIdentityHasExpiredSecondAttempt",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			Donor:                 &donordata.Provided{VouchAttempts: 2, WantVoucher: form.Yes},
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
					Donor:                 tc.Donor,
					BannerContent:         tc.BannerContent,
					NewVoucherLabel:       tc.NewVoucherLabel,
					ProveOwnIdentityLabel: tc.ProveOwnIdentityLabel,
				}).
				Return(nil)

			err := WhatYouCanDoNowExpired(template.Execute, nil)(testAppData, w, r, tc.Donor)

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
	testcases := map[string]struct {
		decision      donordata.NoVoucherDecision
		donor         *donordata.Provided
		expectedPath  string
		expectedDonor *donordata.Provided
	}{
		"prove own identity": {
			decision: donordata.ProveOwnIdentity,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
			expectedPath: donor.PathConfirmYourIdentity.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.No,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		"select new voucher": {
			decision: donordata.SelectNewVoucher,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
			expectedPath: donor.PathEnterVoucher.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.Yes,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		"delete": {
			decision: donordata.WithdrawLPA,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
			expectedPath: donor.PathDeleteThisLpa.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.No,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		"withdraw": {
			decision: donordata.WithdrawLPA,
			donor: &donordata.Provided{
				LpaID:                            "lpa-id",
				IdentityUserData:                 identity.UserData{Status: identity.StatusInsufficientEvidence},
				WitnessedByCertificateProviderAt: testNow,
			},
			expectedPath: donor.PathWithdrawThisLpa.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:                            "lpa-id",
				WantVoucher:                      form.No,
				IdentityUserData:                 identity.UserData{Status: identity.StatusInsufficientEvidence},
				WitnessedByCertificateProviderAt: testNow,
			},
		},
		"apply to court of protection": {
			decision: donordata.ApplyToCOP,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
			expectedPath: donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:                            "lpa-id",
				WantVoucher:                      form.No,
				RegisteringWithCourtOfProtection: true,
				IdentityUserData:                 identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"do-next": {tc.decision.String()},
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.expectedDonor).
				Return(nil)

			err := WhatYouCanDoNowExpired(nil, donorStore)(testAppData, w, r, tc.donor)
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
