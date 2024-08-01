package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestEvidenceReceivedStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.EvidenceReceivedKey(), nil, nil)

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: dynamoClient}

	ok, err := evidenceReceivedStore.Get(ctx)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestEvidenceReceivedStoreGetWhenFalse(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.EvidenceReceivedKey(), nil, dynamo.NotFoundError{})

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: dynamoClient}

	ok, err := evidenceReceivedStore.Get(ctx)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestEvidenceReceivedStoreGetWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: nil}

	_, err := evidenceReceivedStore.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestEvidenceReceivedStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.EvidenceReceivedKey(), &actor.DonorProvidedDetails{LpaID: "an-id"}, expectedError)

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: dynamoClient}

	_, err := evidenceReceivedStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}
