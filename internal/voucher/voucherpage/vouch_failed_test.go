package voucherpage

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = context.WithValue(context.Background(), (*string)(nil), "test")

func TestVouchFailed(t *testing.T) {
	lpa := &lpadata.Lpa{
		LpaUID: "lpa-uid",
		Donor:  lpadata.Donor{Email: "john@example.com"},
	}
	provided := &voucherdata.Provided{
		FirstNames: "Vivian",
		LastName:   "Vaughn",
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(&donordata.Provided{
			FailedVouchAttempts: 1,
			WantVoucher:         form.Yes,
			Voucher:             donordata.Voucher{FirstNames: "A"},
		}, nil)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			FailedVouchAttempts: 2,
			WantVoucher:         form.No,
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(lpa).
		Return("greeting")
	notifyClient.EXPECT().
		SendActorEmail(ctx, "john@example.com", "lpa-uid", notify.VouchingFailedAttemptEmail{
			Greeting:          "greeting",
			VoucherFullName:   "Vivian Vaughn",
			DonorStartPageURL: "app:///start",
		}).
		Return(nil)

	err := vouchFailed(donorStore, notifyClient, "app://")(ctx, provided, lpa)
	assert.Nil(t, err)
}

func TestVouchFailedWhenDonorStoreGetErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := vouchFailed(donorStore, nil, "app://")(ctx, &voucherdata.Provided{}, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestVouchFailedWheNotifyClientErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(&donordata.Provided{}, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("greeting")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := vouchFailed(donorStore, notifyClient, "app://")(ctx, &voucherdata.Provided{}, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestVouchFailedWhenDonorStorePutErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("greeting")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := vouchFailed(donorStore, notifyClient, "app://")(ctx, &voucherdata.Provided{}, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}
