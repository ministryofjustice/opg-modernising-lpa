package main

import (
	"context"
	"os"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/stretchr/testify/assert"
)

func TestFactoryNow(t *testing.T) {
	factory := &Factory{now: testNowFn}

	assert.Equal(t, testNow, factory.Now()())
}

func TestFactoryDynamoClient(t *testing.T) {
	dynamoClient := newMockDynamodbClient(t)
	factory := &Factory{dynamoClient: dynamoClient}

	assert.Equal(t, dynamoClient, factory.DynamoClient())
}

func TestFactoryUuidString(t *testing.T) {
	factory := &Factory{uuidString: testUuidStringFn}

	assert.Equal(t, testUuidString, factory.UuidString()())
}

func TestFactoryAppData(t *testing.T) {
	factory := &Factory{}

	appData, err := factory.AppData()
	assert.Error(t, err)
	assert.Equal(t, appcontext.Data{}, appData)
}

func TestFactoryAppDataWhenSet(t *testing.T) {
	expected := appcontext.Data{Page: "hi"}
	factory := &Factory{appData: &expected}

	appData, err := factory.AppData()
	assert.Nil(t, err)
	assert.Equal(t, expected, appData)
}

func TestFactoryLambdaClient(t *testing.T) {
	factory := &Factory{}

	client := factory.LambdaClient()
	assert.NotNil(t, client)
}

func TestFactoryLambdaClientWhenSet(t *testing.T) {
	expected := newMockLambdaClient(t)

	factory := &Factory{lambdaClient: expected}

	client := factory.LambdaClient()
	assert.Equal(t, expected, client)
}

func TestFactorySecretsClient(t *testing.T) {
	factory := &Factory{}

	client, err := factory.SecretsClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestFactorySecretsClientWhenSet(t *testing.T) {
	expected := newMockSecretsClient(t)

	factory := &Factory{secretsClient: expected}

	client, err := factory.SecretsClient()
	assert.Nil(t, err)
	assert.Equal(t, expected, client)
}

func TestFactoryShareCodeSender(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, secrets.GovUkNotify).
		Return("a-b-c-d-e-f-g-h-i-j-k", nil)

	factory := &Factory{secretsClient: secretsClient, bundle: &localize.Bundle{}}

	sender, err := factory.ShareCodeSender(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, sender)
}

func TestFactoryShareCodeSenderWhenSet(t *testing.T) {
	ctx := context.Background()

	expected := newMockShareCodeSender(t)

	factory := &Factory{shareCodeSender: expected}

	sender, err := factory.ShareCodeSender(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expected, sender)
}

func TestFactoryShareCodeSenderWhenBundleError(t *testing.T) {
	ctx := context.Background()

	factory := &Factory{}

	_, err := factory.ShareCodeSender(ctx)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestFactoryShareCodeSenderWhenSecretsClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, secrets.GovUkNotify).
		Return("", expectedError)

	factory := &Factory{secretsClient: secretsClient, bundle: &localize.Bundle{}}

	_, err := factory.ShareCodeSender(ctx)
	assert.ErrorIs(t, err, expectedError)
}

func TestFactoryShareCodeSenderWhenNotifyClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, secrets.GovUkNotify).
		Return("", nil)

	factory := &Factory{secretsClient: secretsClient, bundle: &localize.Bundle{}}

	_, err := factory.ShareCodeSender(ctx)
	assert.NotNil(t, err)
}

func TestFactoryLpaStoreClient(t *testing.T) {
	secretsClient := newMockSecretsClient(t)

	factory := &Factory{secretsClient: secretsClient}

	client, err := factory.LpaStoreClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestFactoryLpaStoreClientWhenSet(t *testing.T) {
	expected := newMockLpaStoreClient(t)

	factory := &Factory{lpaStoreClient: expected}

	client, err := factory.LpaStoreClient()
	assert.Nil(t, err)
	assert.Equal(t, expected, client)
}

func TestFactoryUidStore(t *testing.T) {
	factory := &Factory{}

	store, err := factory.UidStore()
	assert.Nil(t, err)
	assert.NotNil(t, store)
}

func TestFactoryUidStoreWhenSet(t *testing.T) {
	expected := newMockUidStore(t)

	factory := &Factory{uidStore: expected}

	store, err := factory.UidStore()
	assert.Nil(t, err)
	assert.Equal(t, expected, store)
}

func TestFactoryUidClient(t *testing.T) {
	factory := &Factory{}

	client := factory.UidClient()
	assert.NotNil(t, client)
}

func TestFactoryUidClientWhenSet(t *testing.T) {
	expected := newMockUidClient(t)
	factory := &Factory{uidClient: expected}

	client := factory.UidClient()
	assert.Equal(t, expected, client)
}

func TestFactoryEventClient(t *testing.T) {
	factory := &Factory{}

	client := factory.EventClient()
	assert.NotNil(t, client)
}

func TestFactoryEventClientWhenSet(t *testing.T) {
	expected := newMockEventClient(t)
	factory := &Factory{eventClient: expected}

	client := factory.EventClient()
	assert.Equal(t, expected, client)
}

func TestFactoryScheduledStore(t *testing.T) {
	factory := &Factory{}

	client := factory.ScheduledStore()
	assert.NotNil(t, client)
}

func TestFactoryScheduledStoreWhenSet(t *testing.T) {
	expected := newMockScheduledStore(t)
	factory := &Factory{scheduledStore: expected}

	client := factory.ScheduledStore()
	assert.Equal(t, expected, client)
}

func TestFactoryBundle(t *testing.T) {
	factory := &Factory{}

	bundle, err := factory.Bundle()
	assert.Nil(t, bundle)
	assert.Error(t, err)
}

func TestFactoryBundleWhenSet(t *testing.T) {
	expected := newMockBundle(t)
	factory := &Factory{bundle: expected}

	bundle, err := factory.Bundle()
	assert.Equal(t, expected, bundle)
	assert.Nil(t, err)
}
