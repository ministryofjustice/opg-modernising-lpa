package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAreYouSureYouNoLongerNeedVoucher(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &areYouSureYouNoLongerNeedVoucherData{
			App:   testAppData,
			Donor: &donordata.Provided{},
		}).
		Return(nil)

	err := AreYouSureYouNoLongerNeedVoucher(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouSureYouNoLongerNeedVoucherWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := AreYouSureYouNoLongerNeedVoucher(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouSureYouNoLongerNeedVoucher(t *testing.T) {
	testcases := map[string]struct {
		decision donordata.NoVoucherDecision
		donor    *donordata.Provided
		redirect donor.Path
		provided *donordata.Provided
	}{
		"prove own identity": {
			decision: donordata.ProveOwnIdentity,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			redirect: donor.PathConfirmYourIdentity,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				WantVoucher:      form.No,
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
		},
		"select new voucher": {
			decision: donordata.SelectNewVoucher,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			redirect: donor.PathEnterVoucher,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
		},
		"delete": {
			decision: donordata.WithdrawLPA,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			redirect: donor.PathDeleteThisLpa,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				WantVoucher:      form.No,
			},
		},
		"withdraw": {
			decision: donordata.WithdrawLPA,
			donor: &donordata.Provided{
				LpaID:                            "lpa-id",
				LpaUID:                           "lpa-uid",
				Donor:                            donordata.Donor{FirstNames: "d", LastName: "e"},
				WantVoucher:                      form.Yes,
				Voucher:                          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData:                 identity.UserData{Status: identity.StatusConfirmed},
				WitnessedByCertificateProviderAt: testNow,
			},
			redirect: donor.PathWithdrawThisLpa,
			provided: &donordata.Provided{
				LpaID:                            "lpa-id",
				LpaUID:                           "lpa-uid",
				Donor:                            donordata.Donor{FirstNames: "d", LastName: "e"},
				Voucher:                          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData:                 identity.UserData{Status: identity.StatusConfirmed},
				WantVoucher:                      form.No,
				WitnessedByCertificateProviderAt: testNow,
			},
		},
		"apply to court of protection": {
			decision: donordata.ApplyToCOP,
			donor: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "d", LastName: "e"},
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			redirect: donor.PathWhatHappensNextRegisteringWithCourtOfProtection,
			provided: &donordata.Provided{
				LpaID:                            "lpa-id",
				LpaUID:                           "lpa-uid",
				Donor:                            donordata.Donor{FirstNames: "d", LastName: "e"},
				Voucher:                          donordata.Voucher{FirstNames: "a", LastName: "b", Email: "voucher@example.com"},
				IdentityUserData:                 identity.UserData{Status: identity.StatusConfirmed},
				RegisteringWithCourtOfProtection: true,
				WantVoucher:                      form.No,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?choice="+tc.decision.String(), nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				DeleteVoucher(r.Context(), tc.provided).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(r.Context(), notify.ToVoucher(tc.provided.Voucher), "lpa-uid", notify.VoucherInformedTheyAreNoLongerNeededToVouchEmail{
					DonorFullName:   "d e",
					VoucherFullName: "a b",
				}).
				Return(nil)

			err := AreYouSureYouNoLongerNeedVoucher(nil, donorStore, notifyClient)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathWeHaveInformedVoucherNoLongerNeeded.FormatQuery("lpa-id", url.Values{
				"choice":          {tc.decision.String()},
				"voucherFullName": {"a b"},
				"next":            {tc.redirect.Format("lpa-id")},
			}), resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouSureYouNoLongerNeedVoucherWhenNotifyErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?choice="+donordata.ProveOwnIdentity.String(), nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := AreYouSureYouNoLongerNeedVoucher(nil, nil, notifyClient)(testAppData, w, r, &donordata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostAreYouSureYouNoLongerNeedVoucherWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?choice="+donordata.ProveOwnIdentity.String(), nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		DeleteVoucher(r.Context(), mock.Anything).
		Return(expectedError)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := AreYouSureYouNoLongerNeedVoucher(nil, donorStore, notifyClient)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostAreYouSureYouNoLongerNeedVoucherWhenInvalidChoice(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?choice=what", nil)

	err := AreYouSureYouNoLongerNeedVoucher(nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	assert.Error(t, err)
}
