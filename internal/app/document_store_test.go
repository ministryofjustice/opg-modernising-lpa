package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDocumentStore(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	s3Client := newMockS3Client(t)

	expected := &documentStore{dynamoClient: dynamoClient, s3Client: s3Client, randomUUID: nil}

	assert.Equal(t, expected, NewDocumentStore(dynamoClient, s3Client, nil))
}

func TestDocumentStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllByPartialSk", ctx, "LPA#123", "#DOCUMENT#", mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(page.Documents{{PK: "LPA#123"}})
			json.Unmarshal(b, v)
			return nil
		})

	documentStore := documentStore{dynamoClient: dynamoClient}

	documents, err := documentStore.GetAll(ctx)

	assert.Nil(t, err)
	assert.Equal(t, page.Documents{{PK: "LPA#123"}}, documents)
}

func TestDocumentStoreGetAllMissingSessionData(t *testing.T) {
	documentStore := documentStore{}
	_, err := documentStore.GetAll(context.Background())

	assert.NotNil(t, err)
}

func TestDocumentStoreGetAllMissingLpaIdInSession(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	documentStore := documentStore{}
	_, err := documentStore.GetAll(ctx)

	assert.NotNil(t, err)
}

func TestDocumentStoreGetAllWhenDynamoClientAllByPartialSkError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllByPartialSk", ctx, "LPA#123", "#DOCUMENT#", mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(page.Documents{{PK: "LPA#123"}})
			json.Unmarshal(b, v)
			return expectedError
		})

	documentStore := documentStore{dynamoClient: dynamoClient}
	_, err := documentStore.GetAll(ctx)

	assert.Equal(t, expectedError, err)
}

func TestDocumentStoreGetAllWhenNoResults(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllByPartialSk", ctx, "LPA#123", "#DOCUMENT#", mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(page.Documents{})
			json.Unmarshal(b, v)
			return dynamo.NotFoundError{}
		})

	documentStore := documentStore{dynamoClient: dynamoClient}
	documents, err := documentStore.GetAll(ctx)

	assert.Nil(t, err)
	assert.Equal(t, page.Documents{}, documents)
}

func TestDocumentStoreUpdateScanResults(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Update",
			ctx,
			"LPA#123",
			"#DOCUMENT#object/key",
			map[string]types.AttributeValue{
				":virusDetected": &types.AttributeValueMemberBOOL{Value: true},
				":scanned":       &types.AttributeValueMemberBOOL{Value: true},
			}, "set VirusDetected = :virusDetected, Scanned = :scanned").
		Return(nil)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.UpdateScanResults(ctx, "123", "object/key", true)

	assert.Nil(t, err)
}

func TestDocumentStoreUpdateScanResultsWhenUpdateError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Update",
			ctx,
			"LPA#123",
			"#DOCUMENT#object/key",
			map[string]types.AttributeValue{
				":virusDetected": &types.AttributeValueMemberBOOL{Value: true},
				":scanned":       &types.AttributeValueMemberBOOL{Value: true},
			}, "set VirusDetected = :virusDetected, Scanned = :scanned").
		Return(expectedError)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.UpdateScanResults(ctx, "123", "object/key", true)

	assert.Equal(t, expectedError, err)
}

func TestDocumentStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, page.Document{Key: "a-key"}).
		Return(nil)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.Put(ctx, page.Document{Key: "a-key"})

	assert.Nil(t, err)
}

func TestDocumentStorePutWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, page.Document{Key: "a-key"}).
		Return(expectedError)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.Put(ctx, page.Document{Key: "a-key"})

	assert.Equal(t, expectedError, err)
}

func TestDeleteInfectedDocuments(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("DeleteKeys", ctx, []dynamo.Key{
			{PK: "a-pk", SK: "a-sk"},
			{PK: "another-pk", SK: "another-sk"},
		}).
		Return(nil)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.DeleteInfectedDocuments(ctx, page.Documents{
		{PK: "a-pk", SK: "a-sk", Key: "a-key", VirusDetected: true},
		{PK: "another-pk", SK: "another-sk", Key: "another-key", VirusDetected: true},
	})

	assert.Nil(t, err)
}

func TestDeleteInfectedDocumentsWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("DeleteKeys", ctx, []dynamo.Key{
			{PK: "a-pk", SK: "a-sk"},
			{PK: "another-pk", SK: "another-sk"},
		}).
		Return(expectedError)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.DeleteInfectedDocuments(ctx, page.Documents{
		{PK: "a-pk", SK: "a-sk", Key: "a-key", VirusDetected: true},
		{PK: "another-pk", SK: "another-sk", Key: "another-key", VirusDetected: true},
	})

	assert.Equal(t, expectedError, err)
}

func TestDeleteInfectedDocumentsNonInfectedDocumentsAreNotDeleted(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	documentStore := documentStore{}

	err := documentStore.DeleteInfectedDocuments(ctx, page.Documents{
		{PK: "a-pk", SK: "a-sk", Key: "a-key"},
		{PK: "another-pk", SK: "another-sk", Key: "another-key"},
	})

	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", ctx, "a-key").
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("DeleteOne", ctx, "a-pk", "a-sk").
		Return(nil)

	documentStore := documentStore{s3Client: s3Client, dynamoClient: dynamoClient}

	err := documentStore.Delete(ctx, page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key", VirusDetected: true})

	assert.Nil(t, err)
}

func TestDeleteWhenS3ClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", ctx, "a-key").
		Return(expectedError)

	documentStore := documentStore{s3Client: s3Client}

	err := documentStore.Delete(ctx, page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key", VirusDetected: true})

	assert.Equal(t, expectedError, err)
}

func TestDeleteWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", ctx, "a-key").
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("DeleteOne", ctx, "a-pk", "a-sk").
		Return(expectedError)

	documentStore := documentStore{s3Client: s3Client, dynamoClient: dynamoClient}

	err := documentStore.Delete(ctx, page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key", VirusDetected: true})

	assert.Equal(t, expectedError, err)
}

func TestDocumentKey(t *testing.T) {
	assert.Equal(t, "#DOCUMENT#key", documentKey("key"))
}

func TestBatchPut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("BatchPut", ctx, []interface{}{
			page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key"},
			page.Document{PK: "aanother-pk", SK: "aanother-sk", Key: "aanother-key"},
		}).
		Return(nil)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.BatchPut(ctx, []page.Document{
		{PK: "a-pk", SK: "a-sk", Key: "a-key"},
		{PK: "aanother-pk", SK: "aanother-sk", Key: "aanother-key"},
	})

	assert.Nil(t, err)
}

func TestBatchPutWhenDynamoError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("BatchPut", ctx, []interface{}{
			page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key"},
			page.Document{PK: "aanother-pk", SK: "aanother-sk", Key: "aanother-key"},
		}).
		Return(expectedError)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.BatchPut(ctx, []page.Document{
		{PK: "a-pk", SK: "a-sk", Key: "a-key"},
		{PK: "aanother-pk", SK: "aanother-sk", Key: "aanother-key"},
	})

	assert.Equal(t, expectedError, err)
}
