package main

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	factory := &Factory{now: testNowFn}

	assert.Equal(t, testNow, factory.Now()())
}

func TestDynamoClient(t *testing.T) {
	dynamoClient := newMockDynamodbClient(t)
	factory := &Factory{dynamoClient: dynamoClient}

	assert.Equal(t, dynamoClient, factory.DynamoClient())
}

func TestUuidString(t *testing.T) {
	factory := &Factory{uuidString: testUuidStringFn}

	assert.Equal(t, testUuidString, factory.UuidString()())
}

func TestAppData(t *testing.T) {
	factory := &Factory{}

	appData, err := factory.AppData()
	assert.Error(t, err)
	assert.Equal(t, appcontext.Data{}, appData)
}

func TestAppDataWhenSet(t *testing.T) {
	expected := appcontext.Data{Page: "hi"}
	factory := &Factory{appData: &expected}

	appData, err := factory.AppData()
	assert.Nil(t, err)
	assert.Equal(t, expected, appData)
}

func TestLambdaClient(t *testing.T) {
	factory := &Factory{}

	client := factory.LambdaClient()
	assert.NotNil(t, client)
}

func TestLambdaClientWhenSet(t *testing.T) {
	expected := newMockLambdaClient(t)

	factory := &Factory{lambdaClient: expected}

	client := factory.LambdaClient()
	assert.Equal(t, expected, client)
}

func TestSecretsClient(t *testing.T) {
	factory := &Factory{}

	client, err := factory.SecretsClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestSecretsClientWhenSet(t *testing.T) {
	expected := newMockSecretsClient(t)

	factory := &Factory{secretsClient: expected}

	client, err := factory.SecretsClient()
	assert.Nil(t, err)
	assert.Equal(t, expected, client)
}

func TestShareCodeSender(t *testing.T) {
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

func TestShareCodeSenderWhenSet(t *testing.T) {
	ctx := context.Background()

	expected := newMockShareCodeSender(t)

	factory := &Factory{shareCodeSender: expected}

	sender, err := factory.ShareCodeSender(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expected, sender)
}

func TestShareCodeSenderWhenSecretsClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, secrets.GovUkNotify).
		Return("", expectedError)

	factory := &Factory{secretsClient: secretsClient, bundle: &localize.Bundle{}}

	_, err := factory.ShareCodeSender(ctx)
	assert.ErrorIs(t, err, expectedError)
}

func TestShareCodeSenderWhenNotifyClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, secrets.GovUkNotify).
		Return("", nil)

	factory := &Factory{secretsClient: secretsClient, bundle: &localize.Bundle{}}

	_, err := factory.ShareCodeSender(ctx)
	assert.NotNil(t, err)
}

func TestLpaStoreClient(t *testing.T) {
	secretsClient := newMockSecretsClient(t)

	factory := &Factory{secretsClient: secretsClient}

	client, err := factory.LpaStoreClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestLpaStoreClientWhenSet(t *testing.T) {
	expected := newMockLpaStoreClient(t)

	factory := &Factory{lpaStoreClient: expected}

	client, err := factory.LpaStoreClient()
	assert.Nil(t, err)
	assert.Equal(t, expected, client)
}

func TestUidStore(t *testing.T) {
	factory := &Factory{}

	store, err := factory.UidStore()
	assert.Nil(t, err)
	assert.NotNil(t, store)
}

func TestUidStoreWhenSet(t *testing.T) {
	expected := newMockUidStore(t)

	factory := &Factory{uidStore: expected}

	store, err := factory.UidStore()
	assert.Nil(t, err)
	assert.Equal(t, expected, store)
}

func TestUidClient(t *testing.T) {
	factory := &Factory{}

	client := factory.UidClient()
	assert.NotNil(t, client)
}

func TestUidClientWhenSet(t *testing.T) {
	expected := newMockUidClient(t)
	factory := &Factory{uidClient: expected}

	client := factory.UidClient()
	assert.Equal(t, expected, client)
}

func TestEventClient(t *testing.T) {
	factory := &Factory{}

	client := factory.EventClient()
	assert.NotNil(t, client)
}

func TestEventClientWhenSet(t *testing.T) {
	expected := newMockEventClient(t)
	factory := &Factory{eventClient: expected}

	client := factory.EventClient()
	assert.Equal(t, expected, client)
}
