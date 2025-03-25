package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLpaStoreEventHandlerHandleUnknownEvent(t *testing.T) {
	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, nil, &events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown lpastore event"), err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedWhenChangeTypeNotExpected(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"WHAT"}`),
	}

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, nil, event)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreate(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CREATE"}`),
	}

	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			ContactLanguagePreference: localize.Cy,
		},
	}

	donor := &donordata.Provided{
		PK:            dynamo.LpaKey("an-lpa"),
		Correspondent: donordata.Correspondent{Email: "hey@example.com"},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{
			PK: dynamo.LpaKey("an-lpa"),
			SK: dynamo.DonorKey("a-donor"),
		}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("an-lpa"), dynamo.DonorKey("a-donor"), mock.Anything).
		Return(nil).
		SetData(donor)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(lpa).
		Return("hello")
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToDonor(donor), "M-1111-2222-3333", notify.DigitalDonorLpaSubmittedEmail{
			Greeting: "hello",
			LpaType:  "personal welfare",
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(localizer)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().NotifyClient(ctx).Return(notifyClient, nil)
	factory.EXPECT().Bundle().Return(bundle, nil)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenFactoryErrors(t *testing.T) {
	testcases := map[string]func(*mockFactory){
		"LpaStoreClient": func(factory *mockFactory) {
			factory.EXPECT().LpaStoreClient().Return(nil, expectedError)
		},
		"Bundle": func(factory *mockFactory) {
			factory.EXPECT().LpaStoreClient().Return(nil, nil)
			factory.EXPECT().Bundle().Return(nil, expectedError)
		},
		"NotifyClient": func(factory *mockFactory) {
			factory.EXPECT().LpaStoreClient().Return(nil, nil)
			factory.EXPECT().Bundle().Return(nil, nil)
			factory.EXPECT().NotifyClient(ctx).Return(nil, expectedError)
		},
	}

	for name, setupFactory := range testcases {
		t.Run(name, func(t *testing.T) {
			v := &events.CloudWatchEvent{
				DetailType: "lpa-updated",
				Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CREATE"}`),
			}

			factory := newMockFactory(t)
			setupFactory(factory)

			handler := &lpastoreEventHandler{}

			err := handler.Handle(ctx, factory, v)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenPaperDonor(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CREATE"}`),
	}

	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			Channel:                   lpadata.ChannelPaper,
			ContactLanguagePreference: localize.Cy,
			Mobile:                    "07",
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(ctx, notify.ToLpaDonor(lpa), "M-1111-2222-3333", notify.PaperDonorLpaSubmittedSMS{
			LpaType: "personal welfare",
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(localizer)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(nil)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().NotifyClient(ctx).Return(notifyClient, nil)
	factory.EXPECT().Bundle().Return(bundle, nil)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenPaperDonorWithNoMobile(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CREATE"}`),
	}

	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			Channel:                   lpadata.ChannelPaper,
			ContactLanguagePreference: localize.Cy,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(newMockLocalizer(t))

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(nil)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().NotifyClient(ctx).Return(nil, nil)
	factory.EXPECT().Bundle().Return(bundle, nil)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenPaperDonorAndNotifyErrors(t *testing.T) {
	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			Channel:                   lpadata.ChannelPaper,
			ContactLanguagePreference: localize.Cy,
			Mobile:                    "07",
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(lpa, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(localizer)

	err := handleCreate(ctx, nil, lpaStoreClient, notifyClient, bundle, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenLpaStoreErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := handleCreate(ctx, nil, lpaStoreClient, nil, nil, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenDonorStoreErrors(t *testing.T) {
	testcases := map[string]func(*mockDynamodbClient){
		"first": func(client *mockDynamodbClient) {
			client.EXPECT().
				OneByUID(mock.Anything, mock.Anything).
				Return(dynamo.Keys{}, expectedError)
		},
		"second": func(client *mockDynamodbClient) {
			client.EXPECT().
				OneByUID(mock.Anything, mock.Anything).
				Return(dynamo.Keys{
					PK: dynamo.LpaKey("an-lpa"),
					SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
				}, nil)
			client.EXPECT().
				One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)
		},
	}

	for name, setupDynamodbClient := range testcases {
		t.Run(name, func(t *testing.T) {

			lpa := &lpadata.Lpa{
				Type: lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					ContactLanguagePreference: localize.Cy,
				},
			}

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				Lpa(mock.Anything, mock.Anything).
				Return(lpa, nil)

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(mock.Anything).
				Return(newMockLocalizer(t))

			client := newMockDynamodbClient(t)
			setupDynamodbClient(client)

			err := handleCreate(ctx, client, lpaStoreClient, nil, bundle, lpaUpdatedEvent{})
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCreateWhenNotifyErrors(t *testing.T) {
	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			ContactLanguagePreference: localize.Cy,
		},
	}

	donor := &donordata.Provided{
		PK:            dynamo.LpaKey("an-lpa"),
		Correspondent: donordata.Correspondent{Email: "hey@example.com"},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(lpa, nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{
			PK: dynamo.LpaKey("an-lpa"),
			SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		}, nil)
	client.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(donor)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("hello")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(localizer)

	err := handleCreate(ctx, client, lpaStoreClient, notifyClient, bundle, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSign(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CERTIFICATE_PROVIDER_SIGN"}`),
	}

	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			ContactLanguagePreference: localize.Cy,
		},
		CertificateProvider: lpadata.CertificateProvider{
			FirstNames: "a",
			LastName:   "b",
		},
	}

	donor := &donordata.Provided{
		PK:            dynamo.LpaKey("an-lpa"),
		Correspondent: donordata.Correspondent{Email: "hey@example.com"},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{
			PK: dynamo.LpaKey("an-lpa"),
			SK: dynamo.DonorKey("a-donor"),
		}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("an-lpa"), dynamo.DonorKey("a-donor"), mock.Anything).
		Return(nil).
		SetData(donor)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(lpa).
		Return("hello")
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToDonor(donor), "M-1111-2222-3333", notify.DigitalDonorCertificateProvidedEmail{
			Greeting:                    "hello",
			CertificateProviderFullName: "a b",
			LpaType:                     "personal welfare",
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(localizer)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().NotifyClient(ctx).Return(notifyClient, nil)
	factory.EXPECT().Bundle().Return(bundle, nil)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenFactoryErrors(t *testing.T) {
	testcases := map[string]func(*mockFactory){
		"LpaStoreClient": func(factory *mockFactory) {
			factory.EXPECT().LpaStoreClient().Return(nil, expectedError)
		},
		"Bundle": func(factory *mockFactory) {
			factory.EXPECT().LpaStoreClient().Return(nil, nil)
			factory.EXPECT().Bundle().Return(nil, expectedError)
		},
		"NotifyClient": func(factory *mockFactory) {
			factory.EXPECT().LpaStoreClient().Return(nil, nil)
			factory.EXPECT().Bundle().Return(nil, nil)
			factory.EXPECT().NotifyClient(ctx).Return(nil, expectedError)
		},
	}

	for name, setupFactory := range testcases {
		t.Run(name, func(t *testing.T) {
			v := &events.CloudWatchEvent{
				DetailType: "lpa-updated",
				Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CERTIFICATE_PROVIDER_SIGN"}`),
			}

			factory := newMockFactory(t)
			setupFactory(factory)

			handler := &lpastoreEventHandler{}

			err := handler.Handle(ctx, factory, v)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenPaperDonor(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CERTIFICATE_PROVIDER_SIGN"}`),
	}

	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			Channel:                   lpadata.ChannelPaper,
			ContactLanguagePreference: localize.Cy,
			Mobile:                    "07",
		},
		CertificateProvider: lpadata.CertificateProvider{
			FirstNames: "a",
			LastName:   "b",
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(ctx, notify.ToLpaDonor(lpa), "M-1111-2222-3333", notify.PaperDonorCertificateProvidedSMS{
			LpaType:                     "personal welfare",
			CertificateProviderFullName: "a b",
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(localizer)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(nil)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().NotifyClient(ctx).Return(notifyClient, nil)
	factory.EXPECT().Bundle().Return(bundle, nil)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenPaperDonorWithNoMobile(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CERTIFICATE_PROVIDER_SIGN"}`),
	}

	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			Channel:                   lpadata.ChannelPaper,
			ContactLanguagePreference: localize.Cy,
		},
		CertificateProvider: lpadata.CertificateProvider{
			FirstNames: "a",
			LastName:   "b",
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(newMockLocalizer(t))

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(nil)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().NotifyClient(ctx).Return(nil, nil)
	factory.EXPECT().Bundle().Return(bundle, nil)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenPaperDonorAndNotifyErrors(t *testing.T) {
	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			Channel:                   lpadata.ChannelPaper,
			ContactLanguagePreference: localize.Cy,
			Mobile:                    "07",
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(lpa, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(localizer)

	err := handleCertificateProviderSign(ctx, nil, lpaStoreClient, notifyClient, bundle, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenLpaStoreErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := handleCertificateProviderSign(ctx, nil, lpaStoreClient, nil, nil, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenDonorStoreErrors(t *testing.T) {
	testcases := map[string]func(*mockDynamodbClient){
		"first": func(client *mockDynamodbClient) {
			client.EXPECT().
				OneByUID(mock.Anything, mock.Anything).
				Return(dynamo.Keys{}, expectedError)
		},
		"second": func(client *mockDynamodbClient) {
			client.EXPECT().
				OneByUID(mock.Anything, mock.Anything).
				Return(dynamo.Keys{
					PK: dynamo.LpaKey("an-lpa"),
					SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
				}, nil)
			client.EXPECT().
				One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)
		},
	}

	for name, setupDynamodbClient := range testcases {
		t.Run(name, func(t *testing.T) {

			lpa := &lpadata.Lpa{
				Type: lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					ContactLanguagePreference: localize.Cy,
				},
			}

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				Lpa(mock.Anything, mock.Anything).
				Return(lpa, nil)

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(mock.Anything).
				Return(newMockLocalizer(t))

			client := newMockDynamodbClient(t)
			setupDynamodbClient(client)

			err := handleCertificateProviderSign(ctx, client, lpaStoreClient, nil, bundle, lpaUpdatedEvent{})
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCertificateProviderSignWhenNotifyErrors(t *testing.T) {
	lpa := &lpadata.Lpa{
		Type: lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			ContactLanguagePreference: localize.Cy,
		},
	}

	donor := &donordata.Provided{
		PK:            dynamo.LpaKey("an-lpa"),
		Correspondent: donordata.Correspondent{Email: "hey@example.com"},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(lpa, nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{
			PK: dynamo.LpaKey("an-lpa"),
			SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		}, nil)
	client.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(donor)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("hello")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(localizer)

	err := handleCertificateProviderSign(ctx, client, lpaStoreClient, notifyClient, bundle, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedRegister(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"REGISTER"}`),
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{Type: lpadata.LpaTypePersonalWelfare}, nil)

	donorUID := actoruid.New()
	attorneyUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	keys := []dynamo.Keys{
		{SK: dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("donor-sub")))},
		{SK: dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("attorney-sub")))},
		{SK: dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("certificate-provided-sub")))},
		{SK: dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("replacement-trust-sub")))},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		AllByLpaUIDAndPartialSK(ctx, "M-1111-2222-3333", dynamo.SubKey("")).
		Return(keys, nil)
	client.EXPECT().
		AllByKeys(ctx, keys).
		Return(marshalListOfMaps([]dashboarddata.LpaLink{{
			SK:        dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("donor-sub"))),
			UID:       donorUID,
			ActorType: actor.TypeDonor,
		}, {
			SK:        dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("attorney-sub"))),
			UID:       attorneyUID,
			ActorType: actor.TypeAttorney,
		}, {
			SK:        dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("certificate-provided-sub"))),
			UID:       actoruid.New(),
			ActorType: actor.TypeCertificateProvider,
		}, {
			SK:        dynamo.SubKey(base64.StdEncoding.EncodeToString([]byte("replacement-trust-sub"))),
			UID:       replacementTrustCorporationUID,
			ActorType: actor.TypeReplacementTrustCorporation,
		}}), nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLpaAccessGranted(ctx, event.LpaAccessGranted{
			UID:     "M-1111-2222-3333",
			LpaType: "personal-welfare",
			Actors: []event.LpaAccessGrantedActor{{
				SubjectID: "donor-sub",
				ActorUID:  donorUID.String(),
			}, {
				SubjectID: "attorney-sub",
				ActorUID:  attorneyUID.String(),
			}, {
				SubjectID: "replacement-trust-sub",
				ActorUID:  replacementTrustCorporationUID.String(),
			}},
		}).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().EventClient().Return(eventClient)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedRegisterWhenErrorCreatingLpaStore(t *testing.T) {
	v := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"REGISTER"}`),
	}

	factory := newMockFactory(t)
	factory.EXPECT().LpaStoreClient().Return(nil, expectedError)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, v)
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedRegisterWhenLpaStoreErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := handleRegister(ctx, nil, lpaStoreClient, nil, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedRegisterWhenDynamoErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		AllByLpaUIDAndPartialSK(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := handleRegister(ctx, client, lpaStoreClient, nil, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedRegisterWhenEventClientErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		AllByLpaUIDAndPartialSK(mock.Anything, mock.Anything, mock.Anything).
		Return([]dynamo.Keys{}, nil)
	client.EXPECT().
		AllByKeys(mock.Anything, mock.Anything).
		Return(marshalListOfMaps([]any{}), nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLpaAccessGranted(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleRegister(ctx, client, lpaStoreClient, eventClient, lpaUpdatedEvent{})
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedStatutoryWaitingPeriod(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"STATUTORY_WAITING_PERIOD"}`),
	}

	updated := &donordata.Provided{
		PK:                       dynamo.LpaKey("123"),
		SK:                       dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		StatutoryWaitingPeriodAt: testNow,
		UpdatedAt:                testNow,
	}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().Now().Return(testNowFn)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.Nil(t, err)
}

func TestHandleStatutoryWaitingPeriodWhenDynamoErrors(t *testing.T) {
	updated := &donordata.Provided{
		PK:                       dynamo.LpaKey("123"),
		SK:                       dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		StatutoryWaitingPeriodAt: testNow,
		UpdatedAt:                testNow,
	}
	updated.UpdateHash()

	testcases := map[string]struct {
		dynamoClient  func() *mockDynamodbClient
		expectedError error
	}{
		"OneByUID": {
			dynamoClient: func() *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(ctx, mock.Anything).
					Return(dynamo.Keys{}, expectedError)

				return client
			},
			expectedError: fmt.Errorf("failed to resolve uid: %w", expectedError),
		},
		"One": {
			dynamoClient: func() *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")}, nil)
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
			expectedError: fmt.Errorf("failed to get LPA: %w", expectedError),
		},
		"Put": {
			dynamoClient: func() *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")}, nil)
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					SetData(updated)
				client.EXPECT().
					Put(mock.Anything, updated).
					Return(expectedError)

				return client
			},
			expectedError: fmt.Errorf("failed to update donor details: %w", expectedError),
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			event := lpaUpdatedEvent{
				UID:        "M-1111-2222-3333",
				ChangeType: "STATUTORY_WAITING_PERIOD",
			}

			err := handleStatutoryWaitingPeriod(ctx, tc.dynamoClient(), testNowFn, event)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCannotRegister(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CANNOT_REGISTER"}`),
	}

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllByUID(ctx, "M-1111-2222-3333").
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().ScheduledStore().Return(scheduledStore)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.Nil(t, err)
}

func TestHandleCannotRegisterWhenStoreErrors(t *testing.T) {
	event := lpaUpdatedEvent{
		UID:        "M-1111-2222-3333",
		ChangeType: "CANNOT_REGISTER",
	}

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllByUID(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleCannotRegister(ctx, scheduledStore, event)
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedOpgStatusChange(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"OPG_STATUS_CHANGE"}`),
	}

	updated := &donordata.Provided{
		PK:              dynamo.LpaKey("123"),
		SK:              dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		DoNotRegisterAt: testNow,
		UpdatedAt:       testNow,
	}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{Status: lpadata.StatusDoNotRegister}, nil)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().LpaStoreClient().Return(lpaStoreClient, nil)
	factory.EXPECT().Now().Return(testNowFn)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedOpgStatusChangeWhenFactoryErrors(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"OPG_STATUS_CHANGE"}`),
	}

	factory := newMockFactory(t)
	factory.EXPECT().LpaStoreClient().Return(nil, expectedError)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.ErrorIs(t, err, expectedError)
}

func TestOpgStatusChangeWhenNotDoNotRegister(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")}, nil)
	dynamodbClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	event := lpaUpdatedEvent{
		UID:        "M-1111-2222-3333",
		ChangeType: "OPG_STATUS_CHANGE",
	}

	err := handleOpgStatusChange(ctx, dynamodbClient, lpaStoreClient, testNowFn, event)
	assert.Nil(t, err)
}

func TestOpgStatusChangeWhenErrors(t *testing.T) {
	testcases := map[string]struct {
		dynamoClient   func(*testing.T) *mockDynamodbClient
		lpaStoreClient func(*testing.T) *mockLpaStoreClient
		expectedError  error
	}{
		"OneByUID": {
			dynamoClient: func(t *testing.T) *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(ctx, mock.Anything).
					Return(dynamo.Keys{}, expectedError)

				return client
			},
			lpaStoreClient: func(*testing.T) *mockLpaStoreClient { return nil },
			expectedError:  fmt.Errorf("failed to resolve uid: %w", expectedError),
		},
		"One": {
			dynamoClient: func(t *testing.T) *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")}, nil)
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
			lpaStoreClient: func(*testing.T) *mockLpaStoreClient { return nil },
			expectedError:  fmt.Errorf("failed to get LPA: %w", expectedError),
		},
		"LpaStore": {
			dynamoClient: func(t *testing.T) *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")}, nil)
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
			lpaStoreClient: func(*testing.T) *mockLpaStoreClient {
				client := newMockLpaStoreClient(t)
				client.EXPECT().
					Lpa(mock.Anything, mock.Anything).
					Return(nil, expectedError)

				return client
			},
			expectedError: fmt.Errorf("error getting lpa: %w", expectedError),
		},
		"Put": {
			dynamoClient: func(t *testing.T) *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")}, nil)
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				client.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
			lpaStoreClient: func(*testing.T) *mockLpaStoreClient {
				client := newMockLpaStoreClient(t)
				client.EXPECT().
					Lpa(mock.Anything, mock.Anything).
					Return(&lpadata.Lpa{Status: lpadata.StatusDoNotRegister}, nil)

				return client
			},
			expectedError: fmt.Errorf("failed to update donor details: %w", expectedError),
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			event := lpaUpdatedEvent{
				UID:        "M-1111-2222-3333",
				ChangeType: "OPG_STATUS_CHANGE",
			}

			err := handleOpgStatusChange(ctx, tc.dynamoClient(t), tc.lpaStoreClient(t), testNowFn, event)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}
