package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourDetails(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	donor := &donordata.Provided{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, yourDetailsData{
			App:   testAppData,
			Donor: donor,
		}).
		Return(nil)

	err := YourDetails(template.Execute, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDetailsWhenTemplateError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	donor := &donordata.Provided{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := YourDetails(template.Execute, nil)(testAppData, w, r, donor)

	assert.ErrorIs(t, err, expectedError)
}

func TestGetYourDetailsWhenVoucherInvitedByDonor(t *testing.T) {
	donor := &donordata.Provided{LpaUID: "lpa-uid", VoucherInvitedAt: testNow}

	testcases := map[string]struct {
		setupVoucherStore func(*testing.T, *http.Request) *mockVoucherStore
		voucher           *voucherdata.Provided
		data              yourDetailsData
	}{
		"details confirmed": {
			setupVoucherStore: func(t *testing.T, r *http.Request) *mockVoucherStore {
				s := newMockVoucherStore(t)
				s.EXPECT().
					GetAny(r.Context()).
					Return(&voucherdata.Provided{DonorDetailsMatch: form.Yes}, nil)
				return s
			},
			data: yourDetailsData{
				App:                            testAppData,
				Donor:                          donor,
				DonorDetailsConfirmedByVoucher: true,
			},
		},
		"details not confirmed": {
			setupVoucherStore: func(t *testing.T, r *http.Request) *mockVoucherStore {
				s := newMockVoucherStore(t)
				s.EXPECT().
					GetAny(r.Context()).
					Return(&voucherdata.Provided{}, nil)
				return s
			},
			data: yourDetailsData{
				App:   testAppData,
				Donor: donor,
			},
		},
		"voucher not started": {
			setupVoucherStore: func(t *testing.T, r *http.Request) *mockVoucherStore {
				s := newMockVoucherStore(t)
				s.EXPECT().
					GetAny(r.Context()).
					Return(nil, dynamo.NotFoundError{})
				return s
			},
			data: yourDetailsData{
				App:   testAppData,
				Donor: donor,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.data).
				Return(nil)

			err := YourDetails(template.Execute, tc.setupVoucherStore(t, r))(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetYourDetailsWhenVoucherStoreError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	donor := &donordata.Provided{LpaUID: "lpa-uid", VoucherInvitedAt: testNow}

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := YourDetails(nil, voucherStore)(testAppData, w, r, donor)

	assert.ErrorIs(t, err, expectedError)
}
