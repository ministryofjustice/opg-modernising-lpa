package app

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDocumentStore(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	s3Client := newMockS3Client(t)
	eventClient := newMockEventClient(t)

	expected := &documentStore{dynamoClient: dynamoClient, s3Client: s3Client, eventClient: eventClient}

	assert.Equal(t, expected, NewDocumentStore(dynamoClient, s3Client, eventClient, nil, nil))
}

func TestDocumentStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllByPartialSK", ctx, dynamo.LpaKey("123"), dynamo.DocumentKey(""), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(page.Documents{{PK: dynamo.LpaKey("123")}})
			json.Unmarshal(b, v)
			return nil
		})

	documentStore := documentStore{dynamoClient: dynamoClient}

	documents, err := documentStore.GetAll(ctx)

	assert.Nil(t, err)
	assert.Equal(t, page.Documents{{PK: dynamo.LpaKey("123")}}, documents)
}

func TestDocumentStoreGetAllMissingSessionData(t *testing.T) {
	documentStore := documentStore{}
	_, err := documentStore.GetAll(context.Background())

	assert.NotNil(t, err)
}

func TestDocumentStoreGetAllMissingLpaIdInSession(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{})

	documentStore := documentStore{}
	_, err := documentStore.GetAll(ctx)

	assert.NotNil(t, err)
}

func TestDocumentStoreGetAllWhenDynamoClientAllByPartialSKError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllByPartialSK", ctx, dynamo.LpaKey("123"), dynamo.DocumentKey(""), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(page.Documents{{PK: dynamo.LpaKey("123")}})
			json.Unmarshal(b, v)
			return expectedError
		})

	documentStore := documentStore{dynamoClient: dynamoClient}
	_, err := documentStore.GetAll(ctx)

	assert.Equal(t, expectedError, err)
}

func TestDocumentStoreGetAllWhenNoResults(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllByPartialSK", ctx, dynamo.LpaKey("123"), dynamo.DocumentKey(""), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
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
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(
			ctx,
			dynamo.LpaKey("123"),
			dynamo.DocumentKey("object/key"),
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
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(
			ctx,
			dynamo.LpaKey("123"),
			dynamo.DocumentKey("object/key"),
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
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, page.Document{Key: "a-key"}).
		Return(nil)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.Put(ctx, page.Document{Key: "a-key"})

	assert.Nil(t, err)
}

func TestDocumentStorePutWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, page.Document{Key: "a-key"}).
		Return(expectedError)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.Put(ctx, page.Document{Key: "a-key"})

	assert.Equal(t, expectedError, err)
}

func TestDeleteInfectedDocuments(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, []dynamo.Keys{
			{PK: dynamo.LpaKey("a-pk"), SK: dynamo.DocumentKey("a-sk")},
			{PK: dynamo.LpaKey("another-pk"), SK: dynamo.DocumentKey("another-sk")},
		}).
		Return(nil)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.DeleteInfectedDocuments(ctx, page.Documents{
		{PK: dynamo.LpaKey("a-pk"), SK: dynamo.DocumentKey("a-sk"), Key: "a-key", VirusDetected: true},
		{PK: dynamo.LpaKey("another-pk"), SK: dynamo.DocumentKey("another-sk"), Key: "another-key", VirusDetected: true},
	})

	assert.Nil(t, err)
}

func TestDeleteInfectedDocumentsWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, []dynamo.Keys{
			{PK: dynamo.LpaKey("a-pk"), SK: dynamo.DocumentKey("a-sk")},
			{PK: dynamo.LpaKey("another-pk"), SK: dynamo.DocumentKey("another-sk")},
		}).
		Return(expectedError)

	documentStore := documentStore{dynamoClient: dynamoClient}

	err := documentStore.DeleteInfectedDocuments(ctx, page.Documents{
		{PK: dynamo.LpaKey("a-pk"), SK: dynamo.DocumentKey("a-sk"), Key: "a-key", VirusDetected: true},
		{PK: dynamo.LpaKey("another-pk"), SK: dynamo.DocumentKey("another-sk"), Key: "another-key", VirusDetected: true},
	})

	assert.Equal(t, expectedError, err)
}

func TestDeleteInfectedDocumentsNonInfectedDocumentsAreNotDeleted(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	documentStore := documentStore{}

	err := documentStore.DeleteInfectedDocuments(ctx, page.Documents{
		{PK: "a-pk", SK: "a-sk", Key: "a-key"},
		{PK: "another-pk", SK: "another-sk", Key: "another-key"},
	})

	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		DeleteObject(ctx, "a-key").
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, dynamo.LpaKey("a-pk"), dynamo.DocumentKey("a-sk")).
		Return(nil)

	documentStore := documentStore{s3Client: s3Client, dynamoClient: dynamoClient}

	err := documentStore.Delete(ctx, page.Document{PK: dynamo.LpaKey("a-pk"), SK: dynamo.DocumentKey("a-sk"), Key: "a-key", VirusDetected: true})

	assert.Nil(t, err)
}

func TestDeleteWhenS3ClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		DeleteObject(ctx, "a-key").
		Return(expectedError)

	documentStore := documentStore{s3Client: s3Client}

	err := documentStore.Delete(ctx, page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key", VirusDetected: true})

	assert.Equal(t, expectedError, err)
}

func TestDeleteWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		DeleteObject(ctx, "a-key").
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, dynamo.LpaKey("a-pk"), dynamo.DocumentKey("a-sk")).
		Return(expectedError)

	documentStore := documentStore{s3Client: s3Client, dynamoClient: dynamoClient}

	err := documentStore.Delete(ctx, page.Document{PK: dynamo.LpaKey("a-pk"), SK: dynamo.DocumentKey("a-sk"), Key: "a-key", VirusDetected: true})

	assert.Equal(t, expectedError, err)
}

func TestDocumentStoreSubmit(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}
	documents := page.Documents{
		{PK: "a-pk", SK: "a-sk", Key: "a-key", Filename: "a-filename.pdf"},
		{PK: "b-pk", SK: "b-sk", Key: "b-key", Filename: "b-filename.png"},
	}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObjectTagging(ctx, "a-key", map[string]string{"replicate": "true", "virus-scan-status": "ok"}).
		Return(nil)
	s3Client.EXPECT().
		PutObjectTagging(ctx, "b-key", map[string]string{"replicate": "true", "virus-scan-status": "ok"}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendReducedFeeRequested(ctx, event.ReducedFeeRequested{
			UID:         "lpa-uid",
			RequestType: pay.HalfFee.String(),
			Evidence: []event.Evidence{
				{Path: "a-key", Filename: "a-filename.pdf"},
				{Path: "b-key", Filename: "b-filename.png"},
			},
			EvidenceDelivery: pay.Upload.String(),
		}).
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		BatchPut(ctx, []any{
			page.Document{PK: "a-pk", SK: "a-sk", Key: "a-key", Sent: now, Filename: "a-filename.pdf"},
			page.Document{PK: "b-pk", SK: "b-sk", Key: "b-key", Sent: now, Filename: "b-filename.png"},
		}).
		Return(nil)

	documentStore := &documentStore{
		dynamoClient: dynamoClient,
		eventClient:  eventClient,
		s3Client:     s3Client,
		now:          func() time.Time { return now },
	}

	err := documentStore.Submit(ctx, donor, documents)
	assert.Nil(t, err)
}

func TestDocumentStoreSubmitWhenNoUnsentDocuments(t *testing.T) {
	ctx := context.Background()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}
	documents := page.Documents{{PK: "a-pk", SK: "a-sk", Key: "a-key", Sent: time.Now()}}

	documentStore := &documentStore{}

	err := documentStore.Submit(ctx, donor, documents)
	assert.Nil(t, err)
}

func TestDocumentStoreSubmitWhenS3ClientErrors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}
	documents := page.Documents{{PK: "a-pk", SK: "a-sk", Key: "a-key"}}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObjectTagging(ctx, "a-key", mock.Anything).
		Return(expectedError)

	documentStore := &documentStore{
		s3Client: s3Client,
		now:      func() time.Time { return now },
	}

	err := documentStore.Submit(ctx, donor, documents)
	assert.Equal(t, expectedError, err)
}

func TestDocumentStoreSubmitWhenEventClientErrors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}
	documents := page.Documents{{PK: "a-pk", SK: "a-sk", Key: "a-key"}}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObjectTagging(ctx, "a-key", mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendReducedFeeRequested(ctx, mock.Anything).
		Return(expectedError)

	documentStore := &documentStore{
		eventClient: eventClient,
		s3Client:    s3Client,
		now:         func() time.Time { return now },
	}

	err := documentStore.Submit(ctx, donor, documents)
	assert.Equal(t, expectedError, err)
}

func TestDocumentStoreSubmitWhenDynamoClientErrors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}
	documents := page.Documents{{PK: "a-pk", SK: "a-sk", Key: "a-key"}}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObjectTagging(ctx, "a-key", mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendReducedFeeRequested(ctx, mock.Anything).
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		BatchPut(ctx, mock.Anything).
		Return(expectedError)

	documentStore := &documentStore{
		dynamoClient: dynamoClient,
		eventClient:  eventClient,
		s3Client:     s3Client,
		now:          func() time.Time { return now },
	}

	err := documentStore.Submit(ctx, donor, documents)
	assert.Equal(t, expectedError, err)
}

func TestDocumentCreate(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload, LpaID: "lpa-id"}

	data := []byte("some-data")

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObject(ctx, "lpa-uid/evidence/a-uuid", data).
		Return(nil)

	expectedDocument := page.Document{
		PK:       dynamo.LpaKey("lpa-id"),
		SK:       dynamo.DocumentKey("lpa-uid/evidence/a-uuid"),
		Filename: "a-filename",
		Key:      "lpa-uid/evidence/a-uuid",
		Uploaded: now,
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, expectedDocument).
		Return(nil)

	documentStore := &documentStore{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
		now:          func() time.Time { return now },
		randomUUID:   func() string { return "a-uuid" },
	}

	document, err := documentStore.Create(ctx, donor, "a-filename", data)

	assert.Nil(t, err)
	assert.Equal(t, expectedDocument, document)
}

func TestDocumentCreateWhenS3Error(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload, LpaID: "lpa-id"}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObject(ctx, "lpa-uid/evidence/a-uuid", mock.Anything).
		Return(expectedError)

	documentStore := &documentStore{
		s3Client:   s3Client,
		now:        func() time.Time { return now },
		randomUUID: func() string { return "a-uuid" },
	}

	_, err := documentStore.Create(ctx, donor, "a-filename", []byte("some-data"))

	assert.Equal(t, expectedError, err)
}

func TestDocumentCreateWhenDynamoError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	donor := &donordata.Provided{LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload, LpaID: "lpa-id"}

	data := []byte("some-data")

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		PutObject(ctx, "lpa-uid/evidence/a-uuid", data).
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	documentStore := &documentStore{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
		now:          func() time.Time { return now },
		randomUUID:   func() string { return "a-uuid" },
	}

	_, err := documentStore.Create(ctx, donor, "a-filename", data)

	assert.Equal(t, expectedError, err)
}
