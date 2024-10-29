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

	err := AreYouSureYouNoLongerNeedVoucher(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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

	err := AreYouSureYouNoLongerNeedVoucher(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouSureYouNoLongerNeedVoucher(t *testing.T) {
	testcases := map[donordata.NoVoucherDecision]struct {
		redirect donor.Path
		provided *donordata.Provided
	}{
		donordata.ProveOwnIdentity: {
			redirect: donor.PathTaskList,
			provided: &donordata.Provided{
				LpaID:   "lpa-id",
				Voucher: donordata.Voucher{FirstNames: "a", LastName: "b"},
			},
		},
		donordata.SelectNewVoucher: {
			redirect: donor.PathEnterVoucher,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
		},
		donordata.WithdrawLPA: {
			redirect: donor.PathWithdrawThisLpa,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
		},
		donordata.ApplyToCOP: {
			redirect: donor.PathWhatHappensNextRegisteringWithCourtOfProtection,
			provided: &donordata.Provided{
				LpaID:                            "lpa-id",
				Voucher:                          donordata.Voucher{FirstNames: "a", LastName: "b"},
				IdentityUserData:                 identity.UserData{Status: identity.StatusConfirmed},
				RegisteringWithCourtOfProtection: true,
			},
		},
	}

	for choice, tc := range testcases {
		t.Run(choice.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?choice="+choice.String(), nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				DeleteVoucher(r.Context(), tc.provided).
				Return(nil)

			err := AreYouSureYouNoLongerNeedVoucher(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.Yes,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b"},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathWeHaveInformedVoucherNoLongerNeeded.FormatQuery("lpa-id", url.Values{
				"choice":          {choice.String()},
				"voucherFullName": {"a b"},
				"next":            {tc.redirect.Format("lpa-id")},
			}), resp.Header.Get("Location"))
		})
	}
}

func TestPostAreYouSureYouNoLongerNeedVoucherWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?choice="+donordata.ProveOwnIdentity.String(), nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		DeleteVoucher(r.Context(), mock.Anything).
		Return(expectedError)

	err := AreYouSureYouNoLongerNeedVoucher(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}
