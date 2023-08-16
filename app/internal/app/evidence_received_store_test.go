package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestEvidenceReceivedStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectGet(ctx, "LPA#an-id", "#EVIDENCE_RECEIVED", nil, nil)

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: dynamoClient}

	ok, err := evidenceReceivedStore.Get(ctx)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestEvidenceReceivedStoreGetWhenFalse(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectGet(ctx, "LPA#an-id", "#EVIDENCE_RECEIVED", nil, dynamo.NotFoundError{})

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: dynamoClient}

	ok, err := evidenceReceivedStore.Get(ctx)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestEvidenceReceivedStoreGetWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: nil}

	_, err := evidenceReceivedStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestEvidenceReceivedStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectGet(ctx, "LPA#an-id", "#EVIDENCE_RECEIVED", &page.Lpa{ID: "an-id"}, expectedError)

	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: dynamoClient}

	_, err := evidenceReceivedStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}
