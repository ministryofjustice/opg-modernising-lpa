package voucherpage

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = context.WithValue(context.Background(), (*string)(nil), "test")

func TestVouchFailer(t *testing.T) {
	lpa := &lpadata.Lpa{
		LpaUID: "lpa-uid",
		Donor:  lpadata.Donor{Email: "john@example.com", ContactLanguagePreference: localize.Cy},
	}
	provided := &voucherdata.Provided{
		SK:         dynamo.VoucherKey("a-voucher"),
		FirstNames: "Vivian",
		LastName:   "Vaughn",
	}
	donor := &donordata.Provided{
		FailedVouchAttempts: 1,
		WantVoucher:         form.Yes,
		Voucher:             donordata.Voucher{FirstNames: "A"},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(donor, nil)
	donorStore.EXPECT().
		FailVoucher(ctx, donor, provided.SK).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(lpa).
		Return("greeting")
	notifyClient.EXPECT().
		SendActorEmail(ctx, localize.Cy, "john@example.com", "lpa-uid", notify.VouchingFailedAttemptEmail{
			Greeting:          "greeting",
			VoucherFullName:   "Vivian Vaughn",
			DonorStartPageURL: "app:///start",
		}).
		Return(nil)

	err := makeVouchFailer(donorStore, notifyClient, "app://")(ctx, provided, lpa)
	assert.Nil(t, err)
}

func TestVouchFailerWhenDonorStoreGetErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := makeVouchFailer(donorStore, nil, "app://")(ctx, &voucherdata.Provided{}, &lpadata.Lpa{})
	assert.ErrorIs(t, err, expectedError)
}

func TestVouchFailerWheNotifyClientErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(&donordata.Provided{}, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("greeting")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := makeVouchFailer(donorStore, notifyClient, "app://")(ctx, &voucherdata.Provided{}, &lpadata.Lpa{})
	assert.ErrorIs(t, err, expectedError)
}

func TestVouchFailerWhenDonorStorePutErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		FailVoucher(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("greeting")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := makeVouchFailer(donorStore, notifyClient, "app://")(ctx, &voucherdata.Provided{}, &lpadata.Lpa{})
	assert.ErrorIs(t, err, expectedError)
}
